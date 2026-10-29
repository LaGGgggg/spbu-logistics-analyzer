package logistics

type GraphMetrics struct {
	PageRank    map[string]float64 `json:"page_rank"`
	Betweenness map[string]float64 `json:"betweenness"`
	PowerLoad   []NodePowerLoad    `json:"power_load"`
}

type NodePowerLoad struct {
	NodeID string `json:"node_id"`
	Name   string `json:"name"`
	Degree int    `json:"degree"`
}
