package build

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/application/network"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/buildjob"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
)

var ErrNotFound = errors.New("not found")

type Status = buildjob.Status

const (
	Pending    = buildjob.Pending
	Processing = buildjob.Processing
	Completed  = buildjob.Completed
	Failed     = buildjob.Failed
)

type Job = buildjob.Job

type BuildResult struct {
	Region  *logistics.Region `json:"region,omitempty"`
	JobID   string            `json:"job_id,omitempty"`
	GraphID string            `json:"graph_id,omitempty"`
}

type Repository interface {
	Regions(context.Context) ([]logistics.Region, error)
	RequestBuild(context.Context, string) (BuildResult, error)
	RequestAreaBuild(context.Context, logistics.Region) (BuildResult, error)
	GetJob(context.Context, string) (Job, error)
	GetGraph(context.Context, string) (logistics.LogisticsGraph, error)
	Claim(context.Context) (Job, logistics.Region, error)
	Complete(context.Context, Job, logistics.LogisticsGraph) error
	Fail(context.Context, string, string) error
	Requeue(context.Context, string) error
}

type Manager struct {
	Repository   Repository
	Builder      network.Builder
	Workers      int
	PollInterval time.Duration
	BuildTimeout time.Duration
}

func (m *Manager) Run(ctx context.Context) error {
	if m.Workers < 1 || m.PollInterval <= 0 || m.BuildTimeout <= 0 || m.Repository == nil || m.Builder == nil {
		return fmt.Errorf("invalid job manager configuration")
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

func (m *Manager) execute(ctx context.Context, job Job, region logistics.Region) error {
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
		slog.Error("build failed", "job_id", job.ID, "error", err)
		message := "Не удалось построить граф. Проверьте исходные данные и повторите попытку."
		err = m.Repository.Fail(saveCtx, job.ID, message)
	} else {
		err = m.Repository.Complete(saveCtx, job, result)
	}
	if err != nil {
		return fmt.Errorf("persist build outcome for %s: %w", job.ID, err)
	}
	return nil
}

func safeBuild(ctx context.Context, builder network.Builder, region logistics.Region) (result logistics.LogisticsGraph, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("builder panic: %v", recovered)
		}
	}()
	return builder.Build(ctx, region)
}
