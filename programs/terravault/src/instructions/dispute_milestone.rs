use anchor_lang::prelude::*;
use crate::{
    errors::TerraVaultError,
    events::MilestoneDisputed,
    state::{InvestorPosition, MilestoneRecord, MilestoneStatus, ProjectState, ProjectStatus},
};

#[derive(Accounts)]
#[instruction(milestone_index: u8)]
pub struct DisputeMilestone<'info> {
    pub investor: Signer<'info>,

    #[account(
        constraint = project_state.state == ProjectStatus::InMilestones @ TerraVaultError::InvalidProjectState,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,

    #[account(
        seeds = [b"position", project_state.key().as_ref(), investor.key().as_ref()],
        bump = investor_position.bump,
        constraint = investor_position.tokens_held > 0 @ TerraVaultError::NoTokensHeld,
    )]
    pub investor_position: Account<'info, InvestorPosition>,

    #[account(
        mut,
        seeds = [b"milestone", project_state.key().as_ref(), &[milestone_index]],
        bump = milestone_record.bump,
        constraint = milestone_record.milestone_index == project_state.current_milestone_index @ TerraVaultError::NotCurrentMilestone,
        constraint = milestone_record.status == MilestoneStatus::UnderReview @ TerraVaultError::InvalidMilestoneStatus,
    )]
    pub milestone_record: Account<'info, MilestoneRecord>,
}

pub fn handler(
    ctx: Context<DisputeMilestone>,
    milestone_index: u8,
    _reason: [u8; 128],
) -> Result<()> {
    let clock = Clock::get()?;
    let milestone = &mut ctx.accounts.milestone_record;

    require!(
        clock.unix_timestamp < milestone.dispute_deadline,
        TerraVaultError::DisputeWindowExpired
    );

    milestone.status = MilestoneStatus::Disputed;

    emit!(MilestoneDisputed {
        project: ctx.accounts.project_state.key(),
        milestone_index,
        investor: ctx.accounts.investor.key(),
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
