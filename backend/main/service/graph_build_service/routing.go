package graph_build_service

import (
	"container/heap"
	"context"
	"fmt"
	"log/slog"
	"math"
	"slices"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

type entry struct {
	point    int
	distance float64
}

type queue struct {
	entries []entry
	ids     []string
}

func (q queue) Len() int {
	return len(q.entries)
}
func (q queue) Less(i, j int) bool {

	if q.entries[i].distance == q.entries[j].distance {
		return q.ids[q.entries[i].point] < q.ids[q.entries[j].point]
	}

	return q.entries[i].distance < q.entries[j].distance
}
func (q queue) Swap(i, j int) {
	q.entries[i], q.entries[j] = q.entries[j], q.entries[i]
}
func (q *queue) Push(x any) {
	q.entries = append(q.entries, x.(entry))
}
func (q *queue) Pop() any {

	old := q.entries
	x := old[len(old)-1]
	q.entries = old[:len(old)-1]

	return x
}

func distance(a, b logistics.Coordinates) float64 {

	const r = math.Pi / 180
	dlat := (b.Latitude - a.Latitude) * r
	dlon := (b.Longitude - a.Longitude) * r
	h := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(a.Latitude*r)*math.Cos(b.Latitude*r)*math.Sin(dlon/2)*math.Sin(dlon/2)

	return 6371.0088 * 2 * math.Asin(math.Sqrt(math.Min(1, h)))
}

func connect(ctx context.Context, g *logistics.LogisticsGraph, tracks []logistics.Track, maxSnap, geometryTolerance float64) error {

	simplified := false
	for _, mode := range []logistics.TransportMode{logistics.TransportRoad, logistics.TransportRail, logistics.TransportRiver, logistics.TransportSea, logistics.TransportAir} {
		needed := false
		for _, node := range g.Nodes {
			if slices.Contains(node.Modes, mode) {
				needed = true
				break
			}
		}
		if !needed {
			continue
		}

		n, err := buildNetwork(ctx, tracks, mode)
		if err != nil {
			return err
		}
		slog.Info("transport network prepared", "mode", mode, "points", len(n.points), "arcs", len(n.arcs))

		anchors := map[int][]int{}
		snaps := map[int]int{}
		spatialPoints := make([]spatialPoint, 0, len(n.points))

		for id, c := range n.points {
			spatialPoints = append(spatialPoints, spatialPoint{id: n.ids[id], xyz: xyz(c)})
		}

		index := newSpatialIndex(ctx, spatialPoints, 0)
		if err := ctx.Err(); err != nil {
			return err
		}

		for i, node := range g.Nodes {
			if !slices.Contains(node.Modes, mode) {
				continue
			}

			best := index.nearest(node.Coordinates, maxSnap)

			if best == "" {
				g.Coverage.UnroutedNodeModes = append(g.Coverage.UnroutedNodeModes, node.ID+":"+string(mode))
				continue
			}

			point := n.lookup[best]
			anchors[point] = append(anchors[point], i)
			snaps[i] = point
		}

		// Arrays are allocated once per transport mode and reused across all
		// origins. Epochs avoid clearing the entire regional network per origin.
		n.lookup = nil
		index = nil
		spatialPoints = nil
		dist := make([]float64, len(n.points))
		prev := make([]int, len(n.points))
		sources := make([]int, len(n.points))
		discovered := make([]uint32, len(n.points))
		visited := make([]uint32, len(n.points))
		var epoch uint32
		q := &queue{ids: n.ids}

		for i := range g.Nodes {
			if err := ctx.Err(); err != nil {
				return err
			}

			start, ok := snaps[i]

			if !ok {
				continue
			}

			epoch++
			dist[start] = 0
			discovered[start] = epoch
			q.entries = append(q.entries[:0], entry{start, 0})
			heap.Init(q)

			for q.Len() > 0 {
				if err := ctx.Err(); err != nil {
					return err
				}

				cur := heap.Pop(q).(entry)

				if visited[cur.point] == epoch {
					continue
				}

				visited[cur.point] = epoch

				if cur.point != start && len(anchors[cur.point]) > 0 {
					path := []logistics.Coordinates{}
					routeSources := []string{}

					for at := cur.point; ; at = prev[at] {
						path = append(path, n.points[at])

						if at == start {
							break
						}

						routeSources = append(routeSources, n.trackIDs[sources[at]])
					}

					slices.Reverse(path)
					reduced, err := simplifyRoute(ctx, path, geometryTolerance)
					if err != nil {
						return err
					}
					simplified = simplified || len(reduced) < len(path)
					path = reduced
					slices.Sort(routeSources)
					routeSources = slices.Compact(routeSources)

					for _, j := range anchors[cur.point] {
						g.Edges = append(g.Edges, logistics.LogisticsEdge{ID: fmt.Sprintf("%s:%s:%s", mode, g.Nodes[i].ID, g.Nodes[j].ID), FromNodeID: g.Nodes[i].ID, ToNodeID: g.Nodes[j].ID, Mode: mode, DistanceKM: cur.distance, Geometry: path, Directed: true, Source: fmt.Sprint(routeSources)})
					}

					continue
				}

				for arcIndex := n.offsets[cur.point]; arcIndex < n.offsets[cur.point+1]; arcIndex++ {
					a := n.arcs[arcIndex]
					candidate := cur.distance + a.distance
					old := dist[a.to]

					if discovered[a.to] != epoch || candidate < old {
						discovered[a.to] = epoch
						dist[a.to] = candidate
						prev[a.to] = cur.point
						sources[a.to] = a.source
						heap.Push(q, entry{a.to, candidate})
					}
				}
			}

			if len(anchors[start]) > 1 {
				g.Coverage.Warnings = append(g.Coverage.Warnings, "Several facilities share network attachment "+n.ids[start]+"; no artificial edge added")
			}
		}
	}

	if simplified {
		g.Coverage.Warnings = append(g.Coverage.Warnings, fmt.Sprintf("Route geometries longer than 32 points simplified with %.2f m projected tolerance; distance_km uses original network geometry", geometryTolerance))
	}

	return nil
}
