package osmdata

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
)

type Files struct {
	OSM        string `json:"osm"`
	URL        string `json:"url,omitempty"`
	SHA256     string `json:"sha256,omitempty"`
	References string `json:"references,omitempty"`
	Routes     string `json:"routes,omitempty"`
}
type Catalog struct {
	Regions        map[string]Files        `json:"regions"`
	Policy         logistics.FreightPolicy `json:"policy"`
	MaximumObjects int                     `json:"maximum_objects"`
}
type Source struct{ Catalog Catalog }

func LoadCatalog(path string) (Catalog, error) {
	c := Catalog{Policy: logistics.DefaultFreightPolicy(), MaximumObjects: 5000000}
	f, err := os.Open(path)
	if err != nil {
		return c, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return c, err
	}
	if err := c.Policy.Validate(); err != nil {
		return c, err
	}
	if c.MaximumObjects < 1 {
		return c, fmt.Errorf("maximum_objects must be positive")
	}
	if len(c.Regions) == 0 {
		return c, fmt.Errorf("source catalog has no regions")
	}
	for id, files := range c.Regions {
		if files.OSM == "" {
			return c, fmt.Errorf("region %s has no OSM file", id)
		}
		for _, file := range []*string{&files.OSM, &files.References, &files.Routes} {
			if *file != "" && !filepath.IsAbs(*file) {
				*file = filepath.Join(filepath.Dir(path), *file)
			}
		}
		c.Regions[id] = files
	}
	return c, nil
}

func (s Source) Load(ctx context.Context, region logistics.Region) (logistics.Dataset, error) {
	files, ok := s.Catalog.Regions[region.ID]
	if !ok {
		return logistics.Dataset{}, fmt.Errorf("no source configured for region %s", region.ID)
	}
	if err := ensureExtract(ctx, files); err != nil {
		return logistics.Dataset{}, err
	}
	dataset, err := readOSM(ctx, files.OSM, s.Catalog.MaximumObjects)
	if err != nil {
		return dataset, fmt.Errorf("OSM: %w", err)
	}
	dataset.Sources = []string{filepath.Base(files.OSM), "OpenStreetMap contributors (ODbL)"}
	if files.References != "" {
		if err := readJSON(files.References, &dataset.References); err != nil {
			return dataset, err
		}
		filtered := dataset.References[:0]
		for _, ref := range dataset.References {
			invalidESR := false
			for _, id := range ref.ExternalIDs {
				if id.Namespace == logistics.ExternalESR && logistics.ESRKey(id.Value) == "" {
					invalidESR = true
				}
			}
			if invalidESR {
				dataset.Warnings = append(dataset.Warnings, "Ignored nonphysical ESR registry record: "+ref.Name)
				continue
			}
			filtered = append(filtered, ref)
			if ref.Source == "" || len(ref.ExternalIDs) == 0 || len(ref.Modes) == 0 {
				return dataset, fmt.Errorf("reference requires source, external_ids and modes")
			}
			for _, id := range ref.ExternalIDs {
				if err := id.Validate(); err != nil {
					return dataset, err
				}
			}
			for _, mode := range ref.Modes {
				if !mode.Valid() {
					return dataset, fmt.Errorf("invalid reference mode %s", mode)
				}
			}
			if ref.Coordinates != nil {
				if err := ref.Coordinates.Validate(); err != nil {
					return dataset, err
				}
			}
		}
		dataset.References = filtered
		dataset.Sources = append(dataset.Sources, filepath.Base(files.References))
	} else {
		dataset.Warnings = append(dataset.Warnings, "No reference registry: untagged freight facilities cannot be confirmed")
	}
	if files.Routes != "" {
		var tracks []logistics.Track
		if err := readJSON(files.Routes, &tracks); err != nil {
			return dataset, err
		}
		for _, track := range tracks {
			if !track.Mode.Valid() || strings.TrimSpace(track.ID) == "" {
				return dataset, fmt.Errorf("invalid route dataset track")
			}
		}
		dataset.Tracks = append(dataset.Tracks, tracks...)
		dataset.Sources = append(dataset.Sources, filepath.Base(files.Routes))
	} else {
		dataset.Warnings = append(dataset.Warnings, "No sea/air route dataset; no inferred sea or air connections")
	}
	return dataset, ctx.Err()
}
func readJSON(path string, target any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	return d.Decode(target)
}
