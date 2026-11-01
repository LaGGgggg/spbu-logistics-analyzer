package logistics_repository

import (
	"context"
	"errors"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/logistics_service"
)

// CancelJob atomically transitions an active job to cancelled. Repeating a
// cancellation is idempotent; terminal jobs that completed or failed cannot be
// changed retroactively.
func (p *Repository) CancelJob(ctx context.Context, id string) (logistics_service.Job, error) {

	job, err := scanJob(p.pool.QueryRow(ctx, `UPDATE jobs
		SET status='cancelled', error='', updated_at=now()
		WHERE id=$1 AND status IN ('pending', 'processing')
		RETURNING `+jobColumns, id))

	if err == nil {
		return job, nil
	}

	if !errors.Is(err, logistics_service.ErrNotFound) {
		return logistics_service.Job{}, err
	}

	job, err = p.GetJob(ctx, id)

	if err != nil {
		return logistics_service.Job{}, err
	}

	if job.Status == logistics_service.Cancelled {
		return job, nil
	}

	return logistics_service.Job{}, logistics_service.ErrJobNotCancellable
}
