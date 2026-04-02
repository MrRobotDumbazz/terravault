use anchor_lang::prelude::*;

/// Per-developer reputation and blacklist record.
/// PDA seeds: [b"developer_profile", developer_pubkey]
/// Created on first project init; updated on milestone/dispute events.
#[account]
#[derive(Default)]
pub struct DeveloperProfile {
    pub developer: Pubkey,       // 32
    pub total_projects: u32,     // 4
    pub completed_on_time: u32,  // 4
    pub disputes_raised: u32,    // 4  — incremented when raise_dispute is called
    pub disputes_lost: u32,      // 4  — incremented when admin_resolve → PayInvestors
    pub is_blacklisted: bool,    // 1
    pub rating_score: u8,        // 1  — V2: (completed_on_time/total*100) - (lost*20)
    pub bump: u8,                // 1
    pub reserved: [u8; 29],      // pad to 80 bytes total (+ 8 discriminator)
}

impl DeveloperProfile {
    pub const LEN: usize = 8   // discriminator
        + 32  // developer
        + 4   // total_projects
        + 4   // completed_on_time
        + 4   // disputes_raised
        + 4   // disputes_lost
        + 1   // is_blacklisted
        + 1   // rating_score
        + 1   // bump
        + 29; // reserved → total = 88

    pub const SEED: &'static [u8] = b"developer_profile";
}
