package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/application/build"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/application/network"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/infrastructure/config"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/infrastructure/osm"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/infrastructure/postgres"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/interfaces/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()

	if err != nil {
		return err
	}

	catalog, err := osmdata.LoadCatalog(cfg.SourcesFile)
	if err != nil {
		return err
	}
	for _, region := range cfg.Regions {
		if _, ok := catalog.Regions[region.ID]; !ok {
			return fmt.Errorf("no OSM source configured for region %s", region.ID)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	startup, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(startup, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := pool.Ping(startup); err != nil {
		return err
	}
	repository := &postgres.Postgres{Pool: pool}
	release, err := repository.Initialize(startup, cfg.Regions)
	if err != nil {
		return err
	}
	defer release()
	source := osmdata.Source{Catalog: catalog}
	manager := &build.Manager{Repository: repository, Builder: network.Pipeline{Source: source, Policy: catalog.Policy}, Workers: cfg.Workers, PollInterval: time.Second, BuildTimeout: cfg.BuildTimeout}
	workerDone := make(chan error, 1)
	go func() { workerDone <- manager.Run(ctx) }()
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: httpapi.NewRouter(build.NewService(repository, source)), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: time.Minute}
	serverDone := make(chan error, 1)
	go func() { serverDone <- server.ListenAndServe() }()
	slog.Info("HTTP server started", "address", cfg.HTTPAddr, "workers", cfg.Workers)
	select {
	case <-ctx.Done():
	case err = <-serverDone:
		stop()
	case err = <-workerDone:
		stop()
		workerDone = nil
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
		slog.Error("HTTP shutdown", "error", shutdownErr)
	}
	if workerDone != nil {
		if workerErr := <-workerDone; err == nil {
			err = workerErr
		}
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
