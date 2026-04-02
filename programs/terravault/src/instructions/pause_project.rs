use anchor_lang::prelude::*;
use crate::{
    errors::TerraVaultError,
    events::{ProjectPaused, ProjectUnpaused},
    state::{ProjectState, ProjectStatus},
};

#[derive(Accounts)]
pub struct PauseProject<'info> {
    pub developer: Signer<'info>,

    #[account(
        mut,
        has_one = developer @ TerraVaultError::InvalidDeveloper,
        constraint = project_state.state != ProjectStatus::Cancelled @ TerraVaultError::InvalidProjectState,
        constraint = project_state.state != ProjectStatus::Closed @ TerraVaultError::InvalidProjectState,
    )]
    pub project_state: Account<'info, ProjectState>,
}

// Separate struct required by Anchor — type alias does not generate __client_accounts_*
#[derive(Accounts)]
pub struct UnpauseProject<'info> {
    pub developer: Signer<'info>,

    #[account(
        mut,
        has_one = developer @ TerraVaultError::InvalidDeveloper,
        constraint = project_state.state != ProjectStatus::Cancelled @ TerraVaultError::InvalidProjectState,
        constraint = project_state.state != ProjectStatus::Closed @ TerraVaultError::InvalidProjectState,
    )]
    pub project_state: Account<'info, ProjectState>,
}

pub fn handler_pause(ctx: Context<PauseProject>) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;
    project.paused = true;
    project.updated_at = clock.unix_timestamp;
    emit!(ProjectPaused {
        project: project.key(),
        timestamp: clock.unix_timestamp,
    });
    Ok(())
}

pub fn handler_unpause(ctx: Context<UnpauseProject>) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;
    project.paused = false;
    project.updated_at = clock.unix_timestamp;
    emit!(ProjectUnpaused {
        project: project.key(),
        timestamp: clock.unix_timestamp,
    });
    Ok(())
}
