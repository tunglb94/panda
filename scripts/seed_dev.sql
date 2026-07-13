-- Development seed data for FAIRRIDE local environment.
-- Provides one Rider account, one verified Driver account, and one Admin
-- account (for the Admin Review Dashboard — no provisioning UI exists yet,
-- see gateway/http/handlers/auth_handler.go's AdminLogin doc comment).
-- Auth is password-less: log in with just the phone number.
--
-- Default logins (see README § Development Accounts):
--   Rider  → +84900000001
--   Driver → +84900000002
--   Admin  → +84900000099
--
-- Run against the identity + driver databases:
--   psql $IDENTITY_DB_URL -f scripts/seed_dev.sql
--   psql $DRIVER_DB_URL   -f scripts/seed_dev.sql

-- ─── Roles ───────────────────────────────────────────────────────────────────

INSERT INTO identity_roles (id, name, description, is_system, created_at, updated_at)
VALUES
  ('role-rider-default',  'rider',  'Default rider role',  TRUE, NOW(), NOW()),
  ('role-driver-default', 'driver', 'Default driver role', TRUE, NOW(), NOW()),
  ('role-admin-default',  'admin',  'Default admin role',  TRUE, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ─── Users ───────────────────────────────────────────────────────────────────

INSERT INTO identity_users (id, phone_number, name, email, type, status, role_id, created_at, updated_at)
VALUES
  ('user-rider-dev-01',  '+84900000001', 'Dev Rider',  '', 'rider',  'active', 'role-rider-default',  NOW(), NOW()),
  ('user-driver-dev-01', '+84900000002', 'Dev Driver', '', 'driver', 'active', 'role-driver-default', NOW(), NOW()),
  ('user-admin-dev-01',  '+84900000099', 'Dev Admin',  '', 'admin',  'active', 'role-admin-default',  NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ─── Driver profile ──────────────────────────────────────────────────────────

INSERT INTO driver_profiles (
  driver_id, user_id, license_number, vehicle_type,
  vehicle_brand, vehicle_model, vehicle_color, plate_number,
  online_status, verification_status, created_at, updated_at
)
VALUES (
  'driver-dev-01', 'user-driver-dev-01', 'LIC-DEV-001', 'car',
  'Toyota', 'Vios', 'White', 'DEV-0001',
  'offline', 'verified', NOW(), NOW()
)
ON CONFLICT (driver_id) DO NOTHING;
