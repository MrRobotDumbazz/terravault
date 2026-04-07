use anchor_lang::prelude::*;

pub mod constants;
pub mod errors;
pub mod events;
pub mod instructions;
pub mod state;

use instructions::initialize_project::*;
use instructions::add_milestone::*;
use instructions::start_fundraising::*;
use instructions::buy_tokens::*;
use instructions::cancel_fundraise::*;
use instructions::claim_refund::*;
use instructions::activate_project::*;
use instructions::submit_milestone_proof::*;
use instructions::dispute_milestone::*;
use instructions::resolve_dispute::*;
use instructions::release_milestone_funds::*;
use instructions::deposit_income::*;
use instructions::claim_distribution::*;
use instructions::update_oracle::*;
use instructions::pause_project::*;
use instructions::transfer_hook::*;
use instructions::raise_dispute::*;
use instructions::submit_evidence::*;
use instructions::admin_resolve::*;
use instructions::freeze_developer::*;
use instructions::flag_developer::*;
use state::distribution::DistributionSource;

declare_id!("FpknHE4LQu9etCmFgcmJmS3hURrrSNiDNs2qxasSmAT8");

#[program]
pub mod terravault {
    use super::*;

    /// Create a new real estate project (Residential / Commercial / Agricultural).
    /// Initialises ProjectState PDA, TokenConfig PDA, and USDC escrow vault.
    /// Project starts in Draft state; milestones must be added before fundraising.
    pub fn initialize_project(
        ctx: Context<InitializeProject>,
        params: InitializeProjectParams,
    ) -> Result<()> {
        instructions::initialize_project::handler(ctx, params)
    }

    /// Add a milestone to a Draft project.
    /// Milestones must be added sequentially (index 0, 1, 2 …).
    /// Sum of all milestone release_bps must equal 10_000 before start_fundraising.
    pub fn add_milestone(
        ctx: Context<AddMilestone>,
        params: AddMilestoneParams,
    ) -> Result<()> {
        instructions::add_milestone::handler(ctx, params)
    }

    /// Transition project from Draft → Fundraising.
    /// Requires all milestones added and total release_bps == 10_000.
    pub fn start_fundraising(ctx: Context<StartFundraising>) -> Result<()> {
        instructions::start_fundraising::handler(ctx)
    }

    /// Investor purchases fractional tokens with USDC.
    /// USDC is held in escrow vault; tokens represent ownership shares.
    pub fn buy_tokens(ctx: Context<BuyTokens>, token_amount: u64) -> Result<()> {
        instructions::buy_tokens::handler(ctx, token_amount)
    }

    /// Cancel fundraising. Allowed when soft cap not met, or developer explicit cancel
    /// before target is reached. After cancel, investors can call claim_refund.
    pub fn cancel_fundraise(ctx: Context<CancelFundraise>) -> Result<()> {
        instructions::cancel_fundraise::handler(ctx)
    }

    /// Investor reclaims USDC after fundraise is cancelled.
    pub fn claim_refund(ctx: Context<ClaimRefund>) -> Result<()> {
        instructions::claim_refund::handler(ctx)
    }

    /// Transition Fundraising → Active once soft cap is reached.
    /// Tokens become transferable on secondary markets after activation.
    pub fn activate_project(ctx: Context<ActivateProject>) -> Result<()> {
        instructions::activate_project::handler(ctx)
    }

    /// Oracle submits construction milestone proof.
    /// Opens a 48-hour dispute window before funds can be released.
    pub fn submit_milestone_proof(
        ctx: Context<SubmitMilestoneProof>,
        params: SubmitMilestoneProofParams,
    ) -> Result<()> {
        instructions::submit_milestone_proof::handler(ctx, params)
    }

    /// Token holder disputes a milestone proof during the dispute window.
    pub fn dispute_milestone(
        ctx: Context<DisputeMilestone>,
        milestone_index: u8,
        reason: [u8; 128],
    ) -> Result<()> {
        instructions::dispute_milestone::handler(ctx, milestone_index, reason)
    }

    /// Arbitration authority resolves a disputed milestone.
    /// approved=true: re-opens for release. approved=false: developer must resubmit.
    pub fn resolve_dispute(
        ctx: Context<ResolveDispute>,
        milestone_index: u8,
        approved: bool,
    ) -> Result<()> {
        instructions::resolve_dispute::handler(ctx, milestone_index, approved)
    }

    /// Oracle releases escrowed USDC to developer after dispute window elapses.
    /// Final milestone transitions project → Completed.
    pub fn release_milestone_funds(
        ctx: Context<ReleaseMilestoneFunds>,
        milestone_index: u8,
    ) -> Result<()> {
        instructions::release_milestone_funds::handler(ctx, milestone_index)
    }

    /// Developer deposits rental income or sale proceeds for proportional distribution.
    /// Creates a DistributionPool with a snapshot of circulating token supply.
    pub fn deposit_income(
        ctx: Context<DepositIncome>,
        amount_usdc: u64,
        source: DistributionSource,
    ) -> Result<()> {
        instructions::deposit_income::handler(ctx, amount_usdc, source)
    }

    /// Token holder claims their proportional share of a distribution round.
    /// Rounds must be claimed sequentially to prevent skip-and-claim exploits.
    pub fn claim_distribution(ctx: Context<ClaimDistribution>, round: u32) -> Result<()> {
        instructions::claim_distribution::handler(ctx, round)
    }

    /// Initiate a 48-hour timelock oracle key rotation.
    /// Call again after timelock to finalise the rotation.
    pub fn update_oracle(ctx: Context<UpdateOracle>, new_oracle: Pubkey) -> Result<()> {
        instructions::update_oracle::handler(ctx, new_oracle)
    }

    /// Emergency pause — freezes all state-mutating operations.
    pub fn pause_project(ctx: Context<PauseProject>) -> Result<()> {
        instructions::pause_project::handler_pause(ctx)
    }

    /// Remove emergency pause.
    pub fn unpause_project(ctx: Context<UnpauseProject>) -> Result<()> {
        instructions::pause_project::handler_unpause(ctx)
    }

    /// Token-2022 TransferHook entry point.
    /// Called automatically by the Token-2022 program on every project token transfer.
    /// Updates InvestorPosition.tokens_held for both parties.
    pub fn execute_transfer_hook(
        ctx: Context<ExecuteTransferHook>,
        amount: u64,
    ) -> Result<()> {
        instructions::transfer_hook::handler(ctx, amount)
    }

    /// Raise a project-level dispute. Transitions Active/InMilestones → DisputeActive.
    /// Starts 72-hour resolution window.
    pub fn raise_dispute(
        ctx: Context<RaiseDispute>,
        reason_hash: [u8; 32],
    ) -> Result<()> {
        instructions::raise_dispute::handler(ctx, reason_hash)
    }

    /// Submit IPFS evidence hash during the 72-hour dispute window.
    /// Either party (developer or investor) may call this.
    pub fn submit_evidence(
        ctx: Context<SubmitEvidence>,
        evidence_hash: [u8; 32],
    ) -> Result<()> {
        instructions::submit_evidence::handler(ctx, evidence_hash)
    }

    /// Admin multisig resolves a project dispute.
    /// Decisions: PayInvestors | RefundAndExtend | ForceClose.
    /// Transitions DisputeActive → Resolved.
    pub fn admin_resolve(
        ctx: Context<AdminResolve>,
        decision: ResolutionDecision,
    ) -> Result<()> {
        instructions::admin_resolve::handler(ctx, decision)
    }

    /// Freeze the developer's token account after PayInvestors decision.
    /// Uses Token-2022 freeze authority held by the project_state PDA.
    pub fn freeze_developer(ctx: Context<FreezeDeveloper>) -> Result<()> {
        instructions::freeze_developer::handler(ctx)
    }

    /// Flag a developer as a bad actor. Sets is_blacklisted = true on DeveloperProfile.
    /// Prevents future project creation by this developer.
    pub fn flag_developer(
        ctx: Context<FlagDeveloper>,
        developer_pubkey: Pubkey,
        reason_hash: [u8; 32],
    ) -> Result<()> {
        instructions::flag_developer::handler(ctx, developer_pubkey, reason_hash)
    }
}
