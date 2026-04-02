import { PublicKey } from "@solana/web3.js";

// Replace with actual program ID after `anchor build && anchor deploy`
export const PROGRAM_ID = new PublicKey(
  "DoAFjsoY9Ws7ZTNCokpsYHyNho8Krj9nK5dQFCdgYQqM"
);

// USDC mint addresses
export const USDC_MINT_DEVNET = new PublicKey(
  "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU"
);
export const USDC_MINT_MAINNET = new PublicKey(
  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
);

// Localnet USDC mint created by setup-localnet.ts
export const USDC_MINT_LOCALNET = new PublicKey(
  import.meta.env.VITE_USDC_MINT ?? "DackUugfpmHFvzwNa7oNyZ6RJZTFKUcAsXR9z7FWHk3C"
);

export const USDC_MINT =
  import.meta.env.VITE_CLUSTER === "mainnet-beta"
    ? USDC_MINT_MAINNET
    : import.meta.env.VITE_CLUSTER === "localnet"
    ? USDC_MINT_LOCALNET
    : USDC_MINT_DEVNET;

export const USDC_DECIMALS = 6;
export const TOKEN_DECIMALS = 0; // Project tokens are whole units

export const DISPUTE_WINDOW_SECONDS = 172_800; // 48h
export const BASIS_POINTS = 10_000;

// API base URL — Go oracle/backend (includes /api prefix)
export const API_BASE_URL =
  (import.meta.env.VITE_API_URL ?? "http://localhost:8080") + "/api";
