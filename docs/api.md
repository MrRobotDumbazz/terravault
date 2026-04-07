# API Reference

## On-Chain Instructions

### Project Management

#### `initialize_project`
Create a new real estate project. Initializes `ProjectState` PDA, `TokenConfig` PDA, and USDC escrow vault.

```typescript
await program.methods
  .initializeProject({
    projectId: new BN(1),
    projectType: { residential: {} },
    totalTokens: new BN(1_000_000),
    tokenPriceUsdc: new BN(10_000_000), // 10 USDC (6 decimals)
    fundraiseTargetUsdc: new BN(500_000_000_000), // 500,000 USDC
    fundraiseHardCapUsdc: new BN(1_000_000_000_000),
    fundraiseDeadline: new BN(Math.floor(Date.now() / 1000) + 30 * 86400),
    milestoneCount: 4,
    metadataUri: Buffer.from("ipfs://Qm...").slice(0, 128),
    legalDocHash: Buffer.alloc(32),
    kycRequired: true,
    transferFeeBps: 100, // 1%
  })
  .accounts({ ... })
  .rpc();
```

#### `add_milestone`
Add a construction milestone to a Draft project. Must be called sequentially (index 0, 1, 2…). All `release_bps` must sum to 10,000.

```typescript
await program.methods
  .addMilestone({
    milestoneIndex: 0,
    milestoneType: { foundation: {} },
    description: Buffer.from("Foundation pour and inspection"),
    releaseBps: 2500, // 25% of escrow
  })
  .accounts({ ... })
  .rpc();
```

#### `start_fundraising`
Transition `Draft → Fundraising`. Requires all milestones added and `milestone_bps_total == 10,000`.

#### `buy_tokens`
Investor purchases fractional tokens with USDC. USDC is held in the escrow vault.

```typescript
await program.methods
  .buyTokens(new BN(1000)) // token amount
  .accounts({ investor, investorUsdc, escrowVault, tokenMint, ... })
  .rpc();
```

#### `activate_project`
Transition `Fundraising → Active` once soft cap is reached. Tokens become transferable.

#### `cancel_fundraise` / `claim_refund`
Cancel fundraising before soft cap. Investors call `claim_refund` to recover USDC.

---

### Milestone Execution

#### `submit_milestone_proof`
Oracle submits construction proof. Opens 48-hour dispute window.

```typescript
await program.methods
  .submitMilestoneProof({
    milestoneIndex: 0,
    proofUri: Buffer.from("ipfs://Qm...proof"),
    proofHash: Buffer.alloc(32, proofHashBytes),
  })
  .accounts({ oracle, projectState, milestoneRecord, ... })
  .rpc();
```

#### `dispute_milestone`
Token holder challenges a proof during the dispute window.

```typescript
await program.methods
  .disputeMilestone(0, Buffer.from("Concrete test failed").slice(0, 128))
  .accounts({ tokenHolder, projectState, milestoneRecord, ... })
  .rpc();
```

#### `resolve_dispute`
Arbitration authority resolves a disputed milestone.
- `approved = true` → milestone re-opens for fund release
- `approved = false` → developer must resubmit proof

#### `release_milestone_funds`
Oracle releases USDC tranche to developer after dispute window elapses.

---

### Income Distribution

#### `deposit_income`
Developer deposits rental income or sale proceeds.

```typescript
await program.methods
  .depositIncome(new BN(50_000_000_000), { rental: {} }) // 50,000 USDC
  .accounts({ developer, developerUsdc, distributionPool, ... })
  .rpc();
```

#### `claim_distribution`
Token holder claims proportional USDC for a distribution round. Rounds must be claimed sequentially.

```typescript
await program.methods
  .claimDistribution(0) // round number
  .accounts({ investor, investorPosition, distributionPool, ... })
  .rpc();
```

---

### Dispute Resolution (DAO)

#### `raise_dispute`
Investor or stakeholder raises a project-level dispute. Transitions project → `DisputeActive`. Starts 72-hour resolution window.

```typescript
await program.methods
  .raiseDispute(reasonHashBytes) // [u8; 32] keccak of reason
  .accounts({ raiser, projectState, ... })
  .rpc();
```

#### `submit_evidence`
Either party submits IPFS evidence during the dispute window.

```typescript
await program.methods
  .submitEvidence(evidenceHashBytes) // IPFS CID hash [u8; 32]
  .accounts({ submitter, projectState, ... })
  .rpc();
```

#### `admin_resolve`
Admin multisig resolves a project dispute.

```typescript
await program.methods
  .adminResolve({ payInvestors: {} }) // or { refundAndExtend: {} } | { forceClose: {} }
  .accounts({ adminMultisig, projectState, ... })
  .rpc();
```

| Decision | Effect |
|----------|--------|
| `PayInvestors` | Freeze developer token account, distribute escrow to token holders |
| `RefundAndExtend` | Return escrow to investors, set new deadline |
| `ForceClose` | Split escrow proportionally, close project |

#### `freeze_developer`
Freeze developer's token account after `PayInvestors` decision.

#### `flag_developer`
Mark developer as blacklisted. Prevents future project creation.

---

## Go REST API

### Projects

#### `GET /projects`
List all projects with status, fundraising progress, and APY.

**Response:**
```json
[
  {
    "project_id": 1,
    "name": "Horizon Residences",
    "type": "Residential",
    "status": "Fundraising",
    "total_raised_usdc": 250000,
    "fundraise_target_usdc": 500000,
    "token_price_usdc": 10.0,
    "tokens_available": 750000
  }
]
```

#### `GET /projects/:id`
Full project details including milestone breakdown and distribution history.

#### `GET /projects/:id/milestones`
Milestone list with proof URIs, statuses, and dispute history.

---

### Investors

#### `GET /investors/:wallet/positions`
All positions for a wallet — token balances, unclaimed distributions.

**Response:**
```json
[
  {
    "project_id": 1,
    "tokens_held": 5000,
    "usdc_invested": 50000,
    "unclaimed_usdc": 1250.50,
    "last_claimed_round": 2
  }
]
```

#### `GET /investors/:wallet/distributions/:project_id`
Claimable distribution rounds for a specific project.

---

### Oracle

#### `POST /oracle/submit-proof`
*Internal endpoint.* Trigger oracle to submit a verified construction milestone proof.

**Request:**
```json
{
  "project_id": 1,
  "milestone_index": 2,
  "proof_uri": "ipfs://QmXyz...",
  "proof_hash": "0xabc..."
}
```

---

### KYC

#### `GET /kyc/:wallet/status`
Check KYC verification status for a wallet.

**Response:**
```json
{
  "wallet": "ABC...XYZ",
  "status": "approved",
  "verified_at": "2026-03-15T10:00:00Z"
}
```

#### `POST /kyc/initiate`
Start KYC flow for a wallet (returns Persona session URL).

---

### Health

#### `GET /health`
Returns oracle node status, current Solana slot, and DB connection state.

```json
{
  "status": "ok",
  "current_slot": 312847291,
  "oracle_pubkey": "TVaULT...",
  "db": "connected"
}
```

---

## TypeScript SDK (planned)

```typescript
import { TerraVaultClient } from '@terravault/sdk';

const client = new TerraVaultClient({
  rpcUrl: process.env.SOLANA_RPC_URL,
  apiUrl: 'http://localhost:8080',
  wallet: myWallet,
});

// Buy fractional tokens
const result = await client.buyTokens({
  projectId: 1,
  tokenAmount: 1000,
});

// Claim income distribution
await client.claimDistribution({ projectId: 1, round: 3 });
```
