package logistics_repository

import (
	"context"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/logistics_service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

var _ logistics_service.Repository = (*Repository)(nil)

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func Open(ctx context.Context, databaseURL string) (*Repository, error) {

	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, err
	}

	return New(pool), nil
}

// Close releases the connection pool created by Open.
func (p *Repository) Close() {
	p.pool.Close()
}
