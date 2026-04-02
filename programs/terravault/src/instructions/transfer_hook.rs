use anchor_lang::prelude::*;
use anchor_spl::token_2022::Token2022;
use anchor_spl::token_interface::TokenAccount;
use crate::{
    errors::TerraVaultError,
    events::TokensTransferred,
    state::InvestorPosition,
};

/// Token-2022 TransferHook — called automatically on every project token transfer.
/// Updates InvestorPosition.tokens_held for sender and receiver to maintain
/// an accurate on-chain ledger for proportional income distribution.
#[derive(Accounts)]
pub struct ExecuteTransferHook<'info> {
    /// CHECK: Source token account — provided and validated by Token-2022
    pub source_token: InterfaceAccount<'info, TokenAccount>,

    /// CHECK: Project token mint
    pub mint: InterfaceAccount<'info, anchor_spl::token_interface::Mint>,

    /// CHECK: Destination token account — provided and validated by Token-2022
    pub destination_token: InterfaceAccount<'info, TokenAccount>,

    /// CHECK: Token owner / transfer authority
    pub authority: UncheckedAccount<'info>,

    /// CHECK: Required extra-account-metas PDA by transfer hook interface
    #[account(
        seeds = [b"extra-account-metas", mint.key().as_ref()],
        bump
    )]
    pub extra_account_meta_list: UncheckedAccount<'info>,

    /// Sender's InvestorPosition — optional because sender may not have a position
    /// (e.g. the program itself during mint/burn operations)
    #[account(mut)]
    pub source_position: Option<Account<'info, InvestorPosition>>,

    /// Receiver's InvestorPosition — optional for same reason
    #[account(mut)]
    pub destination_position: Option<Account<'info, InvestorPosition>>,

    pub token_program: Program<'info, Token2022>,
    pub system_program: Program<'info, System>,
}

pub fn handler(ctx: Context<ExecuteTransferHook>, amount: u64) -> Result<()> {
    let clock = Clock::get()?;

    if let Some(src) = ctx.accounts.source_position.as_mut() {
        src.tokens_held = src
            .tokens_held
            .checked_sub(amount)
            .ok_or(TerraVaultError::MathOverflow)?;
    }

    if let Some(dst) = ctx.accounts.destination_position.as_mut() {
        dst.tokens_held = dst
            .tokens_held
            .checked_add(amount)
            .ok_or(TerraVaultError::MathOverflow)?;
    }

    emit!(TokensTransferred {
        project: ctx.accounts.mint.key(),
        from: ctx.accounts.source_token.owner,
        to: ctx.accounts.destination_token.owner,
        amount,
        timestamp: clock.unix_timestamp,
    });

    Ok(())
}
