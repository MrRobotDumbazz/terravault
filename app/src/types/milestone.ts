import { PublicKey } from "@solana/web3.js";

export type MilestoneType =
  | "SitePreparation"
  | "Foundation"
  | "Framing"
  | "Roofing"
  | "MEP"
  | "InteriorWork"
  | "Landscaping"
  | "Completion"
  | "Custom";

export type MilestoneStatus =
  | "Pending"
  | "UnderReview"
  | "Approved"
  | "Released"
  | "Disputed";

export interface Milestone {
  publicKey: PublicKey;
  project: PublicKey;
  milestoneIndex: number;
  milestoneType: MilestoneType;
  description: string;
  releaseBps: number;
  status: MilestoneStatus;
  proofUri?: string;
  proofHash?: Uint8Array;
  submittedAt?: number;
  approvedAt?: number;
  releasedAmountUsdc?: bigint;
  disputeDeadline?: number;
}
