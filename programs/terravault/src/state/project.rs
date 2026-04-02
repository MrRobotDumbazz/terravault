use anchor_lang::prelude::*;

// Default is NOT derived because Rust only auto-implements it for [u8; N] where N <= 32.
// ProjectState contains [u8; 128] and [u8; 64] fields. Anchor's init macro
// zero-initialises account data directly, so Default is not required.
#[account]
pub struct ProjectState {
    pub project_id: u64,
    pub developer: Pubkey,
    pub oracle_authority: Pubkey,
    pub token_mint: Pubkey,
    pub escrow_vault: Pubkey,
    pub state: ProjectStatus,
    pub project_type: ProjectType,
    pub total_tokens: u64,
    pub tokens_sold: u64,
    pub token_price_usdc: u64,
    pub fundraise_target_usdc: u64,
    pub fundraise_hard_cap_usdc: u64,
    pub fundraise_deadline: i64,
    pub escrow_balance_usdc: u64,
    pub total_raised_usdc: u64,
    pub milestone_count: u8,
    pub milestones_added: u8,
    pub milestones_completed: u8,
    pub current_milestone_index: u8,
    pub milestone_bps_total: u16,
    pub distribution_round: u32,
    pub total_distributed_usdc: u64,
    pub metadata_uri: [u8; 128],
    pub legal_doc_hash: [u8; 32],
    pub oracle_update_pending_at: i64,
    pub pending_oracle: Pubkey,
    pub created_at: i64,
    pub updated_at: i64,
    pub bump: u8,
    pub escrow_bump: u8,
    pub paused: bool,
    pub kyc_required: bool,
    pub transfer_fee_bps: u16,
    // ── Dispute fields (absorbed from reserved) ──────────────────────────────
    pub dispute_deadline: i64,         // unix ts when DisputeActive window closes
    pub evidence_hash: [u8; 32],       // IPFS CID hash of latest submitted evidence
    pub admin_decision: AdminDecision, // outcome set by admin_resolve
    pub reserved: [u8; 23],
}

impl ProjectState {
    pub const LEN: usize = 8  // discriminator
        + 8   // project_id
        + 32  // developer
        + 32  // oracle_authority
        + 32  // token_mint
        + 32  // escrow_vault
        + 1   // state
        + 1   // project_type
        + 8   // total_tokens
        + 8   // tokens_sold
        + 8   // token_price_usdc
        + 8   // fundraise_target_usdc
        + 8   // fundraise_hard_cap_usdc
        + 8   // fundraise_deadline
        + 8   // escrow_balance_usdc
        + 8   // total_raised_usdc
        + 1   // milestone_count
        + 1   // milestones_added
        + 1   // milestones_completed
        + 1   // current_milestone_index
        + 2   // milestone_bps_total
        + 4   // distribution_round
        + 8   // total_distributed_usdc
        + 128 // metadata_uri
        + 32  // legal_doc_hash
        + 8   // oracle_update_pending_at
        + 32  // pending_oracle
        + 8   // created_at
        + 8   // updated_at
        + 1   // bump
        + 1   // escrow_bump
        + 1   // paused
        + 1   // kyc_required
        + 2   // transfer_fee_bps
        + 8   // dispute_deadline
        + 32  // evidence_hash
        + 1   // admin_decision
        + 23; // reserved
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, PartialEq, Eq, Default, Debug)]
pub enum ProjectStatus {
    #[default]
    Draft,
    Fundraising,
    Active,
    InMilestones,
    Completed,
    Distributing,
    Closed,
    Cancelled,
    Paused,
    DisputeActive, // project-level dispute raised; 72h resolution window
    Resolved,      // admin_resolve has been called and executed
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, PartialEq, Eq, Default, Debug)]
pub enum AdminDecision {
    #[default]
    None,
    PayInvestors,    // freeze dev collateral, distribute to token holders
    RefundAndExtend, // return escrow to investors, set new deadline
    ForceClose,      // split escrow proportionally and close project
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, PartialEq, Eq, Default, Debug)]
pub enum ProjectType {
    #[default]
    Residential,
    Commercial,
    Agricultural,
    Mixed,
}
