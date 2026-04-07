# TerraVault — Real Estate Tokenization Protocol

[![CI](https://github.com/your-org/terravault/actions/workflows/ci.yml/badge.svg)](https://github.com/your-org/terravault/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-14F195.svg)](LICENSE)
[![Solana](https://img.shields.io/badge/Solana-devnet-9945FF)](https://solana.com)
[![Hackathon](https://img.shields.io/badge/Colosseum-2026-14F195)](https://colosseum.org)

> Decentralized real estate investment protocol on Solana — fractional property ownership with milestone-based escrow, oracle-verified construction proofs, and on-chain income distribution.

[Live Demo](#) · [Video Walkthrough](#) · [Docs](docs/) · [Colosseum Submission](#)

---

![TerraVault Dashboard](assets/project.jpg)

---

## Submission to 2026 Solana National Hackathon

| Name | Role | Contact |
|------|------|---------|
| Abay Mukhammetali | Founder & Lead Engineer | [Telegram](https://t.me/tallyhallfan#) |
| Diana Kalieva | CTO & Backend developer | [Telegram](https://t.me/brewmountain#) |

---

## Problem and Solution

### 1. Illiquid Real Estate Markets
- **Problem:** Traditional real estate requires large capital, is geographically restricted, and is virtually impossible to exit quickly.
- **TerraVault:** Tokenizes properties into fractional shares (SPL Token-2022), allowing anyone to invest from $10 and trade on secondary markets.

### 2. Developer Trust & Construction Risk
- **Problem:** Investors have no on-chain guarantee that construction progresses as promised; developers can take funds and disappear.
- **TerraVault:** Milestone-based escrow with oracle-verified proofs. Funds are released in tranches only after each construction stage is verified and a 48-hour dispute window passes.

### 3. Opaque Income Distribution
- **Problem:** Rental income and sale proceeds in traditional real estate are distributed through opaque intermediaries with high fees.
- **TerraVault:** Developer deposits USDC income on-chain; token holders claim proportional shares directly from the protocol with zero intermediary.

### 4. No Dispute Recourse
- **Problem:** When developer fraud occurs, investors have no on-chain mechanism to recover funds or flag bad actors.
- **TerraVault:** DAO-style dispute system — investors raise disputes, submit IPFS evidence, admin multisig resolves with decisions: PayInvestors, RefundAndExtend, or ForceClose.

---

## Why Solana

- **Speed** — 400 ms block time enables real-time investor dashboards and instant token transfers
- **Cost** — $0.00025 per transaction makes micro-distributions and frequent milestone submissions economically viable
- **Token-2022** — Native transfer hooks allow TerraVault to track investor positions and enforce KYC/transfer restrictions at the protocol level
- **Composability** — Anchor's CPI integrates with Jupiter (secondary market liquidity) and Pyth (real estate price feeds) without forks

---

## Summary of Features

- Fractional real estate token issuance (SPL Token-2022 with TransferHook)
- USDC escrow vault with milestone-based release schedule
- Oracle-submitted construction proofs with 48-hour dispute window
- On-chain income distribution (rental yield + sale proceeds)
- DAO dispute resolution: raise dispute → submit IPFS evidence → admin multisig decision
- Developer blacklist — bad actor flagging prevents future project creation
- Token account freeze after PayInvestors decision
- 48-hour oracle key rotation timelock
- Emergency pause (circuit breaker)
- KYC-gated token transfers

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| On-chain programs | Rust · Anchor Framework 0.32 |
| Token standard | SPL Token-2022 · TransferHook |
| SDK / Client | TypeScript · @solana/web3.js |
| Frontend | React · Vite · TailwindCSS |
| Backend / Oracle | Go · Fastify-style REST |
| Database | PostgreSQL |
| File storage | IPFS |
| KYC | Persona |
| Testing | Anchor Tests · Bankrun |

---

## Architecture

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
                    │  ┌─────────▼──────────┐  │────▶│  Distribution    │
                    │  │ Income Distributor │  │     │  Pool (USDC)     │
                    │  └────────────────────┘  │     └──────────────────┘
                    └──────────────────────────┘
```

See [docs/architecture.md](docs/architecture.md) for full component breakdown.

---

## Quick Start

**Prerequisites:** Node.js 18+, Rust, Anchor CLI 0.32, Solana CLI, Go 1.21+

```bash
# Clone the repository
git clone https://github.com/your-org/terravault
cd terravault

# Install JS dependencies
yarn install

# Copy environment variables
cp .env.example .env

# Build Solana programs
anchor build

# Run tests
anchor test

# Start frontend (Vite)
cd app && npm run dev

# Start Go oracle (separate terminal)
cd backend && go run ./cmd/api
```

---

## Project Lifecycle

```
Draft → Fundraising → Active → InMilestones → Completed → Distributing
                  ↓                     ↓
              Cancelled           DisputeActive → Resolved
```

| Status | Description |
|--------|-------------|
| `Draft` | Milestones being defined, tokens not yet issued |
| `Fundraising` | Investors buying fractional tokens with USDC |
| `Active` | Soft cap reached; tokens transferable |
| `InMilestones` | Construction in progress; oracle submitting proofs |
| `Completed` | All milestones released |
| `Distributing` | Rental income / sale proceeds being distributed |
| `DisputeActive` | 72-hour DAO dispute window open |
| `Resolved` | Admin multisig decision executed |
| `Cancelled` | Fundraise cancelled; investors can claim refunds |

---

## Roadmap

- [x] Core tokenization protocol (SPL Token-2022)
- [x] Milestone-based escrow and proof submission
- [x] On-chain income distribution
- [x] DAO dispute resolution + developer blacklist
- [ ] Mainnet deployment
- [ ] App-defined auction logic SDK
- [ ] ZK-proofed property valuations (Pyth integration)
- [ ] Secondary market DEX integration (Jupiter)
- [ ] Governance token + DAO treasury

Full roadmap: [docs/roadmap.md](docs/roadmap.md)

---

## Resources

- [Project Presentation](#)
- [Video Demo](#)
- [Live Application](#)
- [X / Twitter](#)

---

## License

MIT — see [LICENSE](LICENSE)
