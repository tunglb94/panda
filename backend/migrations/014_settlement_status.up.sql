-- Settlement Status state machine (Fix #7) — guards against a crash between
-- "claimed this trip_id" and "finished posting every ledger entry". Existing
-- rows predate this fix and are already fully posted, hence the 'posted'
-- default.
ALTER TABLE settlements ADD COLUMN status VARCHAR(16) NOT NULL DEFAULT 'posted';
