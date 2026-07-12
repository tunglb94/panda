-- Vehicle/Service Catalog refactor: separates the product-facing
-- ServiceType (bike/bike_plus/car/car_xl) from the physical VehicleType
-- (motorcycle/car/van). Additive only — no column dropped or renamed, every
-- existing row stays valid via the defaults/backfill below.
ALTER TABLE driver_profiles ADD COLUMN service_type TEXT NOT NULL DEFAULT '';
ALTER TABLE driver_profiles ADD COLUMN ride_enabled BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE driver_profiles ADD COLUMN delivery_enabled BOOLEAN NOT NULL DEFAULT false;

-- Backfill: an earlier, since-corrected pass may have written
-- bike_plus/car_xl/delivery_bike/delivery_car into vehicle_type (which must
-- only ever hold a physical vehicle: motorcycle/car/van). Restore
-- vehicle_type to a valid physical value and populate the new service_type
-- column with what was actually intended.
UPDATE driver_profiles SET service_type = vehicle_type WHERE vehicle_type IN ('motorcycle', 'car', 'bike_plus', 'car_xl');
UPDATE driver_profiles SET vehicle_type = 'motorcycle', service_type = 'motorcycle' WHERE vehicle_type = 'delivery_bike';
UPDATE driver_profiles SET vehicle_type = 'car', service_type = 'car' WHERE vehicle_type = 'delivery_car';
UPDATE driver_profiles SET vehicle_type = 'motorcycle' WHERE vehicle_type = 'bike_plus';
UPDATE driver_profiles SET vehicle_type = 'van' WHERE vehicle_type = 'car_xl';
