package logistics_service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

func (s *Service) RequestAreaBuild(ctx context.Context, bounds logistics.Bounds) (BuildResult, error) {

	if err := bounds.Validate(); err != nil {
		return BuildResult{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
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
		Bounds logistics.Bounds
		Source AreaSource
	}{bounds, source})

	region := logistics.Region{ID: fmt.Sprintf("area-%x", sha256.Sum256(key)), Name: fmt.Sprintf("Область %g, %g — %g, %g", bounds.West, bounds.South, bounds.East, bounds.North), Bounds: bounds, SourceID: source.ID, SourceVersion: source.Version}
	result, err := s.repository.RequestAreaBuild(ctx, region)

	if err == nil {
		result.Region = &region
	}

	return result, err
}
