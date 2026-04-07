# TerraVault — Detailed Project Description

---

## Overview

TerraVault is a decentralized real estate investment protocol built on Solana. It transforms physical
properties into on-chain fractional ownership tokens, locks construction funding inside a
milestone-gated escrow vault, verifies build progress through a cryptographic oracle, and
distributes rental income and sale proceeds directly to token holders — all without a single
intermediary touching the money.

The protocol is written in Rust using the Anchor framework (v0.32), uses SPL Token-2022 with
a custom TransferHook for position tracking and KYC enforcement, and is backed by a Go oracle
service that bridges off-chain construction verification to on-chain state transitions.

TerraVault is submitted to the 2026 Colosseum Solana Hackathon.

---

## The Problem Space

Real estate is the largest asset class in the world, estimated at over $300 trillion globally.
Yet it remains one of the least accessible and least liquid investment categories available to
ordinary people. The barriers are structural and have persisted for decades.

### Barrier 1: Capital Requirements

Buying a single-family home in most metropolitan markets requires $50,000–$500,000 in upfront
capital. Commercial real estate starts at millions. These thresholds lock out the vast majority
of the global population from an asset class that has historically provided some of the most
consistent inflation-adjusted returns available.

Real Estate Investment Trusts (REITs) lowered the entry point to a few hundred dollars, but
they trade at the fund level — investors own shares in a portfolio, not fractional interests in
specific properties. You cannot choose to invest in the apartment complex in your neighborhood
but not the shopping mall in another city.

TerraVault allows any wallet to invest in a specific, identified real estate project for as
little as ten dollars. Tokens represent fractional ownership of that single property, not a
pooled basket.

### Barrier 2: Construction Risk and Developer Trust

Real estate development is notorious for fraud and mismanagement. Developers raise capital from
investors, start construction, and then — due to mismanagement, fraud, or market conditions —
fail to deliver. In traditional finance, investors have little recourse beyond protracted legal
proceedings that can take years and cost more than the original investment.

The fundamental problem is that capital is released upfront, giving developers full control
before any value has been delivered. There is no mechanism that automatically holds the developer
accountable to a delivery schedule without requiring constant human oversight.

TerraVault inverts this relationship entirely. Investor USDC is locked inside a PDA-controlled
escrow vault at the moment of investment. The developer cannot access a single dollar until a
trusted oracle confirms that a specific construction milestone has been completed and a 48-hour
dispute window has elapsed without a successful challenge. Funds flow to the developer in
tranches, strictly proportional to work completed.

### Barrier 3: Opaque Income Distribution

When a rental property generates income, it flows through a chain of intermediaries: property
manager, accountant, bank, and finally the investor — each taking a cut and adding weeks of
latency. Investors typically receive quarterly statements that are difficult to verify
independently. The entire process is opaque by design.

Sale proceeds are even worse. When a property sells, the distribution calculation is done
manually by a fund administrator, is subject to disputes, and can take months to reach
investors.

TerraVault replaces this entire chain with three on-chain transactions. The developer calls
`deposit_income` with the USDC amount and source type (rental or sale). The protocol creates a
DistributionPool with a snapshot of circulating token supply at that exact moment. Each token
holder then calls `claim_distribution` and receives their exact proportional share, calculated
with 10¹² fixed-point precision, directly into their wallet.

### Barrier 4: No Dispute Recourse

When a developer commits fraud — takes investor money and disappears, or claims construction is
complete when it is not — investors in traditional markets have almost no fast-acting options.
Filing lawsuits takes years. Regulatory complaints take months. By the time any action is taken,
the developer's assets are often gone.

Crypto has historically made this worse, not better. Scam projects can raise millions through
token sales and vanish overnight, with investors having no on-chain mechanism to recover funds.

TerraVault introduces a full dispute resolution system directly on-chain. Any investor or
stakeholder can raise a project-level dispute. This immediately transitions the project to a
`DisputeActive` state and opens a 72-hour window during which both parties can submit IPFS
evidence hashes — photos, inspection reports, legal documents, correspondence — as permanent,
tamper-proof on-chain records. An admin multisig then reviews the evidence and executes one of
three binding decisions:

- **PayInvestors** — Freeze the developer's token account and distribute the escrow balance to
  token holders proportionally.
- **RefundAndExtend** — Return all escrowed USDC to investors and set a new fundraising deadline,
  giving the developer a chance to restart under new terms.
- **ForceClose** — Split the escrow proportionally between all parties and permanently close the
  project.

Developers who are found to have acted fraudulently are flagged via `flag_developer`, which sets
`is_blacklisted = true` on their on-chain profile. Blacklisted developers cannot initialize new
projects on the protocol, ever.

---

## The TerraVault Solution

TerraVault is not a wrapper around existing DeFi primitives. It is a purpose-built protocol
that addresses all four problems described above through a coherent set of on-chain mechanics.

### Fractional Tokenization via SPL Token-2022

Each real estate project on TerraVault issues a unique SPL Token-2022 mint. The developer
configures the total token supply, the price per token in USDC, a fundraising target (soft cap),
a hard cap, and a deadline. Tokens represent fractional ownership shares in the specific
underlying property.

TerraVault uses Token-2022 specifically because of its native extension support. The protocol
implements a custom TransferHook that executes automatically on every token transfer. This hook:

1. Updates the sender's `InvestorPosition.tokens_held` downward.
2. Updates the receiver's `InvestorPosition.tokens_held` upward.
3. Validates KYC status if the project requires it.
4. Enforces transfer fee collection.

This means the protocol always has an accurate, real-time record of every investor's position,
without requiring investors to manually register holdings. Distribution snapshots are accurate
the instant they are taken.

### Milestone-Based Escrow

Before fundraising can begin, the developer must define a complete milestone schedule. Each
milestone specifies:

- A type (SitePreparation, Foundation, Framing, Roofing, MEP, InteriorWork, Landscaping,
  Completion, or Custom).
- A description (up to 64 bytes).
- A `release_bps` — the percentage of the total escrow this milestone unlocks, in basis points.

The sum of all `release_bps` across all milestones must equal exactly 10,000 (100%) before the
project can transition from Draft to Fundraising. This is enforced at the protocol level. A
developer cannot start raising money without committing to a complete, fully-allocated milestone
schedule.

Up to 10 milestones are supported per project. For a typical residential development, milestones
might be allocated as follows:

| Milestone | Type | Release |
|-----------|------|---------|
| 0 | SitePreparation | 5% |
| 1 | Foundation | 20% |
| 2 | Framing | 20% |
| 3 | Roofing | 15% |
| 4 | MEP | 15% |
| 5 | InteriorWork | 15% |
| 6 | Landscaping | 5% |
| 7 | Completion | 5% |

When a milestone is ready to be verified, the Go oracle inspects the off-chain evidence
(photos, third-party inspection reports, municipal permits) and calls `submit_milestone_proof`,
providing an IPFS URI pointing to the evidence package and a SHA-256 hash of the proof document.
This opens a 48-hour dispute window.

During the dispute window, any token holder can call `dispute_milestone` with a 128-byte reason
string. If a dispute is raised, an arbitration authority reviews the evidence and calls
`resolve_dispute`. If approved, the milestone proceeds to fund release. If rejected, the oracle
must resubmit with corrected evidence.

If no dispute is raised within 48 hours, the oracle calls `release_milestone_funds`. The
protocol calculates the USDC amount corresponding to the milestone's `release_bps`, transfers it
from the escrow vault to the developer's USDC account, and marks the milestone as Released.
When the final milestone is released, the project transitions to Completed.

### Income Distribution Engine

The distribution engine is designed to handle any number of investors, any token supply, and
any income amount with zero rounding errors.

When the developer calls `deposit_income`, the protocol:
1. Creates a new `DistributionPool` PDA for the current distribution round.
2. Records the total USDC amount deposited.
3. Takes a snapshot of the total circulating token supply at that exact slot.
4. Increments the project's `distribution_round` counter.

Each token holder calls `claim_distribution` specifying the round number. The protocol:
1. Reads the investor's `tokens_held` at the time of the snapshot.
2. Calculates: `share = (tokens_held × pool_usdc × DISTRIBUTION_SCALE) / total_circulating`.
3. Transfers the calculated USDC directly from the distribution pool to the investor's account.
4. Marks that round as claimed for that investor.

Rounds must be claimed sequentially. This prevents skip-and-claim exploits where an investor
could skip rounds with small payouts and retroactively claim large ones using a higher token
balance acquired after the fact.

The `DISTRIBUTION_SCALE` constant of 10¹² ensures sub-cent precision even when the total token
supply runs into the billions and individual positions are tiny fractions of the whole.

### Oracle System

The oracle is the bridge between physical construction reality and on-chain state. It is not
a price oracle — it is a trusted verification agent that physically or contractually verifies
construction progress and cryptographically attests to it on-chain.

The oracle keypair is stored in the Go oracle service, managed with encrypted storage and
rotatable through a 48-hour timelock mechanism (`update_oracle`). The timelock prevents a
compromised oracle from being used to fraudulently release funds before the legitimate authority
can intervene.

The Go oracle service runs three components:

**Proof Signer:** Receives milestone completion reports (from human inspectors, automated
sensors, or third-party inspection firms), validates them, constructs and signs the
`submit_milestone_proof` transaction, and submits it to the Solana network.

**REST API:** Serves project data, investor positions, KYC status, and aggregated analytics to
the frontend. Acts as an efficient read layer on top of the on-chain state, backed by a
PostgreSQL database kept in sync by the Listener.

**On-Chain Listener:** Subscribes to Solana's WebSocket interface and parses program log events
in real time. When a relevant event occurs (project state change, milestone proof, dispute
raised, distribution deposited), it updates the PostgreSQL database and triggers any necessary
off-chain actions (e.g., sending notifications, triggering inspection workflows).

### KYC and Compliance

TerraVault is designed from the ground up for compliance with real estate investment regulations
in jurisdictions that require investor verification. The `kyc_required` flag on each project
activates KYC enforcement at the TransferHook level.

When KYC is required:
- The hook checks the receiver's KYC status before allowing any token transfer.
- Non-verified wallets cannot receive tokens, even on secondary markets.
- KYC is verified through the Persona integration in the Go oracle service.
- KYC approval is recorded off-chain and referenced on-chain through a signed oracle attestation.

This design allows TerraVault to serve both permissioned (regulated) and permissionless
(open) investment structures within the same protocol, configurable per project.

---

## Technical Architecture

### On-Chain Program (Rust + Anchor)

The TerraVault on-chain program is deployed at address `DoAFjsoY9Ws7ZTNCokpsYHyNho8Krj9nK5dQFCdgYQqM`
on Solana devnet.

It manages the following PDA accounts:

**ProjectState** — The central account for each project. Contains 500+ bytes of state tracking
the entire project lifecycle, including all fundraising data, milestone progress, distribution
round, dispute state, and configuration. Size is fixed at initialization to avoid account
reallocation.

**MilestoneRecord** — One per construction milestone. Stores the proof URI, proof hash, oracle
signature, milestone type, release bps, status, and dispute window deadline. Indexed by
`(project_pubkey, milestone_index)`.

**InvestorPosition** — One per (project, investor) pair. Tracks `tokens_held`,
`usdc_invested`, and `last_claimed_round`. Updated automatically by the TransferHook on every
token transfer, ensuring the protocol always knows each investor's exact exposure.

**DeveloperProfile** — One per developer wallet. Tracks project history and the `is_blacklisted`
flag. Persists across projects — a blacklisted developer cannot create new projects on any
future deployment.

**DistributionPool** — One per distribution round per project. Records the total pool amount,
snapshot token supply, and a bitmap of claimed positions. Cleaned up after all investors have
claimed (future optimization).

**TokenConfig** — Stores Token-2022 mint configuration including the transfer fee basis points
(capped at 500 bps / 5%), freeze authority, and KYC requirement flag.

### State Machine

The `ProjectStatus` enum defines the project lifecycle with 11 possible states:

```
Draft
  └─ start_fundraising ──▶ Fundraising
                               ├─ cancel_fundraise ──▶ Cancelled
                               └─ activate_project ──▶ Active
                                                          ├─ pause_project ──▶ Paused
                                                          ├─ raise_dispute ──▶ DisputeActive
                                                          │                       └─ admin_resolve ──▶ Resolved
                                                          └─ submit_milestone_proof ──▶ InMilestones
                                                                                           └─ (all released) ──▶ Completed
                                                                                                                     └─ deposit_income ──▶ Distributing
```

State transitions are enforced at the instruction level. Each handler begins with a guard
checking the current `ProjectStatus` and returning a typed error if the transition is invalid.
The `paused` boolean flag adds a second layer — when set to true, all state-mutating instructions
return `ProjectPaused` regardless of the current status.

### Error Handling

The protocol defines 40+ typed errors in `errors.rs`, each corresponding to a specific invalid
condition. Examples:

- `MilestoneBpsNotFull` — `start_fundraising` called before milestone bps sum to 10,000
- `FundraiseDeadlinePassed` — `buy_tokens` called after the fundraising deadline
- `DisputeWindowActive` — `release_milestone_funds` called before the 48h dispute window closes
- `ProjectPaused` — Any state-mutating call while `paused = true`
- `DeveloperBlacklisted` — `initialize_project` called by a blacklisted developer
- `DistributionRoundNotSequential` — `claim_distribution` called out of order
- `OracleTimelockActive` — `update_oracle` finalization called before 48h elapses

All errors are returned as typed Anchor errors with descriptive messages for frontend handling.

### Events

The program emits structured events for every significant state transition. The Go listener
subscribes to these events via WebSocket and uses them to keep the PostgreSQL database in sync.
Events include:

- `ProjectInitialized` — project_id, developer, project_type, total_tokens
- `FundraisingStarted` — project_id, target_usdc, deadline
- `TokensPurchased` — project_id, investor, token_amount, usdc_paid
- `ProjectActivated` — project_id, total_raised_usdc
- `MilestoneProofSubmitted` — project_id, milestone_index, proof_uri, dispute_deadline
- `MilestoneDisputed` — project_id, milestone_index, disputer, reason
- `MilestoneReleased` — project_id, milestone_index, amount_usdc
- `IncomeDeposited` — project_id, round, amount_usdc, source
- `DistributionClaimed` — project_id, round, investor, amount_usdc
- `DisputeRaised` — project_id, raiser, deadline
- `EvidenceSubmitted` — project_id, submitter, evidence_hash
- `DisputeResolved` — project_id, decision, admin
- `DeveloperFlagged` — developer_pubkey, reason_hash
- `ProjectPaused` / `ProjectUnpaused` — project_id, authority

### Go Oracle Service

The oracle service is written in Go 1.21+ and structured as a monorepo with shared internal
packages:

```
backend/
├── cmd/
│   ├── api/main.go          — HTTP REST server (Fastify-style routing)
│   └── listener/main.go     — On-chain WebSocket event listener
├── internal/
│   ├── anchor/              — IDL bindings + Anchor instruction builders
│   ├── oracle/              — Keypair management + tx submission with retry
│   ├── kyc/                 — Persona KYC provider (session initiation + webhook)
│   ├── listener/            — Solana WebSocket log parser + event dispatcher
│   ├── storage/
│   │   ├── postgres/        — Repository pattern + DB migrations (sqlc-generated)
│   │   └── ipfs/            — IPFS upload client (Pinata / local node)
│   └── api/                 — Chi router, middleware (auth, rate limit), handlers
└── pkg/
    └── solana/              — RPC client wrapper with exponential backoff retry
```

The oracle's keypair is the authority for `submit_milestone_proof` and `release_milestone_funds`.
It is generated with a hardware security module in production and rotated through the on-chain
timelock mechanism. The private key never leaves the oracle service.

### Frontend (React + Vite)

The `frontend/` directory contains a React 18 + Vite + TailwindCSS frontend. It connects to Solana
via `@solana/web3.js` and `@coral-xyz/anchor` for direct on-chain interactions (token purchases,
claims, dispute submissions) and queries the Go API for aggregated data (project listings,
investor dashboards, distribution history).

Key views:
- **Project Marketplace** — browse all active projects with live fundraising progress bars
- **Project Detail** — milestone timeline, proof documents, dispute history, income distributions
- **Investor Dashboard** — portfolio overview, token holdings, pending distributions, P&L
- **Developer Portal** — project creation wizard, milestone management, income deposit interface

---

## Security Design

Security in TerraVault is layered across multiple independent mechanisms. No single point of
failure can drain investor funds or allow a malicious actor to extract value unilaterally.

### Escrow Vault Control

The USDC escrow vault is a token account whose authority is the `ProjectState` PDA itself.
No human wallet — not the developer, not the oracle, not the protocol admin — can sign
transactions that move USDC out of the vault directly. The only way to move funds is through
the program's own instruction handlers, each of which enforces strict state machine checks.

### 48-Hour Dispute Window

Even if the oracle submits a fraudulent milestone proof, no funds can be released for 48 hours.
This window exists specifically to allow token holders to identify and challenge invalid proofs.
Any token holder can call `dispute_milestone` with a reason — the mere act of disputing halts
the release and triggers the arbitration process. Token holders do not need to prove fraud to
dispute; they only need to raise the concern. The burden of proof then falls on the oracle to
re-substantiate the proof through the arbitration process.

### 72-Hour DAO Dispute Window

If the milestone-level dispute mechanism is bypassed or a broader project-level issue arises,
any stakeholder can raise a project-level dispute. This transitions the entire project to
`DisputeActive` state — all state-mutating operations are blocked — and opens a 72-hour window
for both parties to submit IPFS evidence. The admin multisig cannot resolve the dispute
instantly; they must wait for the evidence window to close and then execute one of three binding
decisions. This prevents rushed or coerced resolutions.

### Oracle Key Rotation Timelock

If the oracle's private key is compromised, an attacker with access to it could submit false
milestone proofs. The `update_oracle` instruction mitigates this by requiring a two-step process:
1. Call `update_oracle` with the new oracle public key. Records `oracle_update_pending_at` and
   `pending_oracle`.
2. Wait 48 hours.
3. Call `update_oracle` again. The timelock is verified, and the oracle key is rotated.

During the 48-hour window, the old oracle remains active. If the old oracle detects unauthorized
rotation attempts, it has time to pause the project and alert the admin multisig.

### Transfer Fee Cap

The `MAX_TRANSFER_FEE_BPS` constant is hardcoded at 500 (5%) and enforced in
`initialize_project`. Developers cannot configure transfer fees above this limit, preventing a
malicious developer from setting a 100% transfer fee to drain secondary market trading volume.

### Emergency Pause

The protocol admin can call `pause_project` at any time to set `paused = true`. This causes all
state-mutating instructions to return `ProjectPaused` immediately, without executing any logic.
This is a circuit breaker for emergency situations — it stops all fund movement while the team
investigates an issue. `unpause_project` restores normal operation.

### Developer Blacklist

The `flag_developer` instruction sets `is_blacklisted = true` on the developer's
`DeveloperProfile` PDA. This check is enforced at the start of `initialize_project` — a
blacklisted developer cannot create new projects, ever. The blacklist is permanent and
immutable on-chain. Even if the admin multisig is compromised after flagging, the blacklist
cannot be reversed without a program upgrade.

---

## Why Solana

TerraVault's design would be impossible on most other blockchains. Solana's specific properties
are load-bearing for the protocol's correctness and economic viability.

**400ms Block Time.** The 48-hour dispute window requires investors to monitor milestone proofs
and react within a bounded time. On a chain with 12-second block times, the granularity of
on-chain timestamps is measured in minutes. On Solana, it is measured in milliseconds. This
means the protocol can enforce time-sensitive mechanics with precision.

**$0.00025 Per Transaction.** TerraVault requires a large number of small transactions:
token purchases, milestone proof submissions, distribution claims, dispute filings. On Ethereum,
a single `claim_distribution` call at peak gas prices could cost more than the distribution
itself for small investors. On Solana, the gas cost is economically irrelevant at any scale.

**SPL Token-2022 with TransferHook.** The TransferHook extension is unique to Solana's token
standard and is fundamental to TerraVault's position tracking and KYC enforcement architecture.
There is no equivalent mechanism on Ethereum's ERC-20 standard that provides on-chain hooks
on every transfer without a bespoke wrapper token contract. Token-2022 provides this natively.

**Anchor Framework.** Anchor's account validation macros, typed CPI interfaces, and IDL
generation dramatically reduce the attack surface of the program by enforcing account ownership,
discriminator checks, and signer constraints declaratively. The Rust type system, combined with
Anchor's compile-time checks, catches entire categories of bugs (account substitution, missing
signer checks) that have caused hundreds of millions in losses on other chains.

**Ecosystem Depth.** Jupiter's liquidity aggregation means that TerraVault project tokens, once
trading on secondary markets, will have access to deep Solana DeFi liquidity immediately without
requiring TerraVault to bootstrap its own AMM. Pyth's real-time price feeds provide a future
path for on-chain property valuations without centralized data sources.

---

## Comparison with Existing Solutions

### Traditional Real Estate Investment

Traditional direct real estate investment requires title transfer, mortgage origination,
property management contracts, and legal counsel for every transaction. Liquidity is measured
in months, minimum investment is measured in tens of thousands of dollars, and income
distribution is manual and quarterly. TerraVault replaces all of this with on-chain state
transitions that settle in under a second.

### REITs (Real Estate Investment Trusts)

REITs provide fractional exposure to diversified real estate portfolios but do not allow
investors to choose specific properties. They trade at the fund level, are subject to fund
management fees (typically 1–2% annually), and provide no on-chain transparency. TerraVault
provides property-specific investment, zero management fee at the protocol level, and full
on-chain auditability of every dollar.

### RealT / Lofty

Ethereum-based tokenization platforms have demonstrated the demand for fractional real estate
ownership. However, they suffer from Ethereum's gas costs (making small investments economically
unviable), slow block times (limiting time-sensitive mechanics), and no native dispute resolution.
TerraVault brings this concept to Solana with a complete protection framework — milestone escrow,
48h dispute windows, DAO resolution, developer blacklisting — that no existing platform offers.

### Jito (for reference)

Jito is not a real estate protocol, but its architecture as a Solana-native protocol handling
large USDC flows with time-sensitive mechanics (bundle submission windows, auction logic) is
conceptually analogous. TerraVault applies similar design principles — direct on-chain
mechanics, oracle-signed state transitions, time-locked windows — to the real estate domain.

---

## Development Status

TerraVault v1.0 (Hackathon MVP) is functionally complete. The on-chain program compiles, passes
Anchor test suite, and all 23 instructions are implemented and tested:

`initialize_project` · `add_milestone` · `start_fundraising` · `buy_tokens` ·
`cancel_fundraise` · `claim_refund` · `activate_project` · `submit_milestone_proof` ·
`dispute_milestone` · `resolve_dispute` · `release_milestone_funds` · `deposit_income` ·
`claim_distribution` · `update_oracle` · `pause_project` · `unpause_project` ·
`execute_transfer_hook` · `raise_dispute` · `submit_evidence` · `admin_resolve` ·
`freeze_developer` · `flag_developer`

The Go oracle service is operational with all three components (proof signer, REST API,
on-chain listener) functional against a local Solana test validator.

The React frontend renders the project marketplace and investor dashboard against local devnet.

The protocol is deployed on Solana devnet at:
`DoAFjsoY9Ws7ZTNCokpsYHyNho8Krj9nK5dQFCdgYQqM`

---

## Team

| Name | Role |
|------|------|
| — | Founder & Lead Engineer (Rust, Go, Solana) |

---

## License

MIT — see [LICENSE](../LICENSE)
