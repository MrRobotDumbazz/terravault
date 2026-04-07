# Architecture

## System Overview

TerraVault sits between real estate developers, investors, and the Solana ledger. It replaces the traditional intermediary (broker, escrow agent, property manager) with a set of verifiable on-chain programs backed by a Go oracle that signs construction milestone proofs.

```
┌─────────────┐     ┌──────────────────────────┐     ┌──────────────────┐
│  Investor / │────▶│     TerraVault Protocol  │────▶│  Solana Ledger   │
│  Developer  │     │  ┌────────────────────┐  │     │  (Token-2022)    │
└─────────────┘     │  │  Escrow Vault      │  │     └──────────────────┘
                    │  └─────────┬──────────┘  │
                    │  ┌─────────▼──────────┐  │     ┌──────────────────┐
                    │  │ Milestone Engine   │  │◀────│   Go Oracle      │
                    │  └─────────┬──────────┘  │     │  (Proof Signer)  │
                    │  ┌─────────▼──────────┐  │     └──────────────────┘
                    │  │  Dispute / DAO     │  │
                    │  └─────────┬──────────┘  │     ┌──────────────────┐
                    │  ┌─────────▼──────────┐  │────▶│ Distribution     │
                    │  │ Income Distributor │  │     │ Pool (USDC)      │
                    │  └────────────────────┘  │     └──────────────────┘
                    └──────────────────────────┘
```

---

## On-Chain Components

### ProjectState PDA
Central state account for each real estate project. Tracks fundraising progress, milestone completion, distribution rounds, dispute status, and token configuration.

Key fields:
- `state: ProjectStatus` — lifecycle state machine (Draft → Fundraising → Active → InMilestones → Completed)
- `project_type: ProjectType` — Residential / Commercial / Agricultural / Mixed
- `escrow_vault` — USDC token account controlled by the ProjectState PDA
- `milestone_bps_total` — must equal 10,000 bps before fundraising can start
- `dispute_deadline` — unix timestamp when DisputeActive window closes (72h)

### MilestoneRecord PDA
One PDA per construction milestone. Stores proof URI (IPFS), proof hash, oracle signature, and dispute window deadline.

Milestone types: `SitePreparation`, `Foundation`, `Framing`, `Roofing`, `MEP`, `InteriorWork`, `Landscaping`, `Completion`, `Custom`

### InvestorPosition PDA
Tracks each investor's token balance and last claimed distribution round. Enforces sequential distribution claiming to prevent skip-and-claim exploits.

### DeveloperProfile PDA
Stores developer reputation data including `is_blacklisted` flag. Blacklisted developers cannot initialize new projects.

### DistributionPool PDA
Created each time a developer deposits rental income or sale proceeds. Captures a snapshot of circulating token supply for proportional USDC distribution.

### TokenConfig PDA
Stores Token-2022 mint configuration including transfer fee basis points and KYC requirement flag.

---

## Off-Chain Components

### Go Oracle (`backend/`)
Three services in one binary:

| Role | Description |
|------|-------------|
| **Oracle** | Signs and submits `submit_milestone_proof` transactions after off-chain construction verification |
| **API** | REST endpoints for project data, investor positions, KYC status, and frontend aggregation |
| **Listener** | WebSocket subscriber that syncs on-chain events to PostgreSQL for efficient querying |

Internal packages:
- `internal/anchor` — IDL bindings and instruction builders
- `internal/oracle` — Keypair management and transaction submission with retry
- `internal/kyc` — Persona KYC provider integration
- `internal/listener` — Solana WebSocket log parser
- `internal/storage/postgres` — DB repositories and migrations
- `internal/storage/ipfs` — IPFS upload client for evidence hashes
- `internal/api` — Router, middleware, handlers

### Frontend (`frontend/`)
React + Vite + TailwindCSS single-page app. Connects via `@solana/web3.js` and queries the Go API for aggregated project data.

---

## Token-2022 TransferHook

Every transfer of a TerraVault project token triggers `execute_transfer_hook`, which:
1. Updates `InvestorPosition.tokens_held` for both sender and receiver
2. Enforces KYC checks if `kyc_required = true`
3. Validates transfer fee collection

This ensures distribution snapshots are always accurate without requiring investors to manually register.

---

## Milestone Lifecycle

```
add_milestone (up to 10, sum of release_bps = 10,000)
       ↓
start_fundraising
       ↓
buy_tokens (USDC → escrow vault, tokens → investor)
       ↓
activate_project (soft cap reached)
       ↓
submit_milestone_proof (oracle, opens 48h dispute window)
       ↓
  [dispute_milestone] ──→ resolve_dispute (arbitrator)
       ↓
release_milestone_funds (oracle, after dispute window)
       ↓
 (repeat for each milestone)
       ↓
deposit_income + claim_distribution
```

---

## Security Model

| Mechanism | Description |
|-----------|-------------|
| Milestone escrow | USDC locked in PDA-controlled vault; developer cannot withdraw directly |
| 48h dispute window | Token holders can challenge milestone proofs before funds release |
| 72h DAO dispute | Project-level dispute with IPFS evidence submission |
| Developer blacklist | Flagged developers cannot create new projects |
| Token-2022 freeze | After `PayInvestors` decision, developer token account is frozen |
| Oracle timelock | Oracle key rotation requires 48h timelock |
| Emergency pause | Authority can pause all state mutations |
| Transfer fee cap | `MAX_TRANSFER_FEE_BPS = 500` (5%) enforced in constants |

---

## Constants

| Constant | Value | Purpose |
|----------|-------|---------|
| `DISPUTE_WINDOW_SECONDS` | 172,800 (48h) | Milestone-level dispute window |
| `DAO_DISPUTE_DEADLINE_SECONDS` | 259,200 (72h) | Project-level DAO dispute window |
| `MIN_FUNDRAISE_DEADLINE_SECONDS` | 604,800 (7d) | Minimum fundraise duration |
| `MAX_MILESTONES` | 10 | Maximum milestones per project |
| `MAX_TRANSFER_FEE_BPS` | 500 | Maximum transfer fee (5%) |
| `ORACLE_UPDATE_TIMELOCK_SECONDS` | 172,800 (48h) | Oracle key rotation delay |
| `DISTRIBUTION_SCALE` | 10¹² | Fixed-point precision for distributions |
