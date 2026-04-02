use anchor_lang::prelude::*;

#[account]
#[derive(Default)]
pub struct TokenConfig {
    pub project: Pubkey,
    pub mint: Pubkey,
    pub transfer_fee_bps: u16,
    pub transfer_fee_max_lamports: u64,
    pub fee_destination: Pubkey,
    /// Nominal interest rate in bps for display purposes (0 = none)
    pub interest_rate_bps: i16,
    pub bump: u8,
}

impl TokenConfig {
    pub const LEN: usize = 8   // discriminator
        + 32  // project
        + 32  // mint
        + 2   // transfer_fee_bps
        + 8   // transfer_fee_max_lamports
        + 32  // fee_destination
        + 2   // interest_rate_bps
        + 1;  // bump
}
