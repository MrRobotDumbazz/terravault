use anchor_lang::prelude::*;
use crate::{
    constants::DAO_DISPUTE_DEADLINE_SECONDS,
    errors::TerraVaultError,
    events::DisputeRaised,
    state::{DeveloperProfile, ProjectState, ProjectStatus},
};

#[derive(Accounts)]
pub struct RaiseDispute<'info> {
    /// Either the developer or any token holder may raise a dispute.
    pub caller: Signer<'info>,

    #[account(
        mut,
        constraint = matches!(
            project_state.state,
            ProjectStatus::Active | ProjectStatus::InMilestones
        ) @ TerraVaultError::NotDisputable,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,

    #[account(
        mut,
        seeds = [DeveloperProfile::SEED, project_state.developer.as_ref()],
        bump = developer_profile.bump,
    )]
    pub developer_profile: Account<'info, DeveloperProfile>,
}

pub fn handler(
    ctx: Context<RaiseDispute>,
    reason_hash: [u8; 32],
) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;

    project.state = ProjectStatus::DisputeActive;
    project.dispute_deadline = clock.unix_timestamp + DAO_DISPUTE_DEADLINE_SECONDS;
    project.evidence_hash = reason_hash;

    // Track disputes on developer profile
    let profile = &mut ctx.accounts.developer_profile;
    profile.disputes_raised = profile.disputes_raised.saturating_add(1);

    emit!(DisputeRaised {
        project: project.key(),
        raised_by: ctx.accounts.caller.key(),
        reason_hash,
        deadline: project.dispute_deadline,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
