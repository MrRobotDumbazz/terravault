-- TerraVault PostgreSQL schema — Migration 001: initial tables
-- Run via: golang-migrate/migrate

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ── Projects ──────────────────────────────────────────────────────────────────
CREATE TABLE projects (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    solana_pubkey           VARCHAR(44)  UNIQUE NOT NULL,
    project_id              BIGINT       NOT NULL,
    developer_pubkey        VARCHAR(44)  NOT NULL,
    oracle_pubkey           VARCHAR(44),
    token_mint              VARCHAR(44),
    name                    VARCHAR(255) NOT NULL DEFAULT '',
    description             TEXT,
    project_type            VARCHAR(50)  NOT NULL DEFAULT 'Residential',
    location_country        CHAR(2),
    location_city           VARCHAR(100),
    metadata_uri            TEXT,
    legal_doc_hash          VARCHAR(64),
    on_chain_state          VARCHAR(50)  NOT NULL DEFAULT 'Draft',
    total_tokens            BIGINT       NOT NULL DEFAULT 0,
    tokens_sold             BIGINT       NOT NULL DEFAULT 0,
    token_price_usdc        BIGINT       NOT NULL DEFAULT 0,
    fundraise_target_usdc   BIGINT       NOT NULL DEFAULT 0,
    fundraise_hard_cap_usdc BIGINT       NOT NULL DEFAULT 0,
    fundraise_deadline      TIMESTAMPTZ,
    total_raised_usdc       BIGINT       NOT NULL DEFAULT 0,
    milestone_count         SMALLINT     NOT NULL DEFAULT 0,
    milestones_completed    SMALLINT     NOT NULL DEFAULT 0,
    current_milestone_index SMALLINT     NOT NULL DEFAULT 0,
    distribution_round      INT          NOT NULL DEFAULT 0,
    total_distributed_usdc  BIGINT       NOT NULL DEFAULT 0,
    paused                  BOOLEAN      NOT NULL DEFAULT FALSE,
    kyc_required            BOOLEAN      NOT NULL DEFAULT FALSE,
    transfer_fee_bps        SMALLINT     NOT NULL DEFAULT 0,
    last_synced_slot        BIGINT       NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_developer ON projects(developer_pubkey);
CREATE INDEX idx_projects_state     ON projects(on_chain_state);
CREATE INDEX idx_projects_type      ON projects(project_type);

-- ── Milestones ────────────────────────────────────────────────────────────────
CREATE TABLE milestones (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_pubkey       VARCHAR(44)  NOT NULL REFERENCES projects(solana_pubkey),
    milestone_index      SMALLINT     NOT NULL,
    milestone_type       VARCHAR(50)  NOT NULL DEFAULT 'Custom',
    description          TEXT,
    release_bps          SMALLINT     NOT NULL DEFAULT 0,
    status               VARCHAR(50)  NOT NULL DEFAULT 'Pending',
    proof_uri            TEXT,
    proof_hash           VARCHAR(64),
    submitted_at         TIMESTAMPTZ,
    approved_at          TIMESTAMPTZ,
    released_amount_usdc BIGINT       NOT NULL DEFAULT 0,
    dispute_deadline     TIMESTAMPTZ,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (project_pubkey, milestone_index)
);

-- ── KYC Records ───────────────────────────────────────────────────────────────
CREATE TABLE kyc_records (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_pubkey       VARCHAR(44)  UNIQUE NOT NULL,
    provider_session_id VARCHAR(255),
    status              VARCHAR(50)  NOT NULL DEFAULT 'pending',
    country_code        CHAR(2),
    verification_level  VARCHAR(50)  NOT NULL DEFAULT 'basic',
    verified_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ── Documents ─────────────────────────────────────────────────────────────────
CREATE TABLE documents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_pubkey  VARCHAR(44) REFERENCES projects(solana_pubkey),
    milestone_index SMALLINT,
    doc_type        VARCHAR(50) NOT NULL DEFAULT 'other',
    ipfs_cid        VARCHAR(100),
    sha256_hash     VARCHAR(64),
    uploaded_by     VARCHAR(44),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Distribution pools ────────────────────────────────────────────────────────
CREATE TABLE distribution_pools (
    id                        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_pubkey            VARCHAR(44) NOT NULL REFERENCES projects(solana_pubkey),
    round                     INT         NOT NULL,
    total_usdc_deposited      BIGINT      NOT NULL DEFAULT 0,
    total_tokens_at_snapshot  BIGINT      NOT NULL DEFAULT 0,
    source                    VARCHAR(50) NOT NULL DEFAULT 'RentalIncome',
    deposited_at              TIMESTAMPTZ,
    total_claimed             BIGINT      NOT NULL DEFAULT 0,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_pubkey, round)
);
