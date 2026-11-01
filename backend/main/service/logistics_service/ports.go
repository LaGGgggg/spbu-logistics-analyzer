package logistics_service

import (
	"context"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

type Queries interface {
	Regions(context.Context) ([]logistics.Region, error)
	RequestBuild(context.Context, string) (BuildResult, error)
	RequestAreaBuild(context.Context, logistics.Bounds) (BuildResult, error)
	GetJob(context.Context, string) (Job, error)
	CancelJob(context.Context, string) (Job, error)
	GetGraph(context.Context, string) (logistics.LogisticsGraph, error)
}

// ActiveBuildCanceller interrupts a graph build that is currently executing in
// this backend process. Persistent state transitions remain the repository's
// responsibility, so cancellation is safe even when no worker owns the job.
type ActiveBuildCanceller interface {
	CancelBuild(string)
}

type AreaSource struct {
	ID      string
	Version string
}

type AreaResolver interface {
	ResolveArea(context.Context, logistics.Bounds) (AreaSource, error)
}
