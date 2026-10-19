package osmdata

import (
	"context"
	"fmt"
	"io"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

func NewPBFScanner(ctx context.Context, reader io.Reader, workers int) (osm.Scanner, error) {
	if ctx == nil {
		return nil, fmt.Errorf("PBF context must not be nil")
	}
	if reader == nil {
		return nil, fmt.Errorf("PBF reader must not be nil")
	}
	if workers < 1 {
		return nil, fmt.Errorf("PBF decoder count must be positive")
	}
	return osmpbf.New(ctx, reader, workers), nil
}
