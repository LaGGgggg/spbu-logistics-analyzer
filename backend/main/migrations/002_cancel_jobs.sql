-- PostgreSQL gives unnamed inline checks deterministic names. Recreate them so
-- installations initialized by 001_jobs.sql also accept the cancelled status.
ALTER TABLE jobs DROP CONSTRAINT IF EXISTS jobs_status_check;
ALTER TABLE jobs ADD CONSTRAINT jobs_status_check
    CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled'));

ALTER TABLE jobs DROP CONSTRAINT IF EXISTS jobs_check;
ALTER TABLE jobs DROP CONSTRAINT IF EXISTS jobs_graph_id_check;
ALTER TABLE jobs ADD CONSTRAINT jobs_graph_id_check
    CHECK ((status = 'completed') = (graph_id IS NOT NULL));
