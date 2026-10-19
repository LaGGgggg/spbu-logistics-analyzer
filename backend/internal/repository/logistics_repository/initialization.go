package logistics_repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/migrations"
)

func (p *Repository) Initialize(ctx context.Context, regions []logistics.Region) (func(), error) {

	conn, err := p.pool.Acquire(ctx)

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
