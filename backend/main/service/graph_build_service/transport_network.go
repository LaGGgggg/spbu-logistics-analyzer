package graph_build_service

import (
	"context"
	"fmt"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

type arc struct {
	to       int
	distance float64
	source   int
}

// Compressed sparse rows keep arcs contiguous and avoid a string/map lookup
// for every step of every shortest-path search.
type network struct {
	points   []logistics.Coordinates
	ids      []string
	lookup   map[string]int
	offsets  []int
	arcs     []arc
	trackIDs []string
}

func buildNetwork(ctx context.Context, tracks []logistics.Track, mode logistics.TransportMode) (network, error) {

	n := network{lookup: map[string]int{}}
	degrees := []int{}

	for _, track := range tracks {
		if err := ctx.Err(); err != nil {
			return n, err
		}

		if track.Mode != mode {
			continue
		}

		if len(track.Points) < 2 {
			return n, fmt.Errorf("track %s has fewer than two points", track.ID)
		}

		previous := -1

		for _, point := range track.Points {
			if point.ID == "" {
				return n, fmt.Errorf("track point without identity")
			}

			if err := point.Coordinates.Validate(); err != nil {
				return n, err
			}

			id, exists := n.lookup[point.ID]

			if !exists {
				id = len(n.points)
				n.lookup[point.ID] = id
				n.points = append(n.points, point.Coordinates)
				n.ids = append(n.ids, point.ID)
				degrees = append(degrees, 0)
			} else if n.points[id] != point.Coordinates {
				return n, fmt.Errorf("conflicting track point %s", point.ID)
			}

			if previous >= 0 {
				degrees[previous]++

				if !track.OneWay {
					degrees[id]++
				}
			}

			previous = id
		}
	}

	n.offsets = make([]int, len(n.points)+1)

	for id, degree := range degrees {
		n.offsets[id+1] = n.offsets[id] + degree
		degrees[id] = n.offsets[id]
	}

	n.arcs = make([]arc, n.offsets[len(n.points)])

	for _, track := range tracks {
		if err := ctx.Err(); err != nil {
			return n, err
		}

		if track.Mode != mode {
			continue
		}

		source := len(n.trackIDs)
		n.trackIDs = append(n.trackIDs, track.ID)

		for i := 1; i < len(track.Points); i++ {
			a := n.lookup[track.Points[i-1].ID]
			b := n.lookup[track.Points[i].ID]
			d := distance(n.points[a], n.points[b])
			n.arcs[degrees[a]] = arc{to: b, distance: d, source: source}
			degrees[a]++

			if !track.OneWay {
				n.arcs[degrees[b]] = arc{to: a, distance: d, source: source}
				degrees[b]++
			}
		}
	}

	return n, nil
}
