use anchor_lang::prelude::*;
use crate::{
    constants::BASIS_POINTS_DIVISOR,
    errors::TerraVaultError,
    state::{MilestoneRecord, MilestoneStatus, MilestoneType, ProjectState, ProjectStatus},
};

#[derive(AnchorSerialize, AnchorDeserialize, Clone)]
pub struct AddMilestoneParams {
    pub index: u8,
    pub milestone_type: MilestoneType,
    pub description: [u8; 64],
    pub release_bps: u16,
}

#[derive(Accounts)]
#[instruction(params: AddMilestoneParams)]
pub struct AddMilestone<'info> {
    #[account(mut)]
    pub developer: Signer<'info>,

    #[account(
        mut,
        has_one = developer @ TerraVaultError::InvalidDeveloper,
        constraint = project_state.state == ProjectStatus::Draft @ TerraVaultError::InvalidProjectState,
        constraint = !project_state.paused @ TerraVaultError::ProjectPaused,
    )]
    pub project_state: Account<'info, ProjectState>,

    #[account(
        init,
        payer = developer,
        space = MilestoneRecord::LEN,
        seeds = [b"milestone", project_state.key().as_ref(), &[params.index]],
        bump
    )]
    pub milestone_record: Account<'info, MilestoneRecord>,

    pub system_program: Program<'info, System>,
}

pub fn handler(ctx: Context<AddMilestone>, params: AddMilestoneParams) -> Result<()> {
    let project = &mut ctx.accounts.project_state;

    require!(
        params.index < project.milestone_count,
        TerraVaultError::MilestoneIndexOutOfBounds
    );
    // Enforce sequential addition
    require!(
        params.index == project.milestones_added,
        TerraVaultError::MilestoneIndexOutOfBounds
    );

    let new_total = project
        .milestone_bps_total
        .checked_add(params.release_bps)
        .ok_or(TerraVaultError::MathOverflow)?;
    require!(
        new_total <= BASIS_POINTS_DIVISOR,
        TerraVaultError::MilestoneBpsOverflow
    );

    let milestone = &mut ctx.accounts.milestone_record;
    milestone.project = project.key();
    milestone.milestone_index = params.index;
    milestone.milestone_type = params.milestone_type;
    milestone.description = params.description;
    milestone.release_bps = params.release_bps;
    milestone.status = MilestoneStatus::Pending;
    milestone.proof_uri = [0u8; 128];
    milestone.proof_hash = [0u8; 32];
    milestone.submitted_at = 0;
    milestone.approved_at = 0;
    milestone.released_amount_usdc = 0;
    milestone.oracle_signature = [0u8; 64];
    milestone.dispute_deadline = 0;
    milestone.bump = ctx.bumps.milestone_record;
    milestone.reserved = [0u8; 32];

    project.milestones_added = project
        .milestones_added
        .checked_add(1)
        .ok_or(TerraVaultError::MathOverflow)?;
    project.milestone_bps_total = new_total;

    Ok(())
}
