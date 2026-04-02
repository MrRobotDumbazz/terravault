use anchor_lang::prelude::*;

// Default not derived — contains [u8; 64] and [u8; 128] fields (> 32 limit).
#[account]
pub struct MilestoneRecord {
    pub project: Pubkey,
    pub milestone_index: u8,
    pub milestone_type: MilestoneType,
    pub description: [u8; 64],
    pub release_bps: u16,
    pub status: MilestoneStatus,
    pub proof_uri: [u8; 128],
    pub proof_hash: [u8; 32],
    pub submitted_at: i64,
    pub approved_at: i64,
    pub released_amount_usdc: u64,
    pub oracle_signature: [u8; 64],
    pub dispute_deadline: i64,
    pub bump: u8,
    pub reserved: [u8; 32],
}

impl MilestoneRecord {
    pub const LEN: usize = 8   // discriminator
        + 32  // project
        + 1   // milestone_index
        + 1   // milestone_type
        + 64  // description
        + 2   // release_bps
        + 1   // status
        + 128 // proof_uri
        + 32  // proof_hash
        + 8   // submitted_at
        + 8   // approved_at
        + 8   // released_amount_usdc
        + 64  // oracle_signature
        + 8   // dispute_deadline
        + 1   // bump
        + 32; // reserved
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, PartialEq, Eq, Default, Debug)]
pub enum MilestoneType {
    #[default]
    SitePreparation,
    Foundation,
    Framing,
    Roofing,
    MEP,               // Mechanical, Electrical, Plumbing
    InteriorWork,
    Landscaping,
    Completion,
    Custom,
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, PartialEq, Eq, Default, Debug)]
pub enum MilestoneStatus {
    #[default]
    Pending,
    UnderReview,
    Approved,
    Released,
    Disputed,
}
