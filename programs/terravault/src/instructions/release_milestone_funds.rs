use anchor_lang::prelude::*;
use anchor_spl::{
    token_2022::Token2022,
    token_interface::{Mint, TokenAccount, TransferChecked, transfer_checked},
};
use crate::{
    constants::BASIS_POINTS_DIVISOR,
    errors::TerraVaultError,
    events::{MilestoneFundsReleased, ProjectCompleted},
    state::{MilestoneRecord, MilestoneStatus, ProjectState, ProjectStatus},
};

#[derive(Accounts)]
#[instruction(milestone_index: u8)]
pub struct ReleaseMilestoneFunds<'info> {
    pub oracle_authority: Signer<'info>,

    #[account(
        mut,
        constraint = project_state.oracle_authority == oracle_authority.key() @ TerraVaultError::InvalidOracle,
        constraint = project_state.state == ProjectStatus::InMilestones @ TerraVaultError::InvalidProjectState,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,

    #[account(
        mut,
        seeds = [b"milestone", project_state.key().as_ref(), &[milestone_index]],
        bump = milestone_record.bump,
        constraint = milestone_record.milestone_index == project_state.current_milestone_index @ TerraVaultError::NotCurrentMilestone,
        constraint = milestone_record.status == MilestoneStatus::UnderReview @ TerraVaultError::InvalidMilestoneStatus,
    )]
    pub milestone_record: Account<'info, MilestoneRecord>,

    #[account(
        mut,
        seeds = [b"escrow", project_state.key().as_ref()],
        bump = project_state.escrow_bump,
    )]
    pub escrow_vault: InterfaceAccount<'info, TokenAccount>,

    #[account(mut)]
    pub developer_usdc_account: InterfaceAccount<'info, TokenAccount>,

    pub usdc_mint: InterfaceAccount<'info, Mint>,
    pub token_program: Program<'info, Token2022>,
}

pub fn handler(ctx: Context<ReleaseMilestoneFunds>, milestone_index: u8) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;
    let milestone = &mut ctx.accounts.milestone_record;

    // Enforce dispute window elapsed
    require!(
        clock.unix_timestamp >= milestone.dispute_deadline,
        TerraVaultError::DisputeWindowActive
    );

    // u128 intermediate to prevent overflow on large USDC amounts
    let release_amount: u64 = (project.total_raised_usdc as u128)
        .checked_mul(milestone.release_bps as u128)
        .ok_or(TerraVaultError::MathOverflow)?
        .checked_div(BASIS_POINTS_DIVISOR as u128)
        .ok_or(TerraVaultError::DivisionByZero)?
        .try_into()
        .map_err(|_| TerraVaultError::MathOverflow)?;

    // Update state BEFORE CPI
    milestone.status = MilestoneStatus::Released;
    milestone.released_amount_usdc = release_amount;

    project.escrow_balance_usdc = project
        .escrow_balance_usdc
        .checked_sub(release_amount)
        .ok_or(TerraVaultError::MathOverflow)?;
    project.milestones_completed = project
        .milestones_completed
        .checked_add(1)
        .ok_or(TerraVaultError::MathOverflow)?;
    project.current_milestone_index = project
        .current_milestone_index
        .checked_add(1)
        .ok_or(TerraVaultError::MathOverflow)?;
    project.updated_at = clock.unix_timestamp;

    let all_done = project.milestones_completed == project.milestone_count;
    if all_done {
        project.state = ProjectStatus::Completed;
    }

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
            to: ctx.accounts.developer_usdc_account.to_account_info(),
            authority: ctx.accounts.escrow_vault.to_account_info(),
        },
        signer_seeds,
    );
    transfer_checked(cpi_ctx, release_amount, 6)?;

    emit!(MilestoneFundsReleased {
        project: project.key(),
        milestone_index,
        amount_released: release_amount,
        developer: project.developer,
        timestamp: clock.unix_timestamp,
    });

    if all_done {
        emit!(ProjectCompleted {
            project: project.key(),
            developer: project.developer,
            timestamp: clock.unix_timestamp,
        });
    }

    Ok(())
}
