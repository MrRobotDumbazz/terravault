use anchor_lang::prelude::*;

#[error_code]
pub enum TerraVaultError {
    // State errors
    #[msg("Invalid project state for this operation")]
    InvalidProjectState,
    #[msg("Project is paused")]
    ProjectPaused,

    // Fundraising errors
    #[msg("Fundraising deadline has not passed yet")]
    FundraisingDeadlineNotPassed,
    #[msg("Fundraising deadline has passed")]
    FundraisingDeadlinePassed,
    #[msg("Fundraising target not reached")]
    FundraisingTargetNotReached,
    #[msg("Hard cap exceeded")]
    HardCapExceeded,
    #[msg("Token amount must be greater than zero")]
    ZeroTokenAmount,
    #[msg("Insufficient fundraise deadline (min 7 days)")]
    InsufficientFundraiseDuration,

    // Milestone errors
    #[msg("Milestone index out of bounds")]
    MilestoneIndexOutOfBounds,
    #[msg("Not the current active milestone")]
    NotCurrentMilestone,
    #[msg("Milestone release basis points would exceed 10000")]
    MilestoneBpsOverflow,
    #[msg("Total milestone basis points must equal 10000 to start fundraising")]
    MilestoneBpsNotComplete,
    #[msg("Milestone is not in the expected status")]
    InvalidMilestoneStatus,
    #[msg("Dispute window has not elapsed")]
    DisputeWindowActive,
    #[msg("Dispute deadline has passed")]
    DisputeWindowExpired,
    #[msg("Not all milestones have been added")]
    MilestonesIncomplete,

    // KYC errors
    #[msg("KYC verification required to invest")]
    KycRequired,
    #[msg("Investor is not KYC verified")]
    KycNotVerified,

    // Oracle errors
    #[msg("Signer is not the oracle authority")]
    InvalidOracle,
    #[msg("Invalid oracle signature")]
    InvalidOracleSignature,
    #[msg("Oracle update timelock not elapsed")]
    OracleTimelockActive,

    // Distribution errors
    #[msg("No tokens held to claim distribution")]
    NoTokensHeld,
    #[msg("Distribution round already claimed")]
    DistributionAlreadyClaimed,
    #[msg("Must claim rounds in order")]
    DistributionRoundSkipped,
    #[msg("Distribution amount must be greater than zero")]
    ZeroDistributionAmount,

    // Token errors
    #[msg("Transfer fee basis points exceed maximum (500 bps)")]
    TransferFeeTooHigh,

    // Math errors
    #[msg("Arithmetic overflow")]
    MathOverflow,
    #[msg("Division by zero")]
    DivisionByZero,

    // Authority errors
    #[msg("Signer is not the developer")]
    InvalidDeveloper,
    #[msg("Signer is not the arbitration authority")]
    InvalidArbitrationAuthority,

    // Refund errors
    #[msg("No investment to refund")]
    NoInvestmentToRefund,

    // Misc
    #[msg("Proof hash cannot be all zeros")]
    InvalidProofHash,
    #[msg("Metadata URI too long")]
    MetadataUriTooLong,

    // Dispute errors
    #[msg("Project is not in an active disputable state")]
    NotDisputable,
    #[msg("Project is not in DisputeActive state")]
    NotInDispute,
    #[msg("Dispute resolution window has expired")]
    DisputeExpired,
    #[msg("Dispute resolution window is still open")]
    DisputeWindowStillOpen,
    #[msg("Signer is not the admin authority")]
    InvalidAdmin,

    // Blacklist errors
    #[msg("Developer is blacklisted and cannot create new projects")]
    DeveloperBlacklisted,
}
