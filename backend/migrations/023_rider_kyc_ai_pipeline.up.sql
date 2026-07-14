-- Rider KYC AI Pipeline foundation: Upload -> Rule Engine -> OCR -> Vision
-- -> Decision. review_mode records the pipeline's own recommendation
-- (auto_approved/manual_review/rejected) as an audit trail, independent of
-- `status` (which an admin's Approve/Reject can still override — see
-- user/domain/entity/rider_verification.go).
ALTER TABLE rider_verifications ADD COLUMN IF NOT EXISTS review_mode TEXT NOT NULL DEFAULT '';
ALTER TABLE rider_verifications ADD COLUMN IF NOT EXISTS ai_confidence DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE rider_verifications ADD COLUMN IF NOT EXISTS ocr_result TEXT NOT NULL DEFAULT '{}';
ALTER TABLE rider_verifications ADD COLUMN IF NOT EXISTS vision_result TEXT NOT NULL DEFAULT '{}';
