package logistics_service

import (
	"context"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
)

type Queries interface {
	Regions(context.Context) ([]logistics.Region, error)
	RequestBuild(context.Context, string) (BuildResult, error)
	RequestAreaBuild(context.Context, logistics.Bounds) (BuildResult, error)
	GetJob(context.Context, string) (Job, error)
	GetGraph(context.Context, string) (logistics.LogisticsGraph, error)
}

type AreaSource struct {
	ID      string
	Version string
}

type AreaResolver interface {
	ResolveArea(context.Context, logistics.Bounds) (AreaSource, error)
}
