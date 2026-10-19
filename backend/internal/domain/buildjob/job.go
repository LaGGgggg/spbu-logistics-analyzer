package buildjob

import "time"

type Status string

const (
	Pending    Status = "pending"
	Processing Status = "processing"
	Completed  Status = "completed"
	Failed     Status = "failed"
)

type Job struct {
	ID        string    `json:"id"`
	RegionID  string    `json:"region_id"`
	Status    Status    `json:"status"`
	GraphID   string    `json:"graph_id,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
