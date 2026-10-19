package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
)

type Config struct {
	DatabaseURL  string
	HTTPAddr     string
	Workers      int
	BuildTimeout time.Duration
	SourcesFile  string
	Regions      []logistics.Region
}

func Load() (Config, error) {

	result := Config{
		DatabaseURL: value("DATABASE_URL", "postgres://logistics:logistics@localhost:5432/logistics?sslmode=disable"),
		HTTPAddr:    value("HTTP_ADDR", ":8080"),
		SourcesFile: value("SOURCES_FILE", "configs/sources.json"),
	}

	var err error

	result.Workers, err = strconv.Atoi(value("BUILD_WORKERS", "2"))

	if err != nil || result.Workers < 1 || result.Workers > 8 {
		return result, fmt.Errorf("BUILD_WORKERS must be between 1 and 8")
	}

	result.BuildTimeout, err = time.ParseDuration(value("BUILD_TIMEOUT", "30m"))

	if err != nil || result.BuildTimeout <= 0 {
		return result, fmt.Errorf("BUILD_TIMEOUT must be a positive duration")
	}

	content, err := os.ReadFile(value("REGIONS_FILE", "configs/regions.json"))

	if err != nil {
		return result, fmt.Errorf("read regions: %w", err)
	}

	if err = json.Unmarshal(content, &result.Regions); err != nil {
		return result, fmt.Errorf("parse regions: %w", err)
	}

	seen := make(map[string]bool)

	for _, region := range result.Regions {
		if err := region.Validate(); err != nil {
			return result, fmt.Errorf("region %s: %w", region.ID, err)
		}

		if seen[region.ID] {
			return result, fmt.Errorf("duplicate region %s", region.ID)
		}

		seen[region.ID] = true
	}

	return result, nil
}

func value(key, fallback string) string {

	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
