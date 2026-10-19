package logistics_repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/application/logistics_service"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
)

func (p *Repository) Complete(ctx context.Context, job logistics_service.Job, graph logistics.LogisticsGraph) error {

	if graph.RegionID != job.RegionID {
		return fmt.Errorf("graph region differs from job region")
	}

	if err := graph.Validate(); err != nil {
		return err
	}

	tx, err := p.pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	if graph.Nodes == nil {
		graph.Nodes = []logistics.LogisticsNode{}
	}

	if graph.Edges == nil {
		graph.Edges = []logistics.LogisticsEdge{}
	}

	payload, err := json.Marshal(graph)

	if err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, "INSERT INTO graphs(id,region_id,payload) VALUES($1,$2,$3)", graph.ID, job.RegionID, payload); err != nil {
		return err
	}

	changed, err := tx.Exec(ctx, "UPDATE jobs SET status='completed', graph_id=$2, updated_at=now() WHERE id=$1 AND status='processing'", job.ID, graph.ID)

	if err != nil {
		return err
	}

	if changed.RowsAffected() != 1 {
		return fmt.Errorf("job is no longer processing")
	}

	return tx.Commit(ctx)
}
