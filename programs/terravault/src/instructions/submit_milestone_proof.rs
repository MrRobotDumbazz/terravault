use anchor_lang::prelude::*;
use anchor_lang::solana_program::sysvar::instructions as ix_sysvar;
use crate::{
    constants::DISPUTE_WINDOW_SECONDS,
    errors::TerraVaultError,
    events::MilestoneProofSubmitted,
    state::{MilestoneRecord, MilestoneStatus, ProjectState, ProjectStatus},
};

#[derive(AnchorSerialize, AnchorDeserialize, Clone)]
pub struct SubmitMilestoneProofParams {
    pub milestone_index: u8,
    pub proof_uri: [u8; 128],
    pub proof_hash: [u8; 32],
    pub oracle_sig: [u8; 64],
}

#[derive(Accounts)]
#[instruction(params: SubmitMilestoneProofParams)]
pub struct SubmitMilestoneProof<'info> {
    pub oracle_authority: Signer<'info>,

    #[account(
        mut,
        constraint = project_state.oracle_authority == oracle_authority.key() @ TerraVaultError::InvalidOracle,
        constraint = (
            project_state.state == ProjectStatus::Active
            || project_state.state == ProjectStatus::InMilestones
        ) @ TerraVaultError::InvalidProjectState,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,

    #[account(
        mut,
        seeds = [b"milestone", project_state.key().as_ref(), &[params.milestone_index]],
        bump = milestone_record.bump,
        constraint = milestone_record.milestone_index == project_state.current_milestone_index @ TerraVaultError::NotCurrentMilestone,
        constraint = milestone_record.status == MilestoneStatus::Pending @ TerraVaultError::InvalidMilestoneStatus,
    )]
    pub milestone_record: Account<'info, MilestoneRecord>,

    /// CHECK: Solana instructions sysvar — used for Ed25519 instruction introspection
    #[account(address = ix_sysvar::ID)]
    pub instructions_sysvar: UncheckedAccount<'info>,
}

pub fn handler(ctx: Context<SubmitMilestoneProof>, params: SubmitMilestoneProofParams) -> Result<()> {
    let clock = Clock::get()?;

    require!(
        params.proof_hash != [0u8; 32],
        TerraVaultError::InvalidProofHash
    );

    // Production: verify preceding Ed25519 instruction signs
    // (project_pubkey || milestone_index || proof_hash) with oracle_authority key.
    //
    // let ix = ix_sysvar::load_instruction_at_checked(0, &ctx.accounts.instructions_sysvar.to_account_info())?;
    // assert ix.program_id == ed25519_program::ID
    // assert ix data matches expected payload + oracle_authority pubkey
    //
    // Deferred to Phase 2 hardening (see docs/security-audit-checklist.md)

    let project = &mut ctx.accounts.project_state;
    let milestone = &mut ctx.accounts.milestone_record;

    let dispute_deadline = clock.unix_timestamp + DISPUTE_WINDOW_SECONDS;

    milestone.proof_uri = params.proof_uri;
    milestone.proof_hash = params.proof_hash;
    milestone.oracle_signature = params.oracle_sig;
    milestone.submitted_at = clock.unix_timestamp;
    milestone.dispute_deadline = dispute_deadline;
    milestone.status = MilestoneStatus::UnderReview;

    if project.state == ProjectStatus::Active {
        project.state = ProjectStatus::InMilestones;
    }
    project.updated_at = clock.unix_timestamp;

    emit!(MilestoneProofSubmitted {
        project: project.key(),
        milestone_index: params.milestone_index,
        proof_uri: params.proof_uri,
        proof_hash: params.proof_hash,
        dispute_deadline,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
