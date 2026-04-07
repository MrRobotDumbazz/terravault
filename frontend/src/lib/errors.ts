// Anchor program error codes start at 6000 (0x1770) for custom errors.
// TerraVaultError enum order matches index offset from 6000.
const PROGRAM_ERRORS: Record<number, string> = {
  6000: "Invalid project state for this operation.",
  6001: "This project is currently paused.",
  6002: "Fundraising deadline has not passed yet.",
  6003: "Fundraising deadline has passed — you can no longer invest.",
  6004: "Fundraising target not reached.",
  6005: "Hard cap exceeded — the maximum raise amount has been hit.",
  6006: "Token amount must be greater than zero.",
  6007: "Fundraise deadline is too short (minimum 7 days).",
  6008: "Milestone index is out of bounds.",
  6009: "This is not the current active milestone.",
  6010: "Milestone basis points would exceed 100%.",
  6011: "Total milestone basis points must equal 100% before fundraising can start.",
  6012: "Milestone is not in the expected status.",
  6013: "Dispute window has not elapsed yet — wait 48 hours before funds can be released.",
  6014: "Dispute window has expired.",
  6015: "Not all milestones have been added.",
  6016: "KYC verification is required to invest in this project.",
  6017: "Your wallet is not KYC verified. Please complete verification first.",
  6018: "Only the oracle authority can call this instruction.",
  6019: "Invalid oracle signature.",
  6020: "Oracle key rotation timelock has not elapsed (48 hours required).",
  6021: "You have no tokens in this project to claim a distribution.",
  6022: "You have already claimed this distribution round.",
  6023: "You must claim distribution rounds in order — claim earlier rounds first.",
  6024: "Distribution amount must be greater than zero.",
  6025: "Transfer fee exceeds the 5% maximum.",
  6026: "Arithmetic overflow — amount too large.",
  6027: "Division by zero error.",
  6028: "Only the developer can call this instruction.",
  6029: "Only the arbitration authority can resolve disputes.",
  6030: "No investment found to refund.",
  6031: "Proof hash cannot be empty.",
  6032: "Metadata URI is too long (max 128 bytes).",
  6033: "This project cannot be disputed in its current state.",
  6034: "This project is not in an active dispute.",
  6035: "Dispute resolution window has expired.",
  6036: "Dispute window is still open — wait for it to close before resolving.",
  6037: "Only the admin multisig can call this instruction.",
  6038: "This developer is blacklisted and cannot create new projects.",
};

/**
 * Extract a human-readable error message from any Solana/Anchor/wallet error.
 */
export function parseTransactionError(err: unknown): string {
  const raw = err instanceof Error ? err.message : String(err);

  // 1. User rejected in wallet
  if (/user rejected/i.test(raw) || /declined/i.test(raw) || /rejected the request/i.test(raw)) {
    return "Transaction cancelled — you rejected the request in your wallet.";
  }

  // 1b. Generic WalletSendTransactionError "Unexpected error" — almost always
  //     means the program is not deployed or simulation failed silently.
  if (/WalletSendTransactionError/i.test(raw) || /unexpected error/i.test(raw)) {
    return "Transaction failed — the program may not be deployed on this network yet, or your account is missing required token accounts. Make sure you are on devnet and the program is deployed.";
  }

  // 2. Not enough SOL for fees
  if (/insufficient.*(lamport|sol|fee)/i.test(raw) || /0x1\b/.test(raw)) {
    return "Insufficient SOL balance to pay transaction fees. Add SOL to your wallet.";
  }

  // 3. Not enough USDC / token balance
  if (/insufficient funds/i.test(raw) || /insufficient.*token/i.test(raw)) {
    return "Insufficient USDC balance. Make sure you have enough USDC in your wallet.";
  }

  // 4. Wallet not connected
  if (/wallet.*not.*connect/i.test(raw) || /no.*wallet/i.test(raw)) {
    return "Wallet is not connected. Please connect your wallet first.";
  }

  // 5. Blockhash expired / network timeout
  if (/blockhash.*not found/i.test(raw) || /block height exceeded/i.test(raw)) {
    return "Transaction expired — the network was too slow. Please try again.";
  }

  // 6. Simulation failed — try to extract program error code
  // Anchor encodes custom errors as 0x1770 + index
  const hexMatch = raw.match(/0x([0-9a-fA-F]+)/);
  if (hexMatch) {
    const code = parseInt(hexMatch[1], 16);
    if (PROGRAM_ERRORS[code]) {
      return PROGRAM_ERRORS[code];
    }
  }

  // 7. Anchor error in JSON logs: "Error Code: X. Error Number: NNNN."
  const anchorMatch = raw.match(/Error Number:\s*(\d+)/);
  if (anchorMatch) {
    const code = parseInt(anchorMatch[1], 10);
    if (PROGRAM_ERRORS[code]) {
      return PROGRAM_ERRORS[code];
    }
  }

  // 8. Transaction simulation failed generic
  if (/simulation failed/i.test(raw)) {
    return "Transaction simulation failed — the on-chain program rejected the instruction. Check project status and your balances.";
  }

  // 9. Account does not exist yet
  if (/AccountNotFound/i.test(raw) || /account.*not.*exist/i.test(raw)) {
    return "Required account not found on-chain. The project may not be fully initialised yet.";
  }

  // 10. RPC / network error
  if (/failed to fetch/i.test(raw) || /network/i.test(raw) || /econnrefused/i.test(raw)) {
    return "Network error — cannot reach Solana RPC. Check your connection and try again.";
  }

  // Fallback: trim and cap the raw message so it's at least readable
  const trimmed = raw.replace(/\s+/g, " ").trim();
  return trimmed.length > 0 ? trimmed.slice(0, 160) : "An unexpected error occurred. Please try again.";
}
