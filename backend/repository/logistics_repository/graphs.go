package logistics_repository

import (
	"context"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
)

func (p *Repository) GetGraph(ctx context.Context, id string) (logistics.LogisticsGraph, error) {

	var graph logistics.LogisticsGraph

	err := p.pool.QueryRow(ctx, "SELECT payload FROM graphs WHERE id=$1", id).Scan(&graph)

	return graph, translate(err)
}
