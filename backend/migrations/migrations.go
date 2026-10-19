package migrations

import _ "embed"

//go:embed 001_jobs.sql
var InitialSchema string
