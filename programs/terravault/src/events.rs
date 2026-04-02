use anchor_lang::prelude::*;

#[event]
pub struct ProjectCreated {
    pub project: Pubkey,
    pub developer: Pubkey,
    pub project_id: u64,
    pub total_tokens: u64,
    pub token_price_usdc: u64,
    pub timestamp: i64,
}

#[event]
pub struct FundraisingStarted {
    pub project: Pubkey,
    pub fundraise_target: u64,
    pub hard_cap: u64,
    pub deadline: i64,
    pub timestamp: i64,
}

#[event]
pub struct TokensPurchased {
    pub project: Pubkey,
    pub investor: Pubkey,
    pub token_amount: u64,
    pub usdc_paid: u64,
    pub timestamp: i64,
}

#[event]
pub struct ProjectActivated {
    pub project: Pubkey,
    pub total_raised: u64,
    pub timestamp: i64,
}

#[event]
pub struct FundraiseCancelled {
    pub project: Pubkey,
    pub total_raised: u64,
    pub reason: [u8; 64],
    pub timestamp: i64,
}

#[event]
pub struct MilestoneProofSubmitted {
    pub project: Pubkey,
    pub milestone_index: u8,
    pub proof_uri: [u8; 128],
    pub proof_hash: [u8; 32],
    pub dispute_deadline: i64,
    pub timestamp: i64,
}

#[event]
pub struct MilestoneDisputed {
    pub project: Pubkey,
    pub milestone_index: u8,
    pub investor: Pubkey,
    pub timestamp: i64,
}

#[event]
pub struct DisputeResolved {
    pub project: Pubkey,
    pub milestone_index: u8,
    pub approved: bool,
    pub timestamp: i64,
}

#[event]
pub struct MilestoneFundsReleased {
    pub project: Pubkey,
    pub milestone_index: u8,
    pub amount_released: u64,
    pub developer: Pubkey,
    pub timestamp: i64,
}

#[event]
pub struct ProjectCompleted {
    pub project: Pubkey,
    pub developer: Pubkey,
    pub timestamp: i64,
}

#[event]
pub struct IncomeDeposited {
    pub project: Pubkey,
    pub round: u32,
    pub amount_usdc: u64,
    pub total_tokens_snapshot: u64,
    pub timestamp: i64,
}

#[event]
pub struct DistributionClaimed {
    pub project: Pubkey,
    pub investor: Pubkey,
    pub round: u32,
    pub amount_usdc: u64,
    pub timestamp: i64,
}

#[event]
pub struct ProjectPaused {
    pub project: Pubkey,
    pub timestamp: i64,
}

#[event]
pub struct ProjectUnpaused {
    pub project: Pubkey,
    pub timestamp: i64,
}

#[event]
pub struct OracleUpdated {
    pub project: Pubkey,
    pub old_oracle: Pubkey,
    pub new_oracle: Pubkey,
    pub effective_at: i64,
    pub timestamp: i64,
}

#[event]
pub struct RefundClaimed {
    pub project: Pubkey,
    pub investor: Pubkey,
    pub usdc_refunded: u64,
    pub tokens_burned: u64,
    pub timestamp: i64,
}

#[event]
pub struct TokensTransferred {
    pub project: Pubkey,
    pub from: Pubkey,
    pub to: Pubkey,
    pub amount: u64,
    pub timestamp: i64,
}

#[event]
pub struct DisputeRaised {
    pub project: Pubkey,
    pub raised_by: Pubkey,
    pub reason_hash: [u8; 32],
    pub deadline: i64,
    pub timestamp: i64,
}

#[event]
pub struct EvidenceSubmitted {
    pub project: Pubkey,
    pub submitted_by: Pubkey,
    pub evidence_hash: [u8; 32],
    pub timestamp: i64,
}

#[event]
pub struct AdminResolved {
    pub project: Pubkey,
    pub decision: u8, // 1=PayInvestors, 2=RefundAndExtend, 3=ForceClose
    pub timestamp: i64,
}

#[event]
pub struct FraudConfirmed {
    pub project: Pubkey,
    pub developer: Pubkey,
    pub collateral_distributed: u64,
    pub timestamp: i64,
}

#[event]
pub struct DeveloperFlagged {
    pub developer: Pubkey,
    pub reason_hash: [u8; 32],
    pub timestamp: i64,
}
