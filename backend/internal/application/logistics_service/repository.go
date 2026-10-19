package logistics_service

import (
	"context"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
)

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
