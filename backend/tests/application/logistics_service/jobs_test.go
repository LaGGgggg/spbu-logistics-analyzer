package logistics_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/logistics_service"
)

type builderFunc func(context.Context, logistics.Region) (logistics.LogisticsGraph, error)

func (f builderFunc) Build(ctx context.Context, r logistics.Region) (logistics.LogisticsGraph, error) {
	return f(ctx, r)
}

type outcomeRepository struct {
	logistics_service.Repository
	outcome        string
	saveErr        error
	saveContextErr error
	cancel         context.CancelFunc
	claimed        bool
}

func (r *outcomeRepository) Claim(ctx context.Context) (logistics_service.Job, logistics.Region, error) {

	if err := ctx.Err(); err != nil {
		return logistics_service.Job{}, logistics.Region{}, err
	}

	if r.claimed {
		return logistics_service.Job{}, logistics.Region{}, logistics_service.ErrNotFound
	}

	r.claimed = true

	return logistics_service.Job{ID: "job", RegionID: "r"}, logistics.Region{ID: "r"}, nil
}

func (r *outcomeRepository) persist(ctx context.Context, outcome string) error {

	r.outcome = outcome
	r.saveContextErr = ctx.Err()
	r.cancel()

	return r.saveErr
}

func (r *outcomeRepository) Complete(ctx context.Context, _ logistics_service.Job, _ logistics.LogisticsGraph) error {
	return r.persist(ctx, "complete")
}

func (r *outcomeRepository) Fail(ctx context.Context, _, _ string) error {
	return r.persist(ctx, "fail")
}

func (r *outcomeRepository) Requeue(ctx context.Context, _ string) error {
	return r.persist(ctx, "requeue")
}

func TestWorkerPersistsOutcome(t *testing.T) {
	for _, tc := range []struct {
		name     string
		builder  builderFunc
		shutdown bool
		saveErr  error
		want     string
	}{
		{name: "success", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{ID: "g", RegionID: "r"}, nil
		}, want: "complete"},
		{name: "panic becomes failure", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			panic("bad source")
		}, want: "fail"},
		{name: "cross-region rejected", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{ID: "g", RegionID: "other"}, nil
		}, want: "fail"},
		{name: "shutdown requeues", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{}, context.Canceled
		}, shutdown: true, want: "requeue"},
		{name: "persistence failure surfaced", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{}, errors.New("source failed")
		}, saveErr: errors.New("database failed"), want: "fail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			defer cancel()

			repo := &outcomeRepository{saveErr: tc.saveErr, cancel: cancel}
			builder := builderFunc(func(buildCtx context.Context, region logistics.Region) (logistics.LogisticsGraph, error) {
				if tc.shutdown {
					cancel()
				}

				return tc.builder(buildCtx, region)
			})
			manager := logistics_service.WorkerService{Repository: repo, Builder: builder, Workers: 1, PollInterval: time.Millisecond, BuildTimeout: time.Second}
			err := manager.Run(ctx)

			if repo.outcome != tc.want {
				t.Fatalf("outcome=%s want=%s", repo.outcome, tc.want)
			}

			if !errors.Is(err, tc.saveErr) {
				t.Fatalf("error=%v want=%v", err, tc.saveErr)
			}

			if repo.saveContextErr != nil {
				t.Fatalf("outcome persisted with cancelled context: %v", repo.saveContextErr)
			}
		})
	}
}
