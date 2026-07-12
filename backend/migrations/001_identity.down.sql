-- Rollback for 001_identity.up.sql — drops only objects created by that migration,
-- in reverse dependency order (child tables with FKs first).

DROP TABLE IF EXISTS identity_users;
DROP TABLE IF EXISTS identity_role_permissions;
DROP TABLE IF EXISTS identity_roles;
DROP TABLE IF EXISTS identity_permissions;
