package logistics_repository

import (
	"context"
	"crypto/rand"
	"errors"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/logistics_service"
	"github.com/jackc/pgx/v5"
)

func (p *Repository) RequestBuild(ctx context.Context, regionID string) (logistics_service.BuildResult, error) {

	tx, err := p.pool.Begin(ctx)

	if err != nil {
		return logistics_service.BuildResult{}, err
	}

	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	var id string

	if err := tx.QueryRow(ctx, "SELECT id FROM regions WHERE id=$1 FOR UPDATE", regionID).Scan(&id); err != nil {
		return logistics_service.BuildResult{}, translate(err)
	}

	var result logistics_service.BuildResult

	err = tx.QueryRow(ctx, "SELECT id FROM graphs WHERE region_id=$1", regionID).Scan(&result.GraphID)

	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, "SELECT id FROM jobs WHERE region_id=$1 AND status IN ('pending','processing')", regionID).Scan(&result.JobID)

		if errors.Is(err, pgx.ErrNoRows) {
			result.JobID = rand.Text()
			_, err = tx.Exec(ctx, "INSERT INTO jobs(id,region_id,status) VALUES($1,$2,'pending')", result.JobID, regionID)
		}
	}

	if err != nil {
		return logistics_service.BuildResult{}, err
	}

	return result, tx.Commit(ctx)
}
