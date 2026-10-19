package graph_build_service

import (
	"context"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
)

type Builder interface {
	Build(context.Context, logistics.Region) (logistics.LogisticsGraph, error)
}

type Source interface {
	Load(context.Context, logistics.Region) (logistics.Dataset, error)
}
