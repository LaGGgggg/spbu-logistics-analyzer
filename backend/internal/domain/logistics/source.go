package logistics

type Feature struct {
	ExternalID  ExternalID        `json:"external_id"`
	Coordinates Coordinates       `json:"coordinates"`
	Tags        map[string]string `json:"tags"`
	AreaM2      float64           `json:"area_m2"`
}

type Reference struct {
	ExternalIDs []ExternalID    `json:"external_ids"`
	PhysicalID  string          `json:"physical_id"`
	Name        string          `json:"name"`
	Coordinates *Coordinates    `json:"coordinates,omitempty"`
	Modes       []TransportMode `json:"modes"`
	Cargo       bool            `json:"cargo"`
	Source      string          `json:"source"`
}

type Track struct {
	ID     string        `json:"id"`
	Mode   TransportMode `json:"mode"`
	Points []TrackPoint  `json:"points"`
	OneWay bool          `json:"one_way"`
}

type TrackPoint struct {
	ID          string      `json:"id"`
	Coordinates Coordinates `json:"coordinates"`
}

type Dataset struct {
	Features   []Feature
	Tracks     []Track
	References []Reference
	Sources    []string
	Warnings   []string
}

type Coverage struct {
	Candidates            int      `json:"candidates"`
	Accepted              int      `json:"accepted"`
	Rejected              int      `json:"rejected"`
	ReferenceTotal        int      `json:"reference_total"`
	ReferenceMatched      int      `json:"reference_matched"`
	UnmatchedReferenceIDs []string `json:"unmatched_reference_ids"`
	UnroutedNodeModes     []string `json:"unrouted_node_modes"`
	Sources               []string `json:"sources"`
	Warnings              []string `json:"warnings"`
}
