DROP TABLE IF EXISTS kyc_audit_logs;

DROP INDEX IF EXISTS kyc_documents_expires_at_idx;
DROP INDEX IF EXISTS kyc_documents_driver_type_version_idx;
ALTER TABLE kyc_documents DROP COLUMN IF EXISTS uploaded_by;
ALTER TABLE kyc_documents DROP COLUMN IF EXISTS expires_at;
ALTER TABLE kyc_documents DROP COLUMN IF EXISTS version;
ALTER TABLE kyc_documents ADD CONSTRAINT kyc_documents_driver_id_document_type_key UNIQUE (driver_id, document_type);

DROP INDEX IF EXISTS vehicle_verifications_chassis_number_idx;
DROP INDEX IF EXISTS vehicle_verifications_engine_number_idx;
DROP INDEX IF EXISTS vehicle_verifications_vin_idx;
ALTER TABLE vehicle_verifications DROP COLUMN IF EXISTS chassis_number;
ALTER TABLE vehicle_verifications DROP COLUMN IF EXISTS engine_number;
ALTER TABLE vehicle_verifications DROP COLUMN IF EXISTS vin;

DROP INDEX IF EXISTS driver_verifications_license_number_idx;
DROP INDEX IF EXISTS driver_verifications_national_id_idx;
ALTER TABLE driver_verifications DROP COLUMN IF EXISTS national_id_number;

DROP TABLE IF EXISTS license_capabilities;
