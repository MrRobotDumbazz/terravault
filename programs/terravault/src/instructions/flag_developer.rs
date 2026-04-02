use anchor_lang::prelude::*;
use crate::{
    errors::TerraVaultError,
    events::DeveloperFlagged,
    state::DeveloperProfile,
};

use crate::instructions::admin_resolve::ADMIN_PUBKEY;

#[derive(Accounts)]
#[instruction(developer_pubkey: Pubkey)]
pub struct FlagDeveloper<'info> {
    #[account(mut)]
    pub admin: Signer<'info>,

    #[account(
        constraint = admin.key() == ADMIN_PUBKEY @ TerraVaultError::InvalidAdmin,
    )]
    /// CHECK: constraint above validates admin pubkey.
    pub admin_check: UncheckedAccount<'info>,

    #[account(
        mut,
        seeds = [DeveloperProfile::SEED, developer_pubkey.as_ref()],
        bump = developer_profile.bump,
    )]
    pub developer_profile: Account<'info, DeveloperProfile>,
}

pub fn handler(
    ctx: Context<FlagDeveloper>,
    _developer_pubkey: Pubkey,
    reason_hash: [u8; 32],
) -> Result<()> {
    let clock = Clock::get()?;
    let profile = &mut ctx.accounts.developer_profile;

    profile.is_blacklisted = true;

    emit!(DeveloperFlagged {
        developer: profile.developer,
        reason_hash,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
