-- Settlement Engine + Driver Finance (Financial Core). Additive only — no
-- existing table touched. Ledger is the source of truth: wallets never
-- store a balance column; every number the app shows is derived from
-- wallet_ledger_entries at read time (see wallet/app/get_wallet_summary.go).

-- ─── Wallet (Phần 3 — a projection, not a store of money) ──────────────────
CREATE TABLE IF NOT EXISTS wallets (
    wallet_id   TEXT        PRIMARY KEY,
    owner_id    TEXT        NOT NULL UNIQUE,
    wallet_type TEXT        NOT NULL, -- 'rider' | 'driver' | 'platform'
    currency    TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS wallets_type_idx ON wallets(wallet_type);

-- ─── Transaction (Phần 1 — one atomic financial event, immutable) ──────────
CREATE TABLE IF NOT EXISTS wallet_transactions (
    transaction_id TEXT        PRIMARY KEY,
    type           TEXT        NOT NULL, -- ride_income | delivery_income | commission | ...
    reference_id   TEXT        NOT NULL DEFAULT '', -- trip_id, payout_request_id, etc.
    payment_method TEXT        NOT NULL DEFAULT '', -- 'cash' | 'wallet' | '' (non-trip)
    currency       TEXT        NOT NULL,
    description    TEXT        NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS wallet_transactions_reference_idx ON wallet_transactions(reference_id);
CREATE INDEX IF NOT EXISTS wallet_transactions_type_idx ON wallet_transactions(type);

-- ─── Ledger Entry (Phần 1 — immutable, never UPDATE/DELETE) ────────────────
CREATE TABLE IF NOT EXISTS wallet_ledger_entries (
    entry_id       TEXT        PRIMARY KEY,
    wallet_id      TEXT        NOT NULL,
    transaction_id TEXT        NOT NULL,
    direction      TEXT        NOT NULL, -- 'credit' | 'debit'
    amount_cents   BIGINT      NOT NULL,
    currency       TEXT        NOT NULL,
    description    TEXT        NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS wallet_ledger_entries_wallet_idx ON wallet_ledger_entries(wallet_id, created_at);
CREATE INDEX IF NOT EXISTS wallet_ledger_entries_transaction_idx ON wallet_ledger_entries(transaction_id);

-- ─── Settlement (Phần 2/13 — the audit record of one trip's financial
-- resolution; immutable; UNIQUE(trip_id) makes the Settlement Engine
-- idempotent — a retried payment callback can never double-settle) ─────────
CREATE TABLE IF NOT EXISTS settlements (
    settlement_id        TEXT        PRIMARY KEY,
    trip_id               TEXT        NOT NULL UNIQUE,
    driver_id              TEXT        NOT NULL,
    trip_type              TEXT        NOT NULL, -- 'ride' | 'delivery'
    payment_method          TEXT        NOT NULL, -- 'cash' | 'wallet'
    fare_amount_cents        BIGINT      NOT NULL,
    commission_rate           DOUBLE PRECISION NOT NULL,
    commission_amount_cents    BIGINT      NOT NULL,
    driver_income_cents         BIGINT      NOT NULL,
    promotion_subsidy_cents      BIGINT      NOT NULL DEFAULT 0,
    voucher_cost_cents            BIGINT      NOT NULL DEFAULT 0,
    currency                       TEXT        NOT NULL,
    transaction_id                  TEXT        NOT NULL,
    created_at                       TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS settlements_driver_idx ON settlements(driver_id, created_at DESC);

-- ─── Bank Account (Phần 6 — exactly one default account per driver) ────────
CREATE TABLE IF NOT EXISTS bank_accounts (
    bank_account_id      TEXT        PRIMARY KEY,
    driver_id             TEXT        NOT NULL UNIQUE,
    bank_name              TEXT        NOT NULL,
    account_holder_name     TEXT        NOT NULL,
    account_number            TEXT        NOT NULL,
    branch_name                TEXT        NOT NULL DEFAULT '',
    created_at                  TIMESTAMPTZ NOT NULL,
    updated_at                   TIMESTAMPTZ NOT NULL
);

-- ─── Payout Request (Phần 5/8 — Pending -> Approved -> Paid, or Rejected) ──
CREATE TABLE IF NOT EXISTS payout_requests (
    payout_request_id TEXT        PRIMARY KEY,
    driver_id          TEXT        NOT NULL,
    amount_cents        BIGINT      NOT NULL,
    currency              TEXT        NOT NULL,
    bank_account_id        TEXT        NOT NULL,
    bank_name                TEXT        NOT NULL DEFAULT '',
    account_number_masked      TEXT        NOT NULL DEFAULT '',
    status                      TEXT        NOT NULL DEFAULT 'pending',
    requested_at                 TIMESTAMPTZ NOT NULL,
    reviewed_at                   TIMESTAMPTZ,
    reviewed_by                    TEXT        NOT NULL DEFAULT '',
    reject_reason                    TEXT        NOT NULL DEFAULT '',
    paid_at                           TIMESTAMPTZ,
    transaction_id                     TEXT        NOT NULL DEFAULT '',
    created_at                          TIMESTAMPTZ NOT NULL,
    updated_at                           TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS payout_requests_driver_idx ON payout_requests(driver_id, created_at DESC);
CREATE INDEX IF NOT EXISTS payout_requests_status_idx ON payout_requests(status);

-- ─── Audit Log (Phần 12 — "Adjustment phải có Audit Log"; append-only) ─────
CREATE TABLE IF NOT EXISTS wallet_audit_logs (
    id          TEXT        PRIMARY KEY,
    entity_type TEXT        NOT NULL, -- 'settlement' | 'payout_request' | 'bank_account' | 'ledger_entry'
    entity_id   TEXT        NOT NULL,
    driver_id   TEXT        NOT NULL,
    action      TEXT        NOT NULL, -- 'create' | 'approve' | 'reject' | 'paid' | 'manual_adjustment'
    actor_id    TEXT        NOT NULL, -- driver_id, admin user id, or 'system'
    old_value   TEXT        NOT NULL DEFAULT '',
    new_value   TEXT        NOT NULL DEFAULT '',
    reason      TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS wallet_audit_logs_driver_idx ON wallet_audit_logs(driver_id, created_at DESC);
