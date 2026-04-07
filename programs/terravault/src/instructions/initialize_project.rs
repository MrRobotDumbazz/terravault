use anchor_lang::prelude::*;
use anchor_spl::{
    associated_token::AssociatedToken,
    token_2022::Token2022,
    token_interface::{TokenAccount},
};
use crate::{
    constants::*,
    errors::TerraVaultError,
    events::ProjectCreated,
    state::{DeveloperProfile, ProjectState, ProjectStatus, ProjectType, TokenConfig},
};

#[derive(AnchorSerialize, AnchorDeserialize, Clone)]
pub struct InitializeProjectParams {
    pub project_id: u64,
    pub project_type: ProjectType,
    pub total_tokens: u64,
    pub token_price_usdc: u64,
    pub fundraise_target_usdc: u64,
    pub fundraise_hard_cap_usdc: u64,
    pub fundraise_deadline: i64,
    pub milestone_count: u8,
    pub metadata_uri: Vec<u8>,
    pub legal_doc_hash: [u8; 32],
    pub oracle_authority: Pubkey,
    pub kyc_required: bool,
    pub transfer_fee_bps: u16,
    pub fee_destination: Pubkey,
}

#[derive(Accounts)]
#[instruction(params: InitializeProjectParams)]
pub struct InitializeProject<'info> {
    #[account(mut)]
    pub developer: Signer<'info>,

    #[account(
        init,
        payer = developer,
        space = ProjectState::LEN,
        seeds = [b"project", developer.key().as_ref(), &params.project_id.to_le_bytes()],
        bump
    )]
    pub project_state: Account<'info, ProjectState>,

    #[account(
        init,
        payer = developer,
        space = TokenConfig::LEN,
        seeds = [b"token_config", project_state.key().as_ref()],
        bump
    )]
    pub token_config: Account<'info, TokenConfig>,

    /// CHECK: USDC mint — validated off-chain; used only to type the escrow vault
    pub usdc_mint: UncheckedAccount<'info>,

    #[account(
        init,
        payer = developer,
        token::mint = usdc_mint,
        token::authority = project_state,
        token::token_program = token_program,
        seeds = [b"escrow", project_state.key().as_ref()],
        bump
    )]
    pub escrow_vault: Box<InterfaceAccount<'info, TokenAccount>>,

    /// Developer profile PDA — created on first project, checked for blacklist on subsequent ones.
    #[account(
        init_if_needed,
        payer = developer,
        space = DeveloperProfile::LEN,
        seeds = [DeveloperProfile::SEED, developer.key().as_ref()],
        bump
    )]
    pub developer_profile: Account<'info, DeveloperProfile>,

    pub token_program: Program<'info, Token2022>,
    pub associated_token_program: Program<'info, AssociatedToken>,
    pub system_program: Program<'info, System>,
    pub rent: Sysvar<'info, Rent>,
}

pub fn handler(ctx: Context<InitializeProject>, params: InitializeProjectParams) -> Result<()> {
    let clock = Clock::get()?;

    // Blacklist check — reject if developer was previously flagged via flag_developer.
    require!(
        !ctx.accounts.developer_profile.is_blacklisted,
        TerraVaultError::DeveloperBlacklisted
    );

    require!(params.total_tokens > 0, TerraVaultError::ZeroTokenAmount);
    require!(params.token_price_usdc > 0, TerraVaultError::ZeroTokenAmount);
    require!(
        params.fundraise_hard_cap_usdc >= params.fundraise_target_usdc,
        TerraVaultError::HardCapExceeded
    );
    require!(
        params.fundraise_deadline > clock.unix_timestamp + MIN_FUNDRAISE_DEADLINE_SECONDS,
        TerraVaultError::InsufficientFundraiseDuration
    );
    require!(
        params.milestone_count > 0 && params.milestone_count <= MAX_MILESTONES,
        TerraVaultError::MilestonesIncomplete
    );
    require!(
        params.transfer_fee_bps <= MAX_TRANSFER_FEE_BPS,
        TerraVaultError::TransferFeeTooHigh
    );
    require!(
        params.legal_doc_hash != [0u8; 32],
        TerraVaultError::InvalidProofHash
    );
    require!(
        params.metadata_uri.len() <= MAX_METADATA_URI_LEN,
        TerraVaultError::MetadataUriTooLong
    );

    let bumps = ctx.bumps;
    let project = &mut ctx.accounts.project_state;

    let mut metadata_uri = [0u8; 128];
    metadata_uri[..params.metadata_uri.len()].copy_from_slice(&params.metadata_uri);

    project.project_id = params.project_id;
    project.developer = ctx.accounts.developer.key();
    project.oracle_authority = params.oracle_authority;
    project.token_mint = Pubkey::default(); // populated when SPL Token-2022 mint is created
    project.escrow_vault = ctx.accounts.escrow_vault.key();
    project.state = ProjectStatus::Draft;
    project.project_type = params.project_type;
    project.total_tokens = params.total_tokens;
    project.tokens_sold = 0;
    project.token_price_usdc = params.token_price_usdc;
    project.fundraise_target_usdc = params.fundraise_target_usdc;
    project.fundraise_hard_cap_usdc = params.fundraise_hard_cap_usdc;
    project.fundraise_deadline = params.fundraise_deadline;
    project.escrow_balance_usdc = 0;
    project.total_raised_usdc = 0;
    project.milestone_count = params.milestone_count;
    project.milestones_added = 0;
    project.milestones_completed = 0;
    project.current_milestone_index = 0;
    project.milestone_bps_total = 0;
    project.distribution_round = 0;
    project.total_distributed_usdc = 0;
    project.metadata_uri = metadata_uri;
    project.legal_doc_hash = params.legal_doc_hash;
    project.oracle_update_pending_at = 0;
    project.pending_oracle = Pubkey::default();
    project.created_at = clock.unix_timestamp;
    project.updated_at = clock.unix_timestamp;
    project.bump = bumps.project_state;
    project.escrow_bump = bumps.escrow_vault;
    project.paused = false;
    project.kyc_required = params.kyc_required;
    project.transfer_fee_bps = params.transfer_fee_bps;
    project.reserved = [0u8; 23];

    // Initialise DeveloperProfile on first project (init_if_needed zeroes it, so only set
    // non-zero fields when developer field is default/empty).
    let dev_profile = &mut ctx.accounts.developer_profile;
    if dev_profile.developer == Pubkey::default() {
        dev_profile.developer = ctx.accounts.developer.key();
        dev_profile.bump = bumps.developer_profile;
    }
    dev_profile.total_projects = dev_profile.total_projects.saturating_add(1);

    let token_config = &mut ctx.accounts.token_config;
    token_config.project = project.key();
    token_config.mint = Pubkey::default();
    token_config.transfer_fee_bps = params.transfer_fee_bps;
    token_config.transfer_fee_max_lamports = u64::MAX;
    token_config.fee_destination = params.fee_destination;
    token_config.interest_rate_bps = 0;
    token_config.bump = bumps.token_config;

    emit!(ProjectCreated {
        project: project.key(),
        developer: ctx.accounts.developer.key(),
        project_id: params.project_id,
        total_tokens: params.total_tokens,
        token_price_usdc: params.token_price_usdc,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
