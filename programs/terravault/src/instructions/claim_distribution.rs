use anchor_lang::prelude::*;
use anchor_spl::{
    token_2022::Token2022,
    token_interface::{Mint, TokenAccount, TransferChecked, transfer_checked},
};
use crate::{
    constants::DISTRIBUTION_SCALE,
    errors::TerraVaultError,
    events::DistributionClaimed,
    state::{DistributionPool, InvestorPosition, ProjectState},
};

#[derive(Accounts)]
#[instruction(round: u32)]
pub struct ClaimDistribution<'info> {
    pub investor: Signer<'info>,

    pub project_state: Account<'info, ProjectState>,

    #[account(
        mut,
        seeds = [b"position", project_state.key().as_ref(), investor.key().as_ref()],
        bump = investor_position.bump,
        constraint = investor_position.investor == investor.key(),
        constraint = investor_position.tokens_held > 0 @ TerraVaultError::NoTokensHeld,
    )]
    pub investor_position: Account<'info, InvestorPosition>,

    #[account(
        mut,
        seeds = [b"distribution", project_state.key().as_ref(), &round.to_le_bytes()],
        bump = distribution_pool.bump,
        constraint = distribution_pool.round == round,
    )]
    pub distribution_pool: Account<'info, DistributionPool>,

    #[account(
        mut,
        seeds = [b"income", project_state.key().as_ref()],
        bump
    )]
    pub income_vault: InterfaceAccount<'info, TokenAccount>,

    #[account(mut)]
    pub investor_usdc_account: InterfaceAccount<'info, TokenAccount>,

    pub usdc_mint: InterfaceAccount<'info, Mint>,
    pub token_program: Program<'info, Token2022>,
}

pub fn handler(ctx: Context<ClaimDistribution>, round: u32) -> Result<()> {
    let clock = Clock::get()?;
    let position = &mut ctx.accounts.investor_position;
    let pool = &mut ctx.accounts.distribution_pool;
    let project = &ctx.accounts.project_state;

    // Sequential claim enforcement
    require!(
        round == position.last_claimed_round + 1,
        TerraVaultError::DistributionRoundSkipped
    );

    // Proportional claimable via fixed-point math
    let claimable: u64 = (position.tokens_held as u128)
        .checked_mul(pool.usdc_per_token_scaled)
        .ok_or(TerraVaultError::MathOverflow)?
        .checked_div(DISTRIBUTION_SCALE)
        .ok_or(TerraVaultError::DivisionByZero)?
        .try_into()
        .map_err(|_| TerraVaultError::MathOverflow)?;

    // Update state BEFORE CPI
    position.last_claimed_round = round;
    position.total_claimed_usdc = position
        .total_claimed_usdc
        .checked_add(claimable)
        .ok_or(TerraVaultError::MathOverflow)?;
    pool.total_claimed = pool
        .total_claimed
        .checked_add(claimable)
        .ok_or(TerraVaultError::MathOverflow)?;

    // PDA signer seeds for income vault
    let project_key = project.key();
    let income_bump = ctx.bumps.income_vault;
    let seeds: &[&[u8]] = &[b"income", project_key.as_ref(), &[income_bump]];
    let signer_seeds = &[seeds];

    let cpi_ctx = CpiContext::new_with_signer(
        ctx.accounts.token_program.to_account_info(),
        TransferChecked {
            from: ctx.accounts.income_vault.to_account_info(),
            mint: ctx.accounts.usdc_mint.to_account_info(),
            to: ctx.accounts.investor_usdc_account.to_account_info(),
            authority: ctx.accounts.income_vault.to_account_info(),
        },
        signer_seeds,
    );
    transfer_checked(cpi_ctx, claimable, 6)?;

    emit!(DistributionClaimed {
        project: project.key(),
        investor: ctx.accounts.investor.key(),
        round,
        amount_usdc: claimable,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
