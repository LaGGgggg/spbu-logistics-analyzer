package logistics_service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

func (s *Service) RequestAreaBuild(ctx context.Context, bounds logistics.Bounds, optimization *int) (BuildResult, error) {

	if err := bounds.Validate(); err != nil {
		return BuildResult{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if optimization != nil && (*optimization < 0 || *optimization > logistics.MaximumOptimization) {
		return BuildResult{}, fmt.Errorf("%w: optimization must be between 0 and %d", ErrInvalidInput, logistics.MaximumOptimization)
	}

	if s.areas == nil {
		return BuildResult{}, ErrAreaUnavailable
	}

	source, err := s.areas.ResolveArea(ctx, bounds)

	if err != nil {
		return BuildResult{}, err
	}

	for _, v := range []*float64{&bounds.West, &bounds.South, &bounds.East, &bounds.North} {
		if *v == 0 {
			*v = 0
		}
	}

	key, _ := json.Marshal(struct {
		Bounds       logistics.Bounds
		Source       AreaSource
		Optimization *int
	}{bounds, source, optimization})

	region := logistics.Region{ID: fmt.Sprintf("area-%x", sha256.Sum256(key)), Name: fmt.Sprintf("Область %g, %g — %g, %g", bounds.West, bounds.South, bounds.East, bounds.North), Bounds: bounds, SourceID: source.ID, SourceVersion: source.Version, Optimization: optimization}
	result, err := s.repository.RequestAreaBuild(ctx, region)

	if err == nil {
		result.Region = &region
	}

	return result, err
}
