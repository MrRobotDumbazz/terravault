use anchor_lang::prelude::*;
use crate::{
    errors::TerraVaultError,
    events::FundraiseCancelled,
    state::{ProjectState, ProjectStatus},
};

#[derive(Accounts)]
pub struct CancelFundraise<'info> {
    pub developer: Signer<'info>,

    #[account(
        mut,
        has_one = developer @ TerraVaultError::InvalidDeveloper,
        constraint = project_state.state == ProjectStatus::Fundraising @ TerraVaultError::InvalidProjectState,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,
}

pub fn handler(ctx: Context<CancelFundraise>) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;

    let deadline_passed = clock.unix_timestamp >= project.fundraise_deadline;
    let target_met = project.total_raised_usdc >= project.fundraise_target_usdc;

    // Cannot cancel if target is met and deadline hasn't passed yet
    // (investors have already committed funds past the soft cap)
    require!(
        !(target_met && !deadline_passed),
        TerraVaultError::FundraisingTargetNotReached
    );

    project.state = ProjectStatus::Cancelled;
    project.updated_at = clock.unix_timestamp;

    let mut reason = [0u8; 64];
    let msg: &[u8] = if deadline_passed && !target_met {
        b"Deadline passed, target not reached"
    } else {
        b"Cancelled by developer"
    };
    reason[..msg.len()].copy_from_slice(msg);

    emit!(FundraiseCancelled {
        project: project.key(),
        total_raised: project.total_raised_usdc,
        reason,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
