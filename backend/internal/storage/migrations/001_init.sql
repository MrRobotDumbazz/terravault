-- TerraVault initial schema

-- Projects table: mirrors on-chain ProjectState
CREATE TABLE IF NOT EXISTS projects (
    id                   BIGSERIAL PRIMARY KEY,
    on_chain_pubkey      TEXT        NOT NULL UNIQUE,
    developer_wallet     TEXT        NOT NULL,
    state                TEXT        NOT NULL DEFAULT 'Draft',
    project_type         TEXT        NOT NULL DEFAULT 'Residential',
    metadata_uri         TEXT        NOT NULL DEFAULT '',
    fundraise_target     BIGINT      NOT NULL DEFAULT 0,
    fundraise_hard_cap   BIGINT      NOT NULL DEFAULT 0,
    fundraise_deadline   TIMESTAMPTZ NOT NULL,
    escrow_balance       BIGINT      NOT NULL DEFAULT 0,
    total_raised         BIGINT      NOT NULL DEFAULT 0,
    token_price          BIGINT      NOT NULL DEFAULT 0,
    total_tokens         BIGINT      NOT NULL DEFAULT 0,
    tokens_sold          BIGINT      NOT NULL DEFAULT 0,
    milestone_count      INT         NOT NULL DEFAULT 0,
    milestones_completed INT         NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Milestones table: mirrors on-chain MilestoneRecord
CREATE TABLE IF NOT EXISTS milestones (
    id               BIGSERIAL PRIMARY KEY,
    project_id       BIGINT      NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    milestone_index  INT         NOT NULL,
    description      TEXT        NOT NULL DEFAULT '',
    release_bps      INT         NOT NULL DEFAULT 0,
    status           TEXT        NOT NULL DEFAULT 'Pending',
    proof_uri        TEXT        NOT NULL DEFAULT '',
    proof_hash       TEXT        NOT NULL DEFAULT '',
    submitted_at     TIMESTAMPTZ,
    approved_at      TIMESTAMPTZ,
    released_amount  BIGINT      NOT NULL DEFAULT 0,
    dispute_deadline TIMESTAMPTZ,
    UNIQUE (project_id, milestone_index)
);

-- KYC records table
CREATE TABLE IF NOT EXISTS kyc_records (
    id                  BIGSERIAL PRIMARY KEY,
    wallet_address      TEXT        NOT NULL UNIQUE,
    kyc_status          TEXT        NOT NULL DEFAULT 'pending',
    -- 'pending' | 'approved' | 'rejected' | 'expired'
    country             TEXT        NOT NULL DEFAULT '',
    verification_level  INT         NOT NULL DEFAULT 0,
    provider_session_id TEXT        NOT NULL DEFAULT '',
    verified_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Legal and project documents stored on IPFS
CREATE TABLE IF NOT EXISTS documents (
    id          BIGSERIAL PRIMARY KEY,
    project_id  BIGINT      NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ipfs_cid    TEXT        NOT NULL,
    sha256_hash TEXT        NOT NULL,
    doc_type    TEXT        NOT NULL,
    -- 'legal_doc' | 'milestone_proof' | 'audit_report'
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Distribution rounds
CREATE TABLE IF NOT EXISTS distributions (
    id                    BIGSERIAL PRIMARY KEY,
    project_id            BIGINT      NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    round                 INT         NOT NULL,
    total_usdc            BIGINT      NOT NULL DEFAULT 0,
    total_tokens          BIGINT      NOT NULL DEFAULT 0,
    usdc_per_token_scaled NUMERIC(40) NOT NULL DEFAULT 0,
    source                TEXT        NOT NULL DEFAULT 'RentalIncome',
    claim_deadline        TIMESTAMPTZ NOT NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, round)
);

-- Auth tokens / sessions (optional but useful for revocation)
CREATE TABLE IF NOT EXISTS auth_sessions (
    id          BIGSERIAL PRIMARY KEY,
    wallet      TEXT        NOT NULL,
    role        TEXT        NOT NULL DEFAULT 'investor',
    jti         TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_projects_developer ON projects(developer_wallet);
CREATE INDEX IF NOT EXISTS idx_projects_state ON projects(state);
CREATE INDEX IF NOT EXISTS idx_milestones_project ON milestones(project_id);
CREATE INDEX IF NOT EXISTS idx_kyc_wallet ON kyc_records(wallet_address);
CREATE INDEX IF NOT EXISTS idx_distributions_project ON distributions(project_id);
CREATE INDEX IF NOT EXISTS idx_docs_project ON documents(project_id);
