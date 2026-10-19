package osmdata

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/logistics_service"
)

const overpassSource = "@overpass"

func (s Source) ResolveArea(ctx context.Context, b logistics.Bounds) (logistics_service.AreaSource, error) {

	if err := ctx.Err(); err != nil {
		return logistics_service.AreaSource{}, err
	}

	if err := b.Validate(); err != nil {
		return logistics_service.AreaSource{}, fmt.Errorf("%w: %v", logistics_service.ErrInvalidInput, err)
	}

	keys := []string{}

	for id, f := range s.Catalog.Regions {
		if f.Bounds != nil && f.Bounds.West <= b.West && f.Bounds.East >= b.East && f.Bounds.South <= b.South && f.Bounds.North >= b.North {
			keys = append(keys, id)
		}
	}

	sort.Slice(keys, func(i, j int) bool {
		a, c := s.Catalog.Regions[keys[i]].Bounds, s.Catalog.Regions[keys[j]].Bounds
		aa, ca := (a.East-a.West)*(a.North-a.South), (c.East-c.West)*(c.North-c.South)

		if aa == ca {
			return keys[i] < keys[j]
		}

		return aa < ca
	})
	id := overpassSource

	if len(keys) > 0 {
		id = keys[0]
	} else if s.Catalog.OverpassURL == "" {
		return logistics_service.AreaSource{}, logistics_service.ErrAreaUnavailable
	}

	version, _ := json.Marshal(struct {
		Files     Files
		Policy    logistics.FreightPolicy
		URL       string
		Algorithm string
	}{s.Catalog.Regions[id], s.Catalog.Policy, s.Catalog.OverpassURL, "area-v1"})

	return logistics_service.AreaSource{ID: id, Version: fmt.Sprintf("%x", sha256.Sum256(version))}, nil
}
