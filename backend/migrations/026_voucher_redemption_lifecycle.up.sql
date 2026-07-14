-- Voucher & Promotion Hardening, Phase 2 (Voucher Lifecycle): Reserve
-- (booking) -> ConfirmRedeem (trip completed) -> Release (cancelled).
-- status now takes 'reserved' | 'redeemed' | 'released' (was previously
-- only ever written as 'redeemed' | 'released' — additive, no data migration
-- needed since this is a new table with no production rows yet).
-- UNIQUE constraint makes Reserve idempotent: a retried Reserve for the same
-- (voucher, rider, trip) is a no-op rather than a double budget deduction.
ALTER TABLE voucher_redemptions
    ADD CONSTRAINT voucher_redemptions_voucher_rider_trip_key UNIQUE (voucher_id, rider_id, trip_id);
