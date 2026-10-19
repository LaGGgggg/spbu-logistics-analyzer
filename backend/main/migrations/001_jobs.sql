CREATE TABLE IF NOT EXISTS regions (
    id text PRIMARY KEY,
    payload jsonb NOT NULL
);
CREATE TABLE IF NOT EXISTS graphs (
    id text PRIMARY KEY,
    region_id text NOT NULL UNIQUE REFERENCES regions(id),
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS jobs (
    id text PRIMARY KEY,
    region_id text NOT NULL REFERENCES regions(id),
    status text NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    graph_id text REFERENCES graphs(id),
    error text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((status = 'completed') = (graph_id IS NOT NULL))
);
CREATE UNIQUE INDEX IF NOT EXISTS jobs_active_region ON jobs(region_id)
    WHERE status IN ('pending', 'processing');
CREATE INDEX IF NOT EXISTS jobs_pending ON jobs(created_at, id) WHERE status = 'pending';
