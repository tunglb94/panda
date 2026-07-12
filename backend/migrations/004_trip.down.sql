-- Rollback for 004_trip.up.sql
-- NOTE: dispatch_jobs (005) does not have a real FK to trips, but drop
-- dispatch's migration first in practice since it is logically dependent.
DROP INDEX IF EXISTS trips_rider_id_idx;
DROP TABLE IF EXISTS trips;
