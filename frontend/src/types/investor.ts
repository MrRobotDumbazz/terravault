import { PublicKey } from "@solana/web3.js";

export interface InvestorPosition {
  publicKey: PublicKey;
  project: PublicKey;
  investor: PublicKey;
  tokensHeld: bigint;
  usdcInvested: bigint;
  lastClaimedRound: number;
  totalClaimedUsdc: bigint;
  kycVerified: boolean;
  kycTimestamp?: number;
  createdAt: number;
}

export interface DistributionPool {
  publicKey: PublicKey;
  project: PublicKey;
  round: number;
  totalUsdcDeposited: bigint;
  totalTokensAtSnapshot: bigint;
  usdcPerTokenScaled: bigint;
  depositedAt: number;
  source: "RentalIncome" | "SaleProceeds" | "Refinancing" | "Other";
  claimDeadline: number;
  totalClaimed: bigint;
}
