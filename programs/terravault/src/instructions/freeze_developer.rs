use anchor_lang::prelude::*;
use anchor_spl::token_2022::Token2022;
use anchor_spl::token_interface::TokenAccount;
use crate::{
    errors::TerraVaultError,
    events::FraudConfirmed,
    state::{AdminDecision, ProjectState, ProjectStatus},
};

use crate::instructions::admin_resolve::ADMIN_PUBKEY;

#[derive(Accounts)]
pub struct FreezeDeveloper<'info> {
    /// Admin authority — same check as admin_resolve.
    #[account(
        constraint = admin.key() == ADMIN_PUBKEY @ TerraVaultError::InvalidAdmin,
    )]
    pub admin: Signer<'info>,

    #[account(
        constraint = project_state.state == ProjectStatus::Resolved @ TerraVaultError::InvalidProjectState,
        constraint = project_state.admin_decision == AdminDecision::PayInvestors
            @ TerraVaultError::InvalidProjectState,
    )]
    pub project_state: Account<'info, ProjectState>,

    /// Developer's token account holding their collateral.
    /// CHECK: validated via constraint against stored developer pubkey.
    #[account(
        mut,
        token::mint = project_state.token_mint,
        token::authority = project_state.developer,
        token::token_program = token_program,
    )]
    pub developer_token_account: InterfaceAccount<'info, TokenAccount>,

    /// The project token mint — must have freeze_authority set to project_state PDA.
    /// CHECK: validated via Token-2022 freeze CPI below.
    #[account(mut)]
    pub token_mint: UncheckedAccount<'info>,

    pub token_program: Program<'info, Token2022>,
}

pub fn handler(ctx: Context<FreezeDeveloper>) -> Result<()> {
    let clock = Clock::get()?;
    let project = &ctx.accounts.project_state;

    // Freeze the developer's token account using Token-2022 freeze_account.
    // The freeze_authority is the project_state PDA itself.
    let project_key = project.key();
    let developer_key = project.developer;
    let project_id_bytes = project.project_id.to_le_bytes();

    let signer_seeds: &[&[&[u8]]] = &[&[
        b"project",
        developer_key.as_ref(),
        &project_id_bytes,
        &[project.bump],
    ]];

    anchor_spl::token_2022::freeze_account(
        CpiContext::new_with_signer(
            ctx.accounts.token_program.to_account_info(),
            anchor_spl::token_2022::FreezeAccount {
                account: ctx.accounts.developer_token_account.to_account_info(),
                mint: ctx.accounts.token_mint.to_account_info(),
                authority: ctx.accounts.project_state.to_account_info(),
            },
            signer_seeds,
        ),
    )?;

    emit!(FraudConfirmed {
        project: project_key,
        developer: developer_key,
        collateral_distributed: ctx.accounts.developer_token_account.amount,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
