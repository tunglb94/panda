-- Rollback for 003_driver.up.sql
DROP INDEX IF EXISTS vehicles_driver_id_idx;
DROP TABLE IF EXISTS vehicles;
DROP TABLE IF EXISTS driver_profiles;
