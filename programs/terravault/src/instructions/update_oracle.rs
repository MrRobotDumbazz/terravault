use anchor_lang::prelude::*;
use crate::{
    constants::ORACLE_UPDATE_TIMELOCK_SECONDS,
    errors::TerraVaultError,
    events::OracleUpdated,
    state::{ProjectState, ProjectStatus},
};

#[derive(Accounts)]
pub struct UpdateOracle<'info> {
    pub developer: Signer<'info>,

    #[account(
        mut,
        has_one = developer @ TerraVaultError::InvalidDeveloper,
        // Block oracle rotation while milestone proof is under active review
        constraint = project_state.state != ProjectStatus::InMilestones @ TerraVaultError::InvalidProjectState,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,
}

pub fn handler(ctx: Context<UpdateOracle>, new_oracle: Pubkey) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;

    if project.oracle_update_pending_at > 0 {
        // A rotation is already in progress — check if timelock has elapsed
        require!(
            clock.unix_timestamp
                >= project.oracle_update_pending_at + ORACLE_UPDATE_TIMELOCK_SECONDS,
            TerraVaultError::OracleTimelockActive
        );

        let old_oracle = project.oracle_authority;
        project.oracle_authority = project.pending_oracle;
        project.pending_oracle = Pubkey::default();
        project.oracle_update_pending_at = 0;
        project.updated_at = clock.unix_timestamp;

        emit!(OracleUpdated {
            project: project.key(),
            old_oracle,
            new_oracle: project.oracle_authority,
            effective_at: clock.unix_timestamp,
            timestamp: clock.unix_timestamp,
        });
    } else {
        // Initiate a new rotation — starts the 48h timelock
        project.pending_oracle = new_oracle;
        project.oracle_update_pending_at = clock.unix_timestamp;
        project.updated_at = clock.unix_timestamp;
    }

    Ok(())
}
