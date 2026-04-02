use anchor_lang::prelude::*;

#[account]
#[derive(Default)]
pub struct InvestorPosition {
    pub project: Pubkey,
    pub investor: Pubkey,
    pub tokens_held: u64,
    pub usdc_invested: u64,
    pub last_claimed_round: u32,
    pub total_claimed_usdc: u64,
    pub kyc_verified: bool,
    pub kyc_timestamp: i64,
    pub created_at: i64,
    pub bump: u8,
    pub reserved: [u8; 32],
}

impl InvestorPosition {
    pub const LEN: usize = 8   // discriminator
        + 32  // project
        + 32  // investor
        + 8   // tokens_held
        + 8   // usdc_invested
        + 4   // last_claimed_round
        + 8   // total_claimed_usdc
        + 1   // kyc_verified
        + 8   // kyc_timestamp
        + 8   // created_at
        + 1   // bump
        + 32; // reserved
}
