package build

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
)

type builderFunc func(context.Context, logistics.Region) (logistics.LogisticsGraph, error)

func (f builderFunc) Build(ctx context.Context, r logistics.Region) (logistics.LogisticsGraph, error) {
	return f(ctx, r)
}

type outcomeRepository struct {
	Repository
	outcome string
	saveErr error
}

func (r *outcomeRepository) Complete(context.Context, Job, logistics.LogisticsGraph) error {
	r.outcome = "complete"
	return r.saveErr
}
func (r *outcomeRepository) Fail(context.Context, string, string) error {
	r.outcome = "fail"
	return r.saveErr
}
func (r *outcomeRepository) Requeue(context.Context, string) error {
	r.outcome = "requeue"
	return r.saveErr
}
func TestWorkerPersistsOutcome(t *testing.T) {
	region := logistics.Region{ID: "r"}
	for _, tc := range []struct {
		name    string
		builder builderFunc
		cancel  bool
		saveErr error
		want    string
	}{
		{name: "success", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{ID: "g", RegionID: "r"}, nil
		}, want: "complete"},
		{name: "panic becomes failure", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) { panic("bad source") }, want: "fail"},
		{name: "cross-region rejected", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{ID: "g", RegionID: "other"}, nil
		}, want: "fail"},
		{name: "shutdown requeues", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{}, context.Canceled
		}, cancel: true, want: "requeue"},
		{name: "persistence failure surfaced", builder: func(context.Context, logistics.Region) (logistics.LogisticsGraph, error) {
			return logistics.LogisticsGraph{}, errors.New("source failed")
		}, saveErr: errors.New("database failed"), want: "fail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.cancel {
				cancel()
			}
			repo := &outcomeRepository{saveErr: tc.saveErr}
			m := Manager{Repository: repo, Builder: tc.builder, BuildTimeout: time.Second}
			err := m.execute(ctx, Job{ID: "job", RegionID: "r"}, region)
			if repo.outcome != tc.want {
				t.Fatalf("outcome=%s want=%s", repo.outcome, tc.want)
			}
			if (err != nil) != (tc.saveErr != nil) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
