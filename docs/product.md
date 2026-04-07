# Product

## What is TerraVault?

TerraVault is a decentralized real estate investment protocol on Solana. It tokenizes properties into fractional SPL Token-2022 shares, locks investor funds in milestone-based escrow, verifies construction progress through an on-chain oracle, and distributes rental income and sale proceeds proportionally — all without intermediaries.

---

## Target Users

| User | Need |
|------|------|
| **Retail investors** | Access real estate with small capital ($10+), earn rental yield, trade on secondary markets |
| **Institutional investors** | Compliant (KYC-gated), auditable on-chain positions with transparent income history |
| **Real estate developers** | Raise construction capital without bank loans; milestone-locked funding incentivizes delivery |
| **Property managers** | Deposit rental income on-chain; automated proportional distribution removes manual overhead |

---

## Core Value Propositions

1. **Fractional ownership from $10** — 1,000,000 tokens per project at $0.01–$10 each; any wallet can invest
2. **Trustless escrow** — USDC locked in PDA vault; developer gets paid only after oracle-verified construction milestones pass dispute windows
3. **Transparent income distribution** — Rental yield and sale proceeds deposited on-chain; investors claim USDC proportional to token holdings
4. **Investor protection via DAO disputes** — 72-hour dispute window with IPFS evidence; admin multisig can pay investors, force-close, or extend deadlines
5. **Developer accountability** — Blacklist mechanism permanently blocks fraud actors from creating future projects

---

## How It Compares

| Feature | Traditional RE | REITs | RealT / Centrifuge | TerraVault |
|---------|---------------|-------|---------------------|------------|
| Min investment | $50,000+ | $100 | $50 | $10 |
| Liquidity | Years | Days | Days | Minutes (DEX) |
| Construction escrow | No | No | No | Yes (milestone-based) |
| Income distribution | Quarterly, opaque | Quarterly | On-chain | On-chain, instant |
| Dispute mechanism | Courts | None | None | On-chain DAO |
| Developer blacklist | No | No | No | Yes |
| Solana speed | — | — | Ethereum | Solana (400ms) |

---

## Project Types

| Type | Examples | Typical Milestone Count |
|------|---------|------------------------|
| `Residential` | Apartment buildings, condos, housing developments | 4–6 |
| `Commercial` | Office towers, retail centers, warehouses | 5–8 |
| `Agricultural` | Farmland, greenhouse facilities, agri-processing | 3–5 |
| `Mixed` | Mixed-use developments | 6–10 |

---

## Milestone Types

Construction progress is broken into verifiable stages that align with real-world inspection points:

1. `SitePreparation` — Clearing, grading, utilities
2. `Foundation` — Concrete pour and structural inspection
3. `Framing` — Structural framework completion
4. `Roofing` — Weatherproofing and roofing system
5. `MEP` — Mechanical, Electrical, Plumbing rough-in
6. `InteriorWork` — Finishes, fixtures, flooring
7. `Landscaping` — Exterior and common areas
8. `Completion` — Final inspection and certificate of occupancy
9. `Custom` — Developer-defined stages for specialized projects

---

## Revenue Model

| Source | Who Pays | Rate |
|--------|----------|------|
| Transfer fee | Investors (on secondary trades) | 0–5% (configurable, max 5%) |
| Protocol fee on distributions | Developer | TBD (governance) |
| Oracle submission fee | Developer | Gas + oracle service fee |

---

## KYC & Compliance

TerraVault supports KYC-gated token transfers through the Token-2022 TransferHook:
- Developers can set `kyc_required = true` per project
- KYC is verified off-chain via Persona and enforced on every token transfer
- Non-verified wallets cannot receive KYC-required project tokens
- KYC status is managed by the Go oracle API

---

## Token Economics (per project)

```
Total tokens issued: configurable (e.g. 1,000,000)
Token price (USDC):  configurable (e.g. $10.00)
Hard cap:            configurable
Soft cap (target):   configurable (must be ≤ hard cap)
Fundraise duration:  ≥ 7 days (enforced on-chain)
Transfer fee:        0–5% (set by developer, capped by protocol)
```

Tokens represent proportional ownership. Income distribution is calculated as:

```
investor_share = (investor_tokens / total_circulating_tokens) × pool_usdc
```

Using 10¹² fixed-point scaling to avoid rounding errors at any token supply level.
