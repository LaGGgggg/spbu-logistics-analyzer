package logistics

type LogisticsEdge struct {
	Directed   bool          `json:"directed"`
	Source     string        `json:"source,omitempty"`
	ID         string        `json:"id"`
	FromNodeID string        `json:"from_node_id"`
	ToNodeID   string        `json:"to_node_id"`
	Mode       TransportMode `json:"mode"`
	DistanceKM float64       `json:"distance_km"`
	Geometry   []Coordinates `json:"geometry"`
}
