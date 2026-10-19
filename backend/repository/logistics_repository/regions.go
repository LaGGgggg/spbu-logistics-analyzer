package logistics_repository

import (
	"context"
	"encoding/json"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/logistics_service"
)

func (p *Repository) Regions(ctx context.Context) ([]logistics.Region, error) {

	rows, err := p.pool.Query(ctx, "SELECT payload FROM regions WHERE coalesce(payload->>'source_id','')='' ORDER BY id")

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

func (p *Repository) RequestAreaBuild(ctx context.Context, region logistics.Region) (logistics_service.BuildResult, error) {

	if err := region.Validate(); err != nil {
		return logistics_service.BuildResult{}, err
	}

	payload, err := json.Marshal(region)

	if err != nil {
		return logistics_service.BuildResult{}, err
	}

	if _, err := p.pool.Exec(ctx, "INSERT INTO regions(id,payload) VALUES($1,$2) ON CONFLICT(id) DO NOTHING", region.ID, payload); err != nil {
		return logistics_service.BuildResult{}, err
	}

	return p.RequestBuild(ctx, region.ID)
}
