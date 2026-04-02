use anchor_lang::prelude::*;
use crate::{
    constants::BASIS_POINTS_DIVISOR,
    errors::TerraVaultError,
    events::FundraisingStarted,
    state::{ProjectState, ProjectStatus},
};

#[derive(Accounts)]
pub struct StartFundraising<'info> {
    pub developer: Signer<'info>,

    #[account(
        mut,
        has_one = developer @ TerraVaultError::InvalidDeveloper,
        constraint = project_state.state == ProjectStatus::Draft @ TerraVaultError::InvalidProjectState,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,
}

pub fn handler(ctx: Context<StartFundraising>) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;

    require!(
        project.milestones_added == project.milestone_count,
        TerraVaultError::MilestonesIncomplete
    );
    require!(
        project.milestone_bps_total == BASIS_POINTS_DIVISOR,
        TerraVaultError::MilestoneBpsNotComplete
    );

    project.state = ProjectStatus::Fundraising;
    project.updated_at = clock.unix_timestamp;

    emit!(FundraisingStarted {
        project: project.key(),
        fundraise_target: project.fundraise_target_usdc,
        hard_cap: project.fundraise_hard_cap_usdc,
        deadline: project.fundraise_deadline,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
