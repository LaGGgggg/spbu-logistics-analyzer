package logistics_repository

import (
	"errors"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/logistics_service"
	"github.com/jackc/pgx/v5"
)

func translate(err error) error {

	if errors.Is(err, pgx.ErrNoRows) {
		return logistics_service.ErrNotFound
	}

	return err
}
