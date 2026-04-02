use anchor_lang::prelude::*;
use crate::{
    errors::TerraVaultError,
    events::EvidenceSubmitted,
    state::{ProjectState, ProjectStatus},
};

#[derive(Accounts)]
pub struct SubmitEvidence<'info> {
    /// Developer or any investor may submit evidence during the dispute window.
    pub submitter: Signer<'info>,

    #[account(
        mut,
        constraint = project_state.state == ProjectStatus::DisputeActive
            @ TerraVaultError::NotInDispute,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,
}

pub fn handler(ctx: Context<SubmitEvidence>, evidence_hash: [u8; 32]) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;

    require!(
        clock.unix_timestamp < project.dispute_deadline,
        TerraVaultError::DisputeExpired
    );

    project.evidence_hash = evidence_hash;

    emit!(EvidenceSubmitted {
        project: project.key(),
        submitted_by: ctx.accounts.submitter.key(),
        evidence_hash,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
