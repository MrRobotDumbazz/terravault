use anchor_lang::prelude::*;
use crate::{
    errors::TerraVaultError,
    events::AdminResolved,
    state::{AdminDecision, DeveloperProfile, ProjectState, ProjectStatus},
};

/// Admin multisig pubkey — replace with actual 2/3 multisig address before mainnet.
/// On devnet this is a single authority key set at deploy time via env.
pub const ADMIN_PUBKEY: Pubkey = solana_program::pubkey!("11111111111111111111111111111111");

#[derive(AnchorSerialize, AnchorDeserialize, Clone, PartialEq, Eq, Debug)]
pub enum ResolutionDecision {
    PayInvestors,
    RefundAndExtend,
    ForceClose,
}

#[derive(Accounts)]
pub struct AdminResolve<'info> {
    /// Must be the protocol admin (multisig in production).
    #[account(
        constraint = admin.key() == ADMIN_PUBKEY @ TerraVaultError::InvalidAdmin,
    )]
    pub admin: Signer<'info>,

    #[account(
        mut,
        constraint = project_state.state == ProjectStatus::DisputeActive
            @ TerraVaultError::NotInDispute,
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
    ctx: Context<AdminResolve>,
    decision: ResolutionDecision,
) -> Result<()> {
    let clock = Clock::get()?;
    let project = &mut ctx.accounts.project_state;
    let profile = &mut ctx.accounts.developer_profile;

    let admin_decision = match decision {
        ResolutionDecision::PayInvestors => {
            profile.disputes_lost = profile.disputes_lost.saturating_add(1);
            AdminDecision::PayInvestors
        }
        ResolutionDecision::RefundAndExtend => AdminDecision::RefundAndExtend,
        ResolutionDecision::ForceClose => AdminDecision::ForceClose,
    };

    project.admin_decision = admin_decision.clone();
    project.state = ProjectStatus::Resolved;

    let decision_byte: u8 = match admin_decision {
        AdminDecision::PayInvestors => 1,
        AdminDecision::RefundAndExtend => 2,
        AdminDecision::ForceClose => 3,
        AdminDecision::None => 0,
    };

    emit!(AdminResolved {
        project: project.key(),
        decision: decision_byte,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
