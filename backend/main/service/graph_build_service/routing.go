package graph_build_service

import (
	"container/heap"
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

type arc struct {
	to       string
	distance float64
	source   string
}

type network struct {
	points map[string]logistics.Coordinates
	arcs   map[string][]arc
}

type entry struct {
	id       string
	distance float64
}

type queue []entry

func (q queue) Len() int {
	return len(q)
}

func (q queue) Less(i, j int) bool {

	if q[i].distance == q[j].distance {
		return q[i].id < q[j].id
	}

	return q[i].distance < q[j].distance
}

func (q queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *queue) Push(x any) {
	*q = append(*q, x.(entry))
}

func (q *queue) Pop() any {

	old := *q
	x := old[len(old)-1]
	*q = old[:len(old)-1]

	return x
}

func distance(a, b logistics.Coordinates) float64 {

	const r = math.Pi / 180
	dlat := (b.Latitude - a.Latitude) * r
	dlon := (b.Longitude - a.Longitude) * r
	h := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(a.Latitude*r)*math.Cos(b.Latitude*r)*math.Sin(dlon/2)*math.Sin(dlon/2)

	return 6371.0088 * 2 * math.Asin(math.Sqrt(math.Min(1, h)))
}

func connect(ctx context.Context, g *logistics.LogisticsGraph, tracks []logistics.Track, maxSnap float64) error {

	for _, mode := range []logistics.TransportMode{logistics.TransportRail, logistics.TransportRiver, logistics.TransportSea} {
		if err := ctx.Err(); err != nil {
			return err
		}

		n := network{points: map[string]logistics.Coordinates{}, arcs: map[string][]arc{}}
		undirected := map[string][]arc{}

		for _, track := range tracks {
			if err := ctx.Err(); err != nil {
				return err
			}

			if track.Mode != mode {
				continue
			}

			if len(track.Points) < 2 {
				return fmt.Errorf("track %s has fewer than two points", track.ID)
			}

			for i, point := range track.Points {
				if point.ID == "" {
					return fmt.Errorf("track point without identity")
				}

				if err := point.Coordinates.Validate(); err != nil {
					return err
				}

				if old, ok := n.points[point.ID]; ok && old != point.Coordinates {
					return fmt.Errorf("conflicting track point %s", point.ID)
				}

				n.points[point.ID] = point.Coordinates

				if i == 0 {
					continue
				}

				a, b := track.Points[i-1], point
				d := distance(a.Coordinates, b.Coordinates)
				n.arcs[a.ID] = append(n.arcs[a.ID], arc{b.ID, d, track.ID})

				if !track.OneWay {
					n.arcs[b.ID] = append(n.arcs[b.ID], arc{a.ID, d, track.ID})
				}

				undirected[a.ID] = append(undirected[a.ID], arc{b.ID, d, track.ID})
				undirected[b.ID] = append(undirected[b.ID], arc{a.ID, d, track.ID})
			}
		}

		anchors := map[string][]int{}
		snaps := map[int]string{}
		spatialPoints := make([]spatialPoint, 0, len(n.points))

		for id, c := range n.points {
			spatialPoints = append(spatialPoints, spatialPoint{id: id, xyz: xyz(c)})
		}

		index := newSpatialIndex(spatialPoints, 0)

		for i, node := range g.Nodes {
			if !slices.Contains(node.Modes, mode) {
				continue
			}

			best := index.nearest(node.Coordinates, maxSnap)

			if best == "" {
				g.Coverage.UnroutedNodeModes = append(g.Coverage.UnroutedNodeModes, node.ID+":"+string(mode))
				continue
			}

			anchors[best] = append(anchors[best], i)
			snaps[i] = best
		}

		for point, attached := range anchors {
			if len(attached) > 1 {
				g.Coverage.Warnings = append(g.Coverage.Warnings, "Several facilities share network attachment "+point+"; no artificial edge added")
			}
		}

		labels := partition(undirected, anchors)

		pairs := map[[2]int]bool{}

		for from, arcs := range undirected {
			i, ok := labels[from]

			if !ok {
				continue
			}

			for _, a := range arcs {
				j, ok := labels[a.to]

				if !ok || i == j {
					continue
				}

				pair := [2]int{i, j}

				if pair[0] > pair[1] {
					pair[0], pair[1] = pair[1], pair[0]
				}

				pairs[pair] = true
			}
		}

		ordered := make([][2]int, 0, len(pairs))

		for pair := range pairs {
			ordered = append(ordered, pair)
		}

		slices.SortFunc(ordered, func(a, b [2]int) int {
			if a[0] != b[0] {
				return a[0] - b[0]
			}

			return a[1] - b[1]
		})

		for _, pair := range ordered {
			if err := ctx.Err(); err != nil {
				return err
			}

			for _, dir := range [][2]int{{pair[0], pair[1]}, {pair[1], pair[0]}} {
				d, path, routeSources, ok := n.route(dir[0], dir[1], labels, snaps)

				if !ok {
					continue
				}

				for _, f := range anchors[snaps[dir[0]]] {
					for _, t := range anchors[snaps[dir[1]]] {
						g.Edges = append(g.Edges, logistics.LogisticsEdge{ID: fmt.Sprintf("%s:%s:%s", mode, g.Nodes[f].ID, g.Nodes[t].ID), FromNodeID: g.Nodes[f].ID, ToNodeID: g.Nodes[t].ID, Mode: mode, DistanceKM: d, Geometry: path, Directed: true, Source: fmt.Sprint(routeSources)})
					}
				}
			}
		}
	}

	return nil
}

func partition(undirected map[string][]arc, anchors map[string][]int) map[string]int {

	labels := map[string]int{}
	dist := map[string]float64{}
	visited := map[string]bool{}
	q := &queue{}

	for point, attached := range anchors {
		dist[point] = 0
		labels[point] = attached[0]
		*q = append(*q, entry{point, 0})
	}

	heap.Init(q)

	for q.Len() > 0 {
		cur := heap.Pop(q).(entry)

		if visited[cur.id] {
			continue
		}

		visited[cur.id] = true

		for _, a := range undirected[cur.id] {
			candidate := cur.distance + a.distance
			old, exists := dist[a.to]
			closer := !exists || candidate < old

			if !closer && candidate == old && labels[cur.id] < labels[a.to] {
				closer = true
			}

			if !closer || visited[a.to] {
				continue
			}

			dist[a.to] = candidate
			labels[a.to] = labels[cur.id]
			heap.Push(q, entry{a.to, candidate})
		}
	}

	return labels
}

func (n network) route(from, to int, labels map[string]int, snaps map[int]string) (float64, []logistics.Coordinates, []string, bool) {

	start, target := snaps[from], snaps[to]

	if start == "" || target == "" || start == target {
		return 0, nil, nil, false
	}

	dist := map[string]float64{start: 0}
	prev := map[string]string{}
	sources := map[string]string{}
	visited := map[string]bool{}
	q := &queue{{start, 0}}
	heap.Init(q)

	for q.Len() > 0 {
		cur := heap.Pop(q).(entry)

		if visited[cur.id] {
			continue
		}

		visited[cur.id] = true

		if cur.id == target {
			path := []logistics.Coordinates{}
			routeSources := []string{}

			for at := target; ; at = prev[at] {
				path = append(path, n.points[at])

				if at == start {
					break
				}

				routeSources = append(routeSources, sources[at])
			}

			slices.Reverse(path)
			slices.Sort(routeSources)
			routeSources = slices.Compact(routeSources)

			return cur.distance, path, routeSources, true
		}

		for _, a := range n.arcs[cur.id] {
			label, ok := labels[a.to]

			if !ok || (label != from && label != to) {
				continue
			}

			candidate := cur.distance + a.distance
			old, exists := dist[a.to]

			if !exists || candidate < old {
				dist[a.to] = candidate
				prev[a.to] = cur.id
				sources[a.to] = a.source
				heap.Push(q, entry{a.to, candidate})
			}
		}
	}

	return 0, nil, nil, false
}
