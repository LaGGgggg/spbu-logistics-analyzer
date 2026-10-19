package osmdata

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

func areaQuery(b logistics.Bounds) string {

	bbox := fmt.Sprintf("(%g,%g,%g,%g)", b.South, b.West, b.North, b.East)
	selectors := []string{
		`nwr["railway"~"^(station|yard|container_terminal)$"]`,
		`nwr["industrial"~"^(warehouse|logistics|port|intermodal_freight_terminal)$"]`,
		`nwr["cargo"]`, `nwr["port"]`, `nwr["harbour"="yes"]`, `nwr["landuse"="port"]`,
		`nwr["aeroway"~"^(aerodrome|cargo|cargo_terminal|terminal)$"]`,
		`nwr["building"="warehouse"]`, `nwr["man_made"="distribution_centre"]`,
		`way["highway"~"^(motorway|motorway_link|trunk|trunk_link|primary|primary_link|secondary|secondary_link|tertiary|tertiary_link|unclassified|residential|service|living_street)$"]`,
		`way["railway"="rail"]`, `way["waterway"~"^(river|canal)$"]`,
	}
	q := "[out:xml][timeout:180];("

	for _, selector := range selectors {
		q += selector + bbox + ";"
	}

	return q + ");(._;>>;);out body;"
}

func (s Source) downloadArea(ctx context.Context, region logistics.Region) (string, error) {

	if err := region.Bounds.Validate(); err != nil {
		return "", err
	}

	if !strings.HasPrefix(region.ID, "area-") || strings.ContainsAny(region.ID, `/\`) {
		return "", fmt.Errorf("invalid generated area identity")
	}

	if s.Catalog.CacheDirectory == "" || s.Catalog.OverpassURL == "" {
		return "", fmt.Errorf("overpass source is not configured")
	}

	path := filepath.Join(s.Catalog.CacheDirectory, region.ID+".osm")
	lock, _ := downloads.LoadOrStore(path, make(chan struct{}, 1))
	semaphore := lock.(chan struct{})

	select {
	case semaphore <- struct{}{}:
	case <-ctx.Done():
		return "", ctx.Err()
	}

	defer func() {
		<-semaphore
	}()

	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	form := url.Values{"data": {areaQuery(region.Bounds)}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.Catalog.OverpassURL, strings.NewReader(form.Encode()))

	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "SPBU-Logistics-Analyzer/1.0")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("overpass HTTP %d", resp.StatusCode)
	}

	if err := os.MkdirAll(s.Catalog.CacheDirectory, 0755); err != nil {
		return "", err
	}

	f, err := os.CreateTemp(s.Catalog.CacheDirectory, ".area-*")

	if err != nil {
		return "", err
	}

	defer os.Remove(f.Name())

	n, copyErr := io.Copy(f, io.LimitReader(resp.Body, maximumDownloadBytes+1))
	closeErr := f.Close()

	if copyErr != nil {
		return "", copyErr
	}

	if closeErr != nil {
		return "", closeErr
	}

	if n > maximumDownloadBytes {
		return "", fmt.Errorf("overpass response exceeds download limit")
	}

	if err := validateOverpass(ctx, f.Name()); err != nil {
		return "", err
	}

	if err := os.Rename(f.Name(), path); err != nil {
		return "", err
	}

	return path, nil
}

func validateOverpass(ctx context.Context, path string) error {

	f, err := os.Open(path)

	if err != nil {
		return err
	}

	defer f.Close()

	d := xml.NewDecoder(f)
	root := false

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		token, err := d.Token()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if e, ok := token.(xml.StartElement); ok {
			if !root {
				if e.Name.Local != "osm" {
					return fmt.Errorf("overpass response is not OSM XML")
				}

				root = true
			}

			if e.Name.Local == "remark" || e.Name.Local == "error" {
				return fmt.Errorf("overpass returned an incomplete result or runtime error")
			}
		}
	}

	if !root {
		return fmt.Errorf("empty Overpass response")
	}

	return nil
}
