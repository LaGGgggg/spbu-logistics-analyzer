package graph_build_service

import (
	"container/heap"
	"math"
	"sort"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

type metricEdge struct {
	to     string
	weight float64
}

type dijkstraItem struct {
	node string
	dist float64
}

type dijkstraQueue []dijkstraItem

func (q dijkstraQueue) Len() int           { return len(q) }
func (q dijkstraQueue) Less(i, j int) bool { return q[i].dist < q[j].dist }
func (q dijkstraQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *dijkstraQueue) Push(x any)        { *q = append(*q, x.(dijkstraItem)) }
func (q *dijkstraQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[:n-1]
	return item
}

// CalculateMetrics computes PageRank, Betweenness centrality, and Power Load for the graph.
func CalculateMetrics(g *logistics.LogisticsGraph) {
	if g == nil {
		return
	}

	n := len(g.Nodes)
	if n == 0 {
		g.Metrics = &logistics.GraphMetrics{
			PageRank:    map[string]float64{},
			Betweenness: map[string]float64{},
			PowerLoad:   []logistics.NodePowerLoad{},
		}
		return
	}

	nodeMap := make(map[string]int, n)
	for i, node := range g.Nodes {
		nodeMap[node.ID] = i
	}

	// 1. Build adjacency
	// Undirected neighbors for degree/power load
	neighbors := make(map[string]map[string]bool, n)
	// Directed adjacency for PageRank and Betweenness
	outMap := make(map[string]map[string]float64, n)
	inMap := make(map[string]map[string]bool, n)

	for _, node := range g.Nodes {
		neighbors[node.ID] = make(map[string]bool)
		outMap[node.ID] = make(map[string]float64)
		inMap[node.ID] = make(map[string]bool)
	}

	for _, edge := range g.Edges {
		u, v := edge.FromNodeID, edge.ToNodeID
		if u == v || nodeMap[u] == 0 && u != g.Nodes[0].ID || nodeMap[v] == 0 && v != g.Nodes[0].ID {
			if _, ok := nodeMap[u]; !ok {
				continue
			}
			if _, ok := nodeMap[v]; !ok {
				continue
			}
		}

		w := edge.DistanceKM
		if w <= 0 {
			w = 1.0
		}

		neighbors[u][v] = true
		neighbors[v][u] = true

		if existing, ok := outMap[u][v]; !ok || w < existing {
			outMap[u][v] = w
		}
		inMap[v][u] = true

		if !edge.Directed {
			if existing, ok := outMap[v][u]; !ok || w < existing {
				outMap[v][u] = w
			}
			inMap[u][v] = true
		}
	}

	// 2. Power Load (degree of each node, sorted in ascending order)
	powerLoads := make([]logistics.NodePowerLoad, 0, n)
	for i := range g.Nodes {
		nodeID := g.Nodes[i].ID
		degree := len(neighbors[nodeID])
		g.Nodes[i].PowerLoad = degree
		powerLoads = append(powerLoads, logistics.NodePowerLoad{
			NodeID: nodeID,
			Name:   g.Nodes[i].Name,
			Degree: degree,
		})
	}

	sort.Slice(powerLoads, func(i, j int) bool {
		if powerLoads[i].Degree == powerLoads[j].Degree {
			return powerLoads[i].Name < powerLoads[j].Name
		}
		return powerLoads[i].Degree < powerLoads[j].Degree
	})

	// 3. PageRank
	d := 0.85
	pr := make(map[string]float64, n)
	initPR := 1.0 / float64(n)
	for _, node := range g.Nodes {
		pr[node.ID] = initPR
	}

	for iter := 0; iter < 100; iter++ {
		danglingSum := 0.0
		for _, node := range g.Nodes {
			if len(outMap[node.ID]) == 0 {
				danglingSum += pr[node.ID]
			}
		}

		nextPR := make(map[string]float64, n)
		diff := 0.0

		for _, node := range g.Nodes {
			sumIn := 0.0
			for inNode := range inMap[node.ID] {
				outCount := len(outMap[inNode])
				if outCount > 0 {
					sumIn += pr[inNode] / float64(outCount)
				}
			}

			val := (1.0-d)/float64(n) + d*(danglingSum/float64(n)+sumIn)
			nextPR[node.ID] = val
			diff += math.Abs(val - pr[node.ID])
		}

		pr = nextPR
		if diff < 1e-7 {
			break
		}
	}

	for i := range g.Nodes {
		val := math.Round(pr[g.Nodes[i].ID]*100000) / 100000
		g.Nodes[i].PageRank = val
		pr[g.Nodes[i].ID] = val
	}

	// 4. Betweenness Centrality (Brandes' Algorithm)
	cb := make(map[string]float64, n)
	for _, node := range g.Nodes {
		cb[node.ID] = 0.0
	}

	if n > 2 {
		for _, sNode := range g.Nodes {
			s := sNode.ID
			stack := make([]string, 0, n)
			p := make(map[string][]string, n)
			sigma := make(map[string]float64, n)
			dist := make(map[string]float64, n)
			visited := make(map[string]bool, n)

			for _, node := range g.Nodes {
				dist[node.ID] = math.Inf(1)
				sigma[node.ID] = 0
			}

			dist[s] = 0
			sigma[s] = 1

			pq := &dijkstraQueue{}
			heap.Init(pq)
			heap.Push(pq, dijkstraItem{node: s, dist: 0})

			for pq.Len() > 0 {
				cur := heap.Pop(pq).(dijkstraItem)
				v := cur.node
				if visited[v] {
					continue
				}
				visited[v] = true
				stack = append(stack, v)

				for w, weight := range outMap[v] {
					alt := dist[v] + weight
					if alt < dist[w]-1e-9 {
						dist[w] = alt
						sigma[w] = sigma[v]
						p[w] = []string{v}
						heap.Push(pq, dijkstraItem{node: w, dist: alt})
					} else if math.Abs(alt-dist[w]) <= 1e-9 {
						sigma[w] += sigma[v]
						p[w] = append(p[w], v)
					}
				}
			}

			delta := make(map[string]float64, n)
			for i := len(stack) - 1; i >= 0; i-- {
				w := stack[i]
				for _, v := range p[w] {
					if sigma[w] > 0 {
						delta[v] += (sigma[v] / sigma[w]) * (1.0 + delta[w])
					}
				}
				if w != s {
					cb[w] += delta[w]
				}
			}
		}

		// Normalize betweenness
		normFactor := float64((n - 1) * (n - 2))
		for k := range cb {
			cb[k] /= normFactor
			cb[k] = math.Round(cb[k]*100000) / 100000
		}
	}

	for i := range g.Nodes {
		val := cb[g.Nodes[i].ID]
		g.Nodes[i].Betweenness = val
	}

	g.Metrics = &logistics.GraphMetrics{
		PageRank:    pr,
		Betweenness: cb,
		PowerLoad:   powerLoads,
	}
}
