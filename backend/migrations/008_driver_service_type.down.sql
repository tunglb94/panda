-- Rollback for 008_driver_service_type.up.sql
ALTER TABLE driver_profiles DROP COLUMN IF EXISTS service_type;
ALTER TABLE driver_profiles DROP COLUMN IF EXISTS ride_enabled;
ALTER TABLE driver_profiles DROP COLUMN IF EXISTS delivery_enabled;
