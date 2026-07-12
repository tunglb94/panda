-- Rollback for 005_dispatch.up.sql
DROP INDEX IF EXISTS dispatch_jobs_status_expires_idx;
DROP TABLE IF EXISTS dispatch_jobs;
