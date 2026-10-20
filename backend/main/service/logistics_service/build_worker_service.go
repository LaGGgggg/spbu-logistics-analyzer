package logistics_service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/graph_build_service"
)

type WorkerService struct {
	Repository   Repository
	Builder      graph_build_service.Builder
	Updates      JobNotifier
	Workers      int
	PollInterval time.Duration
	BuildTimeout time.Duration
}

func (m *WorkerService) Run(ctx context.Context) error {

	if m.Workers < 1 || m.PollInterval <= 0 || m.BuildTimeout <= 0 || m.Repository == nil || m.Builder == nil {
		return fmt.Errorf("invalid build worker service configuration")
	}

	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	failures := make(chan error, m.Workers)
	var wg sync.WaitGroup

	for range m.Workers {
		wg.Go(func() {
			ticker := time.NewTicker(m.PollInterval)

			defer ticker.Stop()

			for {
				if ctx.Err() != nil {
					return
				}

				job, region, err := m.Repository.Claim(ctx)

				if err == nil {
					if m.Updates != nil {
						m.Updates.NotifyJobChanged(job.ID)
					}

					if err := m.execute(ctx, job, region); err != nil {
						failures <- err
						cancel()

						return
					}

					continue
				}

				if !errors.Is(err, ErrNotFound) && ctx.Err() == nil {
					slog.Error("claim build job", "error", err)
				}

				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}
		})
	}

	wg.Wait()

	select {
	case err := <-failures:
		return err
	default:
		return nil
	}
}

func (m *WorkerService) execute(ctx context.Context, job Job, region logistics.Region) error {

	started := time.Now()
	slog.Info("build started", "job_id", job.ID, "region_id", region.ID)
	buildCtx, cancel := context.WithTimeout(ctx, m.BuildTimeout)

	defer cancel()

	result, err := safeBuild(buildCtx, m.Builder, region)

	if err == nil {
		err = buildCtx.Err()
	}

	if err == nil {
		if result.RegionID != region.ID {
			err = fmt.Errorf("builder returned a different region")
		} else {
			err = result.Validate()
		}
	}

	saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer saveCancel()

	if ctx.Err() != nil {
		err = m.Repository.Requeue(saveCtx, job.ID)
	} else if err != nil {
		slog.Error("build failed", "job_id", job.ID, "duration", time.Since(started), "error", err)
		message := "Не удалось построить граф. Проверьте исходные данные и повторите попытку."
		err = m.Repository.Fail(saveCtx, job.ID, message)
	} else {
		err = m.Repository.Complete(saveCtx, job, result)

		if err == nil {
			slog.Info("build completed", "job_id", job.ID, "graph_id", result.ID, "duration", time.Since(started), "nodes", len(result.Nodes), "edges", len(result.Edges))
		}
	}

	if err != nil {
		return fmt.Errorf("persist build outcome for %s: %w", job.ID, err)
	}

	if m.Updates != nil {
		m.Updates.NotifyJobChanged(job.ID)
	}

	return nil
}

func safeBuild(ctx context.Context, builder graph_build_service.Builder, region logistics.Region) (result logistics.LogisticsGraph, err error) {

	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("builder panic: %v", recovered)
		}
	}()

	return builder.Build(ctx, region)
}
