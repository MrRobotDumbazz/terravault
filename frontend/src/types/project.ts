import { PublicKey } from "@solana/web3.js";

/** Project as returned by the Go REST API (PascalCase, no PublicKey objects) */
export interface ApiProject {
  ID: number;
  OnChainPubkey: string;
  DeveloperWallet: string;
  State: ProjectStatus;
  ProjectType: ProjectType;
  MetadataURI: string;
  FundraiseTarget: number;
  FundraiseHardCap: number;
  FundraiseDeadline: string;
  EscrowBalance: number;
  TotalRaised: number;
  TokenPrice: number;
  TotalTokens: number;
  TokensSold: number;
  MilestoneCount: number;
  MilestonesCompleted: number;
  CreatedAt: string;
  UpdatedAt: string;
  // optional enriched fields
  name?: string;
  imageUrl?: string;
}

export type ProjectStatus =
  | "Draft"
  | "Fundraising"
  | "Active"
  | "InMilestones"
  | "Completed"
  | "Distributing"
  | "Closed"
  | "Cancelled"
  | "Paused";

export type ProjectType = "Residential" | "Commercial" | "Agricultural" | "Mixed";

export interface Project {
  publicKey: PublicKey;
  projectId: bigint;
  developer: PublicKey;
  oracleAuthority: PublicKey;
  tokenMint: PublicKey;
  escrowVault: PublicKey;
  state: ProjectStatus;
  projectType: ProjectType;
  totalTokens: bigint;
  tokensSold: bigint;
  tokenPriceUsdc: bigint;
  fundraiseTargetUsdc: bigint;
  fundraiseHardCapUsdc: bigint;
  fundraiseDeadline: number;
  escrowBalanceUsdc: bigint;
  totalRaisedUsdc: bigint;
  milestoneCount: number;
  milestonesAdded: number;
  milestonesCompleted: number;
  currentMilestoneIndex: number;
  distributionRound: number;
  totalDistributedUsdc: bigint;
  metadataUri: string;
  legalDocHash: Uint8Array;
  kyc_required: boolean;
  transferFeeBps: number;
  paused: boolean;
  createdAt: number;
  updatedAt: number;
}

/** Off-chain metadata from the Go API, referenced by metadataUri */
export interface ProjectMetadata {
  name: string;
  description: string;
  location: {
    country: string;
    city: string;
    address?: string;
    coordinates?: { lat: number; lng: number };
  };
  images: string[];
  documents: {
    title: string;
    type: "legal_agreement" | "inspection_report" | "title_deed" | "other";
    uri: string;
    hash: string;
  }[];
  developer: {
    name: string;
    website?: string;
    verifiedAt?: number;
  };
  expectedRoi?: number;
  constructionStartDate?: number;
  expectedCompletionDate?: number;
}
