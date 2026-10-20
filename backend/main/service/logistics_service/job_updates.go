package logistics_service

import (
	"context"
	"time"
)

const MaxJobWait = 25 * time.Second

type JobNotifier interface {
	NotifyJobChanged(string)
}

// Subscribe before reading the persisted state, so a concurrent commit cannot
// be lost between the initial read and registering interest in the update.
func (s *Service) subscribeJob(id string) (<-chan struct{}, func()) {

	changed := make(chan struct{})
	s.updatesMu.Lock()

	if s.updates == nil {
		s.updates = map[string]map[chan struct{}]struct{}{}
	}

	if s.updates[id] == nil {
		s.updates[id] = map[chan struct{}]struct{}{}
	}

	s.updates[id][changed] = struct{}{}
	s.updatesMu.Unlock()

	return changed, func() {
		s.updatesMu.Lock()
		defer s.updatesMu.Unlock()

		delete(s.updates[id], changed)

		if len(s.updates[id]) == 0 {
			delete(s.updates, id)
		}
	}
}

// Notify only after the repository has committed the state transition.
func (s *Service) NotifyJobChanged(id string) {

	s.updatesMu.Lock()
	defer s.updatesMu.Unlock()

	for changed := range s.updates[id] {
		close(changed)
	}

	delete(s.updates, id)
}

func (s *Service) WaitJob(ctx context.Context, id string, after time.Time, wait time.Duration) (Job, error) {

	if !validID(id) || wait < 0 || wait > MaxJobWait || (wait > 0 && after.IsZero()) {
		return Job{}, ErrInvalidInput
	}

	if wait == 0 {
		return s.GetJob(ctx, id)
	}

	changed, unsubscribe := s.subscribeJob(id)
	defer unsubscribe()

	job, err := s.GetJob(ctx, id)

	if err != nil || !job.UpdatedAt.Equal(after) || job.Status == Completed || job.Status == Failed {
		return job, err
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return Job{}, ctx.Err()
	case <-changed:
	case <-timer.C:
	}

	return s.GetJob(ctx, id)
}
