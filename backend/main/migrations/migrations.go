package migrations

import _ "embed"

//go:embed 001_jobs.sql
var InitialSchema string

//go:embed 002_cancel_jobs.sql
var CancelJobs string
