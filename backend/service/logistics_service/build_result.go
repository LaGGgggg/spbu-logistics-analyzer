package logistics_service

import (
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/buildjob"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
)

type Status = buildjob.Status

const (
	Pending    = buildjob.Pending
	Processing = buildjob.Processing
	Completed  = buildjob.Completed
	Failed     = buildjob.Failed
)

type Job = buildjob.Job

type BuildResult struct {
	Region  *logistics.Region `json:"region,omitempty"`
	JobID   string            `json:"job_id,omitempty"`
	GraphID string            `json:"graph_id,omitempty"`
}
