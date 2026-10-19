package logistics_repository

import (
	"context"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/logistics_service"
	"github.com/jackc/pgx/v5"
)

const jobColumns = "id, region_id, status, coalesce(graph_id,''), error, created_at, updated_at"

func scanJob(row pgx.Row) (logistics_service.Job, error) {

	var job logistics_service.Job

	err := row.Scan(&job.ID, &job.RegionID, &job.Status, &job.GraphID, &job.Error, &job.CreatedAt, &job.UpdatedAt)

	return job, translate(err)
}

func (p *Repository) GetJob(ctx context.Context, id string) (logistics_service.Job, error) {

	return scanJob(p.pool.QueryRow(ctx, "SELECT "+jobColumns+" FROM jobs WHERE id=$1", id))
}

func (p *Repository) Claim(ctx context.Context) (logistics_service.Job, logistics.Region, error) {

	tx, err := p.pool.Begin(ctx)

	if err != nil {
		return logistics_service.Job{}, logistics.Region{}, err
	}

	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	job, err := scanJob(tx.QueryRow(ctx, `UPDATE jobs SET status='processing', updated_at=now()
		WHERE id=(SELECT id FROM jobs WHERE status='pending' ORDER BY created_at,id FOR UPDATE SKIP LOCKED LIMIT 1)
		RETURNING `+jobColumns))

	if err != nil {
		return logistics_service.Job{}, logistics.Region{}, err
	}

	var region logistics.Region

	if err := tx.QueryRow(ctx, "SELECT payload FROM regions WHERE id=$1", job.RegionID).Scan(&region); err != nil {
		return logistics_service.Job{}, region, err
	}

	return job, region, tx.Commit(ctx)
}

func (p *Repository) Fail(ctx context.Context, id, message string) error {

	_, err := p.pool.Exec(ctx, "UPDATE jobs SET status='failed', error=$2, updated_at=now() WHERE id=$1 AND status='processing'", id, message)

	return err
}

func (p *Repository) Requeue(ctx context.Context, id string) error {

	_, err := p.pool.Exec(ctx, "UPDATE jobs SET status='pending', updated_at=now() WHERE id=$1 AND status='processing'", id)

	return err
}
