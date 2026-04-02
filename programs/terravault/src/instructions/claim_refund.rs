use anchor_lang::prelude::*;
use anchor_spl::{
    token_2022::Token2022,
    token_interface::{Mint, TokenAccount, TransferChecked, transfer_checked},
};
use crate::{
    errors::TerraVaultError,
    events::RefundClaimed,
    state::{InvestorPosition, ProjectState, ProjectStatus},
};

#[derive(Accounts)]
pub struct ClaimRefund<'info> {
    pub investor: Signer<'info>,

    #[account(
        constraint = project_state.state == ProjectStatus::Cancelled @ TerraVaultError::InvalidProjectState,
    )]
    pub project_state: Account<'info, ProjectState>,

    #[account(
        mut,
        seeds = [b"position", project_state.key().as_ref(), investor.key().as_ref()],
        bump = investor_position.bump,
        constraint = investor_position.investor == investor.key(),
        constraint = investor_position.usdc_invested > 0 @ TerraVaultError::NoInvestmentToRefund,
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
}

pub fn handler(ctx: Context<ClaimRefund>) -> Result<()> {
    let clock = Clock::get()?;
    let project = &ctx.accounts.project_state;
    let position = &mut ctx.accounts.investor_position;

    let refund_amount = position.usdc_invested;
    let tokens_burned = position.tokens_held;

    // Zero out BEFORE CPI
    position.usdc_invested = 0;
    position.tokens_held = 0;

    // PDA signer seeds for escrow vault
    let project_key = project.key();
    let seeds: &[&[u8]] = &[
        b"escrow",
        project_key.as_ref(),
        &[project.escrow_bump],
    ];
    let signer_seeds = &[seeds];

    let cpi_ctx = CpiContext::new_with_signer(
        ctx.accounts.token_program.to_account_info(),
        TransferChecked {
            from: ctx.accounts.escrow_vault.to_account_info(),
            mint: ctx.accounts.usdc_mint.to_account_info(),
            to: ctx.accounts.investor_usdc_account.to_account_info(),
            authority: ctx.accounts.escrow_vault.to_account_info(),
        },
        signer_seeds,
    );
    transfer_checked(cpi_ctx, refund_amount, 6)?;

    emit!(RefundClaimed {
        project: project.key(),
        investor: ctx.accounts.investor.key(),
        usdc_refunded: refund_amount,
        tokens_burned,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
