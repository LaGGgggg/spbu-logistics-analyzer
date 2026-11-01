package logistics_service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/logistics_service"
)

type cancellableRepository struct {
	logistics_service.Repository

	mu      sync.Mutex
	job     logistics_service.Job
	claimed bool
	outcome string
}

func (r *cancellableRepository) Claim(ctx context.Context) (logistics_service.Job, logistics.Region, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return logistics_service.Job{}, logistics.Region{}, err
	}

	if r.claimed {
		return logistics_service.Job{}, logistics.Region{}, logistics_service.ErrNotFound
	}

	r.claimed = true
	r.job.Status = logistics_service.Processing

	return r.job, logistics.Region{ID: r.job.RegionID}, nil
}

func (r *cancellableRepository) GetJob(_ context.Context, _ string) (logistics_service.Job, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.job, nil
}

func (r *cancellableRepository) CancelJob(_ context.Context, id string) (logistics_service.Job, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	if id != r.job.ID {
		return logistics_service.Job{}, logistics_service.ErrNotFound
	}

	if r.job.Status == logistics_service.Pending || r.job.Status == logistics_service.Processing {
		r.job.Status = logistics_service.Cancelled
	}

	return r.job, nil
}

func (r *cancellableRepository) Complete(context.Context, logistics_service.Job, logistics.LogisticsGraph) error {

	r.mu.Lock()
	defer r.mu.Unlock()
	r.outcome = "complete"

	return nil
}

func (r *cancellableRepository) Fail(context.Context, string, string) error {

	r.mu.Lock()
	defer r.mu.Unlock()
	r.outcome = "fail"

	return nil
}

func (r *cancellableRepository) Requeue(_ context.Context, id string) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	// The PostgreSQL repository requeues only a processing job. A cancellation
	// that won the race must therefore stay cancelled even during shutdown.
	if id != r.job.ID || r.job.Status != logistics_service.Processing {
		return nil
	}

	r.job.Status = logistics_service.Pending
	r.outcome = "requeue"

	return nil
}

func TestWorkerCancellationStopsActiveBuild(t *testing.T) {

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	repository := &cancellableRepository{job: logistics_service.Job{ID: "job", RegionID: "region", Status: logistics_service.Pending}}
	started := make(chan struct{})
	interrupted := make(chan struct{})
	builder := builderFunc(func(ctx context.Context, _ logistics.Region) (logistics.LogisticsGraph, error) {
		close(started)
		<-ctx.Done()
		close(interrupted)

		return logistics.LogisticsGraph{}, ctx.Err()
	})
	worker := &logistics_service.WorkerService{Repository: repository, Builder: builder, Workers: 1, PollInterval: time.Millisecond, BuildTimeout: time.Second}
	service := logistics_service.NewService(repository)
	service.SetActiveBuildCanceller(worker)
	done := make(chan error, 1)

	go func() {
		done <- worker.Run(ctx)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start build")
	}

	job, err := service.CancelJob(context.Background(), "job")

	if err != nil {
		t.Fatal(err)
	}

	if job.Status != logistics_service.Cancelled {
		t.Fatalf("status=%s", job.Status)
	}

	select {
	case <-interrupted:
	case <-time.After(time.Second):
		t.Fatal("active build ignored cancellation")
	}

	stop()

	if err := <-done; err != nil {
		t.Fatal(err)
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	if repository.outcome != "" {
		t.Fatalf("cancelled build persisted outcome %q", repository.outcome)
	}
}
