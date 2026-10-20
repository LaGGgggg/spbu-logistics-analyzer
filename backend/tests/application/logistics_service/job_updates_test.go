package logistics_service_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/logistics_service"
)

type watchedRepository struct {
	logistics_service.Repository
	mu     sync.Mutex
	job    logistics_service.Job
	reads  chan struct{}
	onRead func()
}

func (r *watchedRepository) GetJob(ctx context.Context, _ string) (logistics_service.Job, error) {

	if err := ctx.Err(); err != nil {
		return logistics_service.Job{}, err
	}

	r.mu.Lock()
	job := r.job
	hook := r.onRead
	r.onRead = nil
	r.mu.Unlock()

	if hook != nil {
		hook()
	}

	r.reads <- struct{}{}

	return job, nil
}

func (r *watchedRepository) complete() {

	r.mu.Lock()
	defer r.mu.Unlock()

	r.job.Status = logistics_service.Completed
	r.job.GraphID = "graph"
	r.job.UpdatedAt = r.job.UpdatedAt.Add(time.Second)
}

func watchedJob() *watchedRepository {
	return &watchedRepository{job: logistics_service.Job{ID: "job", Status: logistics_service.Processing, UpdatedAt: time.Now().UTC()}, reads: make(chan struct{}, 64)}
}

func TestWaitJobNotifiesAllClientsWithoutDatabasePolling(t *testing.T) {

	repo := watchedJob()
	service := logistics_service.NewService(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	results := make(chan error, 8)

	for range 8 {
		go func() {
			job, err := service.WaitJob(ctx, "job", repo.job.UpdatedAt, time.Second)

			if err == nil && (job.Status != logistics_service.Completed || job.GraphID != "graph") {
				err = errors.New("client missed completed state")
			}

			results <- err
		}()
	}

	for range 8 {
		select {
		case <-repo.reads:
		case <-ctx.Done():
			t.Fatal("clients did not start waiting")
		}
	}

	repo.complete()
	service.NotifyJobChanged("unrelated-job")
	service.NotifyJobChanged("job")
	service.NotifyJobChanged("job")

	for range 8 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}

	if len(repo.reads) != 8 {
		t.Fatalf("expected one final read per client, got %d", len(repo.reads))
	}
}

func TestWaitJobDoesNotLoseUpdateDuringInitialRead(t *testing.T) {

	repo := watchedJob()
	service := logistics_service.NewService(repo)
	repo.onRead = func() {
		repo.complete()
		service.NotifyJobChanged("job")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	job, err := service.WaitJob(ctx, "job", repo.job.UpdatedAt, time.Second)

	if err != nil || job.Status != logistics_service.Completed {
		t.Fatalf("update between subscribe and initial read lost: %+v, %v", job, err)
	}
}

func TestWaitJobTimeoutCancellationAndImmediateStates(t *testing.T) {

	repo := watchedJob()
	service := logistics_service.NewService(repo)
	job, err := service.WaitJob(context.Background(), "job", repo.job.UpdatedAt, time.Millisecond)

	if err != nil || job.Status != logistics_service.Processing || len(repo.reads) != 2 {
		t.Fatalf("timeout must return latest snapshot with two reads: %+v, %v", job, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := service.WaitJob(ctx, "job", repo.job.UpdatedAt, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled client kept waiting: %v", err)
	}

	// Notification after cancellation must be safe and must not block a worker.
	service.NotifyJobChanged("job")

	if _, err := service.WaitJob(context.Background(), "job", time.Time{}, time.Second); !errors.Is(err, logistics_service.ErrInvalidInput) {
		t.Fatalf("missing cursor accepted: %v", err)
	}

	if _, err := service.WaitJob(context.Background(), "job", repo.job.UpdatedAt, 26*time.Second); !errors.Is(err, logistics_service.ErrInvalidInput) {
		t.Fatalf("unbounded wait accepted: %v", err)
	}

	for _, status := range []logistics_service.Status{logistics_service.Completed, logistics_service.Failed} {
		repo.job.Status = status
		job, err := service.WaitJob(context.Background(), "job", repo.job.UpdatedAt, time.Second)

		if err != nil || job.Status != status {
			t.Fatalf("terminal status must return immediately: %+v, %v", job, err)
		}
	}
}
