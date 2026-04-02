use anchor_lang::prelude::*;

#[account]
#[derive(Default)]
pub struct DistributionPool {
    pub project: Pubkey,
    pub round: u32,
    pub total_usdc_deposited: u64,
    pub total_tokens_at_snapshot: u64,
    /// Fixed-point: (total_usdc * 10^12) / total_tokens
    pub usdc_per_token_scaled: u128,
    pub deposited_at: i64,
    pub source: DistributionSource,
    /// 0 = no expiry
    pub claim_deadline: i64,
    pub total_claimed: u64,
    pub bump: u8,
    pub reserved: [u8; 32],
}

impl DistributionPool {
    pub const LEN: usize = 8   // discriminator
        + 32  // project
        + 4   // round
        + 8   // total_usdc_deposited
        + 8   // total_tokens_at_snapshot
        + 16  // usdc_per_token_scaled (u128)
        + 8   // deposited_at
        + 1   // source
        + 8   // claim_deadline
        + 8   // total_claimed
        + 1   // bump
        + 32; // reserved
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, PartialEq, Eq, Default, Debug)]
pub enum DistributionSource {
    #[default]
    RentalIncome,
    SaleProceeds,
    Refinancing,
    Other,
}
