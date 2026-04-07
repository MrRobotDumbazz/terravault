import { AnchorProvider, Program, Idl } from "@coral-xyz/anchor";
import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { useMemo } from "react";
import { PROGRAM_ID } from "./constants";

// Import generated IDL after `anchor build`
// import IDL from "@/idl/terravault.json";

export function useAnchorProvider() {
  const { connection } = useConnection();
  const wallet = useWallet();

  return useMemo(() => {
    if (!wallet.publicKey || !wallet.signTransaction) return null;
    return new AnchorProvider(connection, wallet as any, {
      commitment: "confirmed",
      preflightCommitment: "confirmed",
    });
  }, [connection, wallet]);
}

export function useProgram(idl: Idl) {
  const provider = useAnchorProvider();
  return useMemo(() => {
    if (!provider) return null;
    return new Program(idl, provider);
  }, [provider, idl]);
}
