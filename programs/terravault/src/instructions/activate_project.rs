use anchor_lang::prelude::*;
use crate::{
    errors::TerraVaultError,
    events::ProjectActivated,
    state::{ProjectState, ProjectStatus},
};

#[derive(Accounts)]
pub struct ActivateProject<'info> {
    pub developer: Signer<'info>,

    #[account(
        mut,
        has_one = developer @ TerraVaultError::InvalidDeveloper,
        constraint = project_state.state == ProjectStatus::Fundraising @ TerraVaultError::InvalidProjectState,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,
}

pub fn handler(ctx: Context<ActivateProject>) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;

    require!(
        project.total_raised_usdc >= project.fundraise_target_usdc,
        TerraVaultError::FundraisingTargetNotReached
    );

    project.state = ProjectStatus::Active;
    project.current_milestone_index = 0;
    project.updated_at = clock.unix_timestamp;

    emit!(ProjectActivated {
        project: project.key(),
        total_raised: project.total_raised_usdc,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
