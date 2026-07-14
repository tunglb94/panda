ALTER TABLE rider_verifications DROP COLUMN IF EXISTS vision_result;
ALTER TABLE rider_verifications DROP COLUMN IF EXISTS ocr_result;
ALTER TABLE rider_verifications DROP COLUMN IF EXISTS ai_confidence;
ALTER TABLE rider_verifications DROP COLUMN IF EXISTS review_mode;
