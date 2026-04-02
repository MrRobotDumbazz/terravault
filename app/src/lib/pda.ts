import { PublicKey } from "@solana/web3.js";
import BN from "bn.js";
import { PROGRAM_ID } from "./constants";

/** Seeds must match the Rust program exactly — byte for byte. */

export function getProjectStatePDA(
  developer: PublicKey,
  projectId: BN
): [PublicKey, number] {
  return PublicKey.findProgramAddressSync(
    [
      Buffer.from("project"),
      developer.toBuffer(),
      projectId.toArrayLike(Buffer, "le", 8),
    ],
    PROGRAM_ID
  );
}

export function getEscrowVaultPDA(
  projectState: PublicKey
): [PublicKey, number] {
  return PublicKey.findProgramAddressSync(
    [Buffer.from("escrow"), projectState.toBuffer()],
    PROGRAM_ID
  );
}

export function getTokenConfigPDA(
  projectState: PublicKey
): [PublicKey, number] {
  return PublicKey.findProgramAddressSync(
    [Buffer.from("token_config"), projectState.toBuffer()],
    PROGRAM_ID
  );
}

export function getMilestonePDA(
  projectState: PublicKey,
  index: number
): [PublicKey, number] {
  return PublicKey.findProgramAddressSync(
    [Buffer.from("milestone"), projectState.toBuffer(), Buffer.from([index])],
    PROGRAM_ID
  );
}

export function getInvestorPositionPDA(
  projectState: PublicKey,
  investor: PublicKey
): [PublicKey, number] {
  return PublicKey.findProgramAddressSync(
    [
      Buffer.from("position"),
      projectState.toBuffer(),
      investor.toBuffer(),
    ],
    PROGRAM_ID
  );
}

export function getDistributionPoolPDA(
  projectState: PublicKey,
  round: number
): [PublicKey, number] {
  const roundBuf = Buffer.alloc(4);
  roundBuf.writeUInt32LE(round);
  return PublicKey.findProgramAddressSync(
    [Buffer.from("distribution"), projectState.toBuffer(), roundBuf],
    PROGRAM_ID
  );
}

export function getIncomeVaultPDA(
  projectState: PublicKey
): [PublicKey, number] {
  return PublicKey.findProgramAddressSync(
    [Buffer.from("income"), projectState.toBuffer()],
    PROGRAM_ID
  );
}

export function getExtraAccountMetasPDA(
  mint: PublicKey
): [PublicKey, number] {
  return PublicKey.findProgramAddressSync(
    [Buffer.from("extra-account-metas"), mint.toBuffer()],
    PROGRAM_ID
  );
}
