# TerraVault Oracle & API (Go)

## Overview

The Go backend serves three roles:

| Role | Description |
|---|---|
| **Oracle** | Signs and submits on-chain milestone proof transactions |
| **API** | REST endpoints for project data, KYC, and frontend aggregation |
| **Listener** | WebSocket subscriber syncing on-chain events to PostgreSQL |

## Structure

```
oracle/
├── cmd/
│   ├── api/main.go          # HTTP API server
│   └── listener/main.go     # On-chain event listener
├── internal/
│   ├── anchor/              # IDL bindings + instruction builders
│   ├── oracle/              # Keypair management + tx submission
│   ├── kyc/                 # KYC provider integration (Persona)
│   ├── listener/            # Solana WebSocket + log parser
│   ├── storage/
│   │   ├── postgres/        # DB repositories + migrations
│   │   └── ipfs/            # IPFS upload client
│   └── api/                 # Router, middleware, handlers
└── pkg/
    └── solana/              # RPC client wrapper with retry
```

## API Endpoints

### Public
- `GET /v1/projects` — list with filters (type, status, page)
- `GET /v1/projects/:id` — full project detail
- `GET /v1/projects/:id/milestones` — milestone list
- `GET /v1/projects/:id/distributions` — distribution history

### Authenticated (investor)
- `GET /v1/me/positions` — investor portfolio
- `GET /v1/me/kyc-status`
- `POST /v1/me/kyc/initiate` — returns KYC session URL
- `POST /v1/me/kyc/webhook` — KYC completion webhook

### Authenticated (developer)
- `POST /v1/projects` — create project metadata
- `PUT /v1/projects/:id`
- `POST /v1/projects/:id/milestones/:index/proof` — submit milestone proof → triggers oracle

### Internal (API-key protected)
- `POST /internal/oracle/submit-milestone`
- `POST /internal/oracle/release-milestone`
- `GET  /internal/oracle/pending-milestones`

### Admin
- `GET  /admin/disputes`
- `POST /admin/disputes/:id/resolve`
- `POST /admin/projects/:id/pause`

## Setup

```bash
cp .env.example .env
# Edit .env with your database URL, Solana RPC, oracle keypair path

# Run database migrations
go run ./cmd/migrate up

# Start API server
go run ./cmd/api

# Start event listener (separate process)
go run ./cmd/listener
```

## Oracle Signing Flow

1. Inspector submits proof docs → `POST /v1/projects/:id/milestones/:index/proof`
2. API uploads docs to IPFS, gets CID + SHA-256 hash
3. Oracle signs `(project_pubkey || milestone_index || proof_hash)` with oracle keypair
4. Builds two-instruction transaction (Ed25519 + `submit_milestone_proof`)
5. Submits with exponential backoff on `BlockhashNotFound`
6. On confirmation at `Confirmed` level → updates PostgreSQL

## Phase 4 Implementation Checklist

- [ ] Anchor IDL parsing and instruction builders
- [ ] Oracle keypair management (env var for dev, HSM/Vault for prod)
- [ ] PostgreSQL schema migrations (see `internal/storage/postgres/migrations/`)
- [ ] All REST handlers wired up
- [ ] WebSocket logsSubscribe + Anchor event discriminator parsing
- [ ] KYC integration (Persona webhook)
- [ ] IPFS upload client
- [ ] Retry logic with exponential backoff
- [ ] Rate limiting, CORS, JWT middleware
- [ ] Docker Compose for local dev
