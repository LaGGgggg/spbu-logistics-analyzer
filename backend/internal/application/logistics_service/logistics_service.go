package logistics_service

import (
	"context"
	"strings"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
)

type Service struct {
	repository Repository
	areas      AreaResolver
}

func NewService(repository Repository, resolvers ...AreaResolver) *Service {

	s := &Service{repository: repository}

	if len(resolvers) > 0 {
		s.areas = resolvers[0]
	}

	return s
}

func validID(id string) bool {
	return len(id) > 0 && len(id) <= 128 && strings.TrimSpace(id) == id
}

func (s *Service) Regions(ctx context.Context) ([]logistics.Region, error) {

	return s.repository.Regions(ctx)
}

func (s *Service) RequestBuild(ctx context.Context, id string) (BuildResult, error) {

	if !validID(id) {
		return BuildResult{}, ErrInvalidInput
	}

	return s.repository.RequestBuild(ctx, id)
}

func (s *Service) GetJob(ctx context.Context, id string) (Job, error) {

	if !validID(id) {
		return Job{}, ErrInvalidInput
	}

	return s.repository.GetJob(ctx, id)
}

func (s *Service) GetGraph(ctx context.Context, id string) (logistics.LogisticsGraph, error) {

	if !validID(id) {
		return logistics.LogisticsGraph{}, ErrInvalidInput
	}

	return s.repository.GetGraph(ctx, id)
}
