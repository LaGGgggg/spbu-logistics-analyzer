package logistics_service_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/graph_build_service"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/logistics_service"
)

type builderFunc func(context.Context, logistics.Region) (logistics.LogisticsGraph, error)

func (f builderFunc) Build(ctx context.Context, r logistics.Region) (logistics.LogisticsGraph, error) {
	return f(ctx, r)
}

type outcomeRepository struct {
	logistics_service.Repository
	outcome        string
	message        string
	saveErr        error
	completeErr    error
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
	if r.completeErr != nil {
		return r.completeErr
	}
	return r.persist(ctx, "complete")
}

func (r *outcomeRepository) Fail(ctx context.Context, _ string, message string) error {
	r.message = message
	return r.persist(ctx, "fail")
}

func (r *outcomeRepository) Requeue(ctx context.Context, _ string) error {
	return r.persist(ctx, "requeue")
}

type recordedUpdates struct {
	repo   *outcomeRepository
	states []string
}

func (n *recordedUpdates) NotifyJobChanged(id string) {
	n.states = append(n.states, id+":"+n.repo.outcome)
}

func TestWorkerPersistsOutcome(t *testing.T) {
	for _, tc := range []struct {
		name          string
		builder       builderFunc
		shutdown      bool
		saveErr       error
		completeErr   error
		want          string
		publicMessage string
	}{
		{name: "success", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{ID: "g", RegionID: "r"}, nil
		}, want: "complete"},
		{name: "data limit has actionable message", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{}, fmt.Errorf("private filename: %w", graph_build_service.ErrDatasetLimit)
		}, want: "fail", publicMessage: "Превышен лимит"},
		{name: "timeout has actionable message", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{}, context.DeadlineExceeded
		}, want: "fail", publicMessage: "BUILD_TIMEOUT"},
		{name: "oversized result fails job without stopping worker", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{ID: "g", RegionID: "r"}, nil
		}, want: "fail", completeErr: logistics_service.ErrGraphTooLarge, publicMessage: "route_geometry_tolerance_m"},
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

			repo := &outcomeRepository{saveErr: tc.saveErr, completeErr: tc.completeErr, cancel: cancel}
			builder := builderFunc(func(buildCtx context.Context, region logistics.Region) (logistics.LogisticsGraph, error) {
				if tc.shutdown {
					cancel()
				}

				return tc.builder(buildCtx, region)
			})
			updates := &recordedUpdates{repo: repo}
			manager := logistics_service.WorkerService{Repository: repo, Updates: updates, Builder: builder, Workers: 1, PollInterval: time.Millisecond, BuildTimeout: time.Second}
			err := manager.Run(ctx)

			if tc.publicMessage != "" && (!strings.Contains(repo.message, tc.publicMessage) || strings.Contains(repo.message, "private filename")) {
				t.Fatalf("unsafe or unhelpful public error: %s", repo.message)
			}
			if repo.outcome != tc.want {
				t.Fatalf("outcome=%s want=%s", repo.outcome, tc.want)
			}

			if !errors.Is(err, tc.saveErr) {
				t.Fatalf("error=%v want=%v", err, tc.saveErr)
			}

			expectedNotifications := 2
			if tc.saveErr != nil {
				expectedNotifications = 1
			}
			if len(updates.states) != expectedNotifications || updates.states[0] != "job:" {
				t.Fatalf("invalid notifications: %v", updates.states)
			}
			if tc.saveErr == nil && updates.states[1] != "job:"+tc.want {
				t.Fatalf("notification published before persistence: %v", updates.states)
			}

			if repo.saveContextErr != nil {
				t.Fatalf("outcome persisted with cancelled context: %v", repo.saveContextErr)
			}
		})
	}
}
