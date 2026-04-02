use anchor_lang::prelude::*;
use anchor_spl::{
    token_2022::Token2022,
    token_interface::{Mint, TokenAccount, TransferChecked, transfer_checked},
};
use crate::{
    constants::DISTRIBUTION_SCALE,
    errors::TerraVaultError,
    events::IncomeDeposited,
    state::{DistributionPool, DistributionSource, ProjectState, ProjectStatus},
};

#[derive(Accounts)]
pub struct DepositIncome<'info> {
    #[account(mut)]
    pub developer: Signer<'info>,

    #[account(
        mut,
        has_one = developer @ TerraVaultError::InvalidDeveloper,
        constraint = (
            project_state.state == ProjectStatus::Completed
            || project_state.state == ProjectStatus::Distributing
        ) @ TerraVaultError::InvalidProjectState,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,

    #[account(
        init,
        payer = developer,
        space = DistributionPool::LEN,
        seeds = [
            b"distribution",
            project_state.key().as_ref(),
            &(project_state.distribution_round + 1).to_le_bytes()
        ],
        bump
    )]
    pub distribution_pool: Account<'info, DistributionPool>,

    #[account(
        mut,
        seeds = [b"income", project_state.key().as_ref()],
        bump
    )]
    pub income_vault: InterfaceAccount<'info, TokenAccount>,

    #[account(mut)]
    pub developer_usdc_account: InterfaceAccount<'info, TokenAccount>,

    /// Project token mint — snapshot of supply taken here
    pub project_token_mint: InterfaceAccount<'info, Mint>,

    pub usdc_mint: InterfaceAccount<'info, Mint>,
    pub token_program: Program<'info, Token2022>,
    pub system_program: Program<'info, System>,
}

pub fn handler(
    ctx: Context<DepositIncome>,
    amount_usdc: u64,
    source: DistributionSource,
) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;

    require!(amount_usdc > 0, TerraVaultError::ZeroDistributionAmount);

    let total_supply = ctx.accounts.project_token_mint.supply;
    require!(total_supply > 0, TerraVaultError::DivisionByZero);

    let new_round = project
        .distribution_round
        .checked_add(1)
        .ok_or(TerraVaultError::MathOverflow)?;

    // Fixed-point price: (amount * 10^12) / supply
    let usdc_per_token_scaled = (amount_usdc as u128)
        .checked_mul(DISTRIBUTION_SCALE)
        .ok_or(TerraVaultError::MathOverflow)?
        .checked_div(total_supply as u128)
        .ok_or(TerraVaultError::DivisionByZero)?;

    // Update state BEFORE CPI
    project.distribution_round = new_round;
    project.total_distributed_usdc = project
        .total_distributed_usdc
        .checked_add(amount_usdc)
        .ok_or(TerraVaultError::MathOverflow)?;
    if project.state == ProjectStatus::Completed {
        project.state = ProjectStatus::Distributing;
    }
    project.updated_at = clock.unix_timestamp;

    let pool = &mut ctx.accounts.distribution_pool;
    pool.project = project.key();
    pool.round = new_round;
    pool.total_usdc_deposited = amount_usdc;
    pool.total_tokens_at_snapshot = total_supply;
    pool.usdc_per_token_scaled = usdc_per_token_scaled;
    pool.deposited_at = clock.unix_timestamp;
    pool.source = source;
    pool.claim_deadline = 0;
    pool.total_claimed = 0;
    pool.bump = ctx.bumps.distribution_pool;
    pool.reserved = [0u8; 32];

    // CPI: USDC developer → income_vault
    let cpi_ctx = CpiContext::new(
        ctx.accounts.token_program.to_account_info(),
        TransferChecked {
            from: ctx.accounts.developer_usdc_account.to_account_info(),
            mint: ctx.accounts.usdc_mint.to_account_info(),
            to: ctx.accounts.income_vault.to_account_info(),
            authority: ctx.accounts.developer.to_account_info(),
        },
    );
    transfer_checked(cpi_ctx, amount_usdc, 6)?;

    emit!(IncomeDeposited {
        project: project.key(),
        round: new_round,
        amount_usdc,
        total_tokens_snapshot: total_supply,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
