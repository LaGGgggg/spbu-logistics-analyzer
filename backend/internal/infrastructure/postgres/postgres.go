package postgres

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/application/build"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct{ Pool *pgxpool.Pool }

func (p *Postgres) Initialize(ctx context.Context, regions []logistics.Region) (func(), error) {
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	var locked bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock(81743021)").Scan(&locked); err != nil || !locked {
		conn.Release()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("another backend instance is already running")
	}
	release := func() {
		// Close the dedicated session: the lock is never returned to the pool.
		_ = conn.Conn().Close(context.Background())
		conn.Release()
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		release()
		return nil, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err = tx.Exec(ctx, migrations.InitialSchema); err == nil {
		for _, region := range regions {
			var payload []byte
			payload, err = json.Marshal(region)
			if err != nil {
				break
			}
			_, err = tx.Exec(ctx, "INSERT INTO regions(id,payload) VALUES($1,$2) ON CONFLICT(id) DO UPDATE SET payload=excluded.payload", region.ID, payload)
			if err != nil {
				break
			}
		}
	}
	if err == nil {
		_, err = tx.Exec(ctx, "UPDATE jobs SET status='pending', updated_at=now() WHERE status='processing'")
	}
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		release()
		return nil, err
	}
	return release, nil
}

func (p *Postgres) Regions(ctx context.Context) ([]logistics.Region, error) {
	rows, err := p.Pool.Query(ctx, "SELECT payload FROM regions WHERE coalesce(payload->>'source_id','')='' ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []logistics.Region{}
	for rows.Next() {
		var region logistics.Region
		if err := rows.Scan(&region); err != nil {
			return nil, err
		}
		result = append(result, region)
	}
	return result, rows.Err()
}

func (p *Postgres) RequestBuild(ctx context.Context, regionID string) (build.BuildResult, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return build.BuildResult{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var id string
	// Region lock serializes simultaneous requests, including cache misses.
	if err := tx.QueryRow(ctx, "SELECT id FROM regions WHERE id=$1 FOR UPDATE", regionID).Scan(&id); err != nil {
		return build.BuildResult{}, translate(err)
	}
	var result build.BuildResult
	err = tx.QueryRow(ctx, "SELECT id FROM graphs WHERE region_id=$1", regionID).Scan(&result.GraphID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, "SELECT id FROM jobs WHERE region_id=$1 AND status IN ('pending','processing')", regionID).Scan(&result.JobID)
		if errors.Is(err, pgx.ErrNoRows) {
			result.JobID = rand.Text()
			_, err = tx.Exec(ctx, "INSERT INTO jobs(id,region_id,status) VALUES($1,$2,'pending')", result.JobID, regionID)
		}
	}
	if err != nil {
		return build.BuildResult{}, err
	}
	return result, tx.Commit(ctx)
}

func (p *Postgres) RequestAreaBuild(ctx context.Context, region logistics.Region) (build.BuildResult, error) {
	if err := region.Validate(); err != nil {
		return build.BuildResult{}, err
	}
	payload, err := json.Marshal(region)
	if err != nil {
		return build.BuildResult{}, err
	}
	if _, err := p.Pool.Exec(ctx, "INSERT INTO regions(id,payload) VALUES($1,$2) ON CONFLICT(id) DO NOTHING", region.ID, payload); err != nil {
		return build.BuildResult{}, err
	}
	return p.RequestBuild(ctx, region.ID)
}

const jobColumns = "id, region_id, status, coalesce(graph_id,''), error, created_at, updated_at"

func scanJob(row pgx.Row) (build.Job, error) {
	var job build.Job
	err := row.Scan(&job.ID, &job.RegionID, &job.Status, &job.GraphID, &job.Error, &job.CreatedAt, &job.UpdatedAt)
	return job, translate(err)
}

func (p *Postgres) GetJob(ctx context.Context, id string) (build.Job, error) {
	return scanJob(p.Pool.QueryRow(ctx, "SELECT "+jobColumns+" FROM jobs WHERE id=$1", id))
}

func (p *Postgres) GetGraph(ctx context.Context, id string) (logistics.LogisticsGraph, error) {
	var graph logistics.LogisticsGraph
	err := p.Pool.QueryRow(ctx, "SELECT payload FROM graphs WHERE id=$1", id).Scan(&graph)
	return graph, translate(err)
}

func (p *Postgres) Claim(ctx context.Context) (build.Job, logistics.Region, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return build.Job{}, logistics.Region{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	job, err := scanJob(tx.QueryRow(ctx, `UPDATE jobs SET status='processing', updated_at=now()
WHERE id=(SELECT id FROM jobs WHERE status='pending' ORDER BY created_at,id FOR UPDATE SKIP LOCKED LIMIT 1)
RETURNING `+jobColumns))
	if err != nil {
		return build.Job{}, logistics.Region{}, err
	}
	var region logistics.Region
	if err := tx.QueryRow(ctx, "SELECT payload FROM regions WHERE id=$1", job.RegionID).Scan(&region); err != nil {
		return build.Job{}, region, err
	}
	return job, region, tx.Commit(ctx)
}

func (p *Postgres) Complete(ctx context.Context, job build.Job, graph logistics.LogisticsGraph) error {
	if graph.RegionID != job.RegionID {
		return fmt.Errorf("graph region differs from job region")
	}
	if err := graph.Validate(); err != nil {
		return err
	}
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
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

func (p *Postgres) Fail(ctx context.Context, id, message string) error {
	_, err := p.Pool.Exec(ctx, "UPDATE jobs SET status='failed', error=$2, updated_at=now() WHERE id=$1 AND status='processing'", id, message)
	return err
}

func (p *Postgres) Requeue(ctx context.Context, id string) error {
	_, err := p.Pool.Exec(ctx, "UPDATE jobs SET status='pending', updated_at=now() WHERE id=$1 AND status='processing'", id)
	return err
}

func translate(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return build.ErrNotFound
	}
	return err
}
