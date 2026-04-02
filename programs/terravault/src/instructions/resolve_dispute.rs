use anchor_lang::prelude::*;
use crate::{
    errors::TerraVaultError,
    events::DisputeResolved,
    state::{MilestoneRecord, MilestoneStatus, ProjectState},
};

#[derive(Accounts)]
#[instruction(milestone_index: u8)]
pub struct ResolveDispute<'info> {
    /// Arbitration authority — multisig in production. Set per-project at init
    /// or use a protocol-wide governance key stored in a separate ProtocolConfig PDA.
    pub arbitration_authority: Signer<'info>,

    #[account(mut)]
    pub project_state: Account<'info, ProjectState>,

    #[account(
        mut,
        seeds = [b"milestone", project_state.key().as_ref(), &[milestone_index]],
        bump = milestone_record.bump,
        constraint = milestone_record.status == MilestoneStatus::Disputed @ TerraVaultError::InvalidMilestoneStatus,
    )]
    pub milestone_record: Account<'info, MilestoneRecord>,
}

pub fn handler(
    ctx: Context<ResolveDispute>,
    milestone_index: u8,
    approved: bool,
) -> Result<()> {
    let clock = Clock::get()?;
    let milestone = &mut ctx.accounts.milestone_record;

    if approved {
        // Approved: reset dispute_deadline to now so release can proceed immediately
        milestone.status = MilestoneStatus::UnderReview;
        milestone.approved_at = clock.unix_timestamp;
        milestone.dispute_deadline = clock.unix_timestamp;
    } else {
        // Rejected: developer must resubmit proof
        milestone.status = MilestoneStatus::Pending;
        milestone.proof_uri = [0u8; 128];
        milestone.proof_hash = [0u8; 32];
        milestone.oracle_signature = [0u8; 64];
        milestone.submitted_at = 0;
        milestone.dispute_deadline = 0;
    }

    emit!(DisputeResolved {
        project: ctx.accounts.project_state.key(),
        milestone_index,
        approved,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
