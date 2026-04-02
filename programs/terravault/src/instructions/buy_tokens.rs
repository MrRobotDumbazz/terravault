use anchor_lang::prelude::*;
use anchor_spl::{
    token_2022::Token2022,
    token_interface::{Mint, TokenAccount, TransferChecked, transfer_checked},
};
use crate::{
    errors::TerraVaultError,
    events::TokensPurchased,
    state::{InvestorPosition, ProjectState, ProjectStatus},
};

#[derive(Accounts)]
pub struct BuyTokens<'info> {
    #[account(mut)]
    pub investor: Signer<'info>,

    #[account(
        mut,
        constraint = project_state.state == ProjectStatus::Fundraising @ TerraVaultError::InvalidProjectState,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,

    #[account(
        init_if_needed,
        payer = investor,
        space = InvestorPosition::LEN,
        seeds = [b"position", project_state.key().as_ref(), investor.key().as_ref()],
        bump
    )]
    pub investor_position: Account<'info, InvestorPosition>,

    #[account(
        mut,
        seeds = [b"escrow", project_state.key().as_ref()],
        bump = project_state.escrow_bump,
    )]
    pub escrow_vault: InterfaceAccount<'info, TokenAccount>,

    #[account(mut)]
    pub investor_usdc_account: InterfaceAccount<'info, TokenAccount>,

    pub usdc_mint: InterfaceAccount<'info, Mint>,

    pub token_program: Program<'info, Token2022>,
    pub system_program: Program<'info, System>,
}

pub fn handler(ctx: Context<BuyTokens>, token_amount: u64) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;
    let position = &mut ctx.accounts.investor_position;

    require!(token_amount > 0, TerraVaultError::ZeroTokenAmount);
    require!(
        clock.unix_timestamp < project.fundraise_deadline,
        TerraVaultError::FundraisingDeadlinePassed
    );

    // KYC gate
    if project.kyc_required {
        // Position must already exist and be KYC-verified
        require!(
            position.project != Pubkey::default() && position.kyc_verified,
            TerraVaultError::KycNotVerified
        );
    }

    // Hard cap check
    let new_tokens_sold = project
        .tokens_sold
        .checked_add(token_amount)
        .ok_or(TerraVaultError::MathOverflow)?;
    require!(
        new_tokens_sold <= project.total_tokens,
        TerraVaultError::HardCapExceeded
    );

    // USDC cost via u128 to prevent overflow
    let usdc_cost: u64 = (token_amount as u128)
        .checked_mul(project.token_price_usdc as u128)
        .ok_or(TerraVaultError::MathOverflow)?
        .try_into()
        .map_err(|_| TerraVaultError::MathOverflow)?;

    // Update state BEFORE CPI (re-entrancy pattern)
    project.tokens_sold = new_tokens_sold;
    project.total_raised_usdc = project
        .total_raised_usdc
        .checked_add(usdc_cost)
        .ok_or(TerraVaultError::MathOverflow)?;
    project.escrow_balance_usdc = project
        .escrow_balance_usdc
        .checked_add(usdc_cost)
        .ok_or(TerraVaultError::MathOverflow)?;
    project.updated_at = clock.unix_timestamp;

    if position.project == Pubkey::default() {
        position.project = project.key();
        position.investor = ctx.accounts.investor.key();
        position.created_at = clock.unix_timestamp;
        position.bump = ctx.bumps.investor_position;
    }
    position.tokens_held = position
        .tokens_held
        .checked_add(token_amount)
        .ok_or(TerraVaultError::MathOverflow)?;
    position.usdc_invested = position
        .usdc_invested
        .checked_add(usdc_cost)
        .ok_or(TerraVaultError::MathOverflow)?;

    // CPI: USDC investor → escrow
    let cpi_ctx = CpiContext::new(
        ctx.accounts.token_program.to_account_info(),
        TransferChecked {
            from: ctx.accounts.investor_usdc_account.to_account_info(),
            mint: ctx.accounts.usdc_mint.to_account_info(),
            to: ctx.accounts.escrow_vault.to_account_info(),
            authority: ctx.accounts.investor.to_account_info(),
        },
    );
    transfer_checked(cpi_ctx, usdc_cost, 6)?;

    emit!(TokensPurchased {
        project: project.key(),
        investor: ctx.accounts.investor.key(),
        token_amount,
        usdc_paid: usdc_cost,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
