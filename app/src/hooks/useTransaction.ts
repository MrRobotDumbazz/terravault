import { useCallback, useState } from "react";
import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { Transaction, VersionedTransaction } from "@solana/web3.js";
import { useUIStore } from "@/store/ui";

interface UseTransactionOptions {
  onSuccess?: (signature: string) => void;
  onError?: (error: Error) => void;
}

export function useTransaction(options?: UseTransactionOptions) {
  const { connection } = useConnection();
  const { sendTransaction, publicKey } = useWallet();
  const addToast = useUIStore((s) => s.addToast);
  const [loading, setLoading] = useState(false);
  const [signature, setSignature] = useState<string | null>(null);

  const execute = useCallback(
    async (tx: Transaction | VersionedTransaction) => {
      if (!publicKey) {
        addToast({ type: "error", title: "Wallet not connected" });
        return;
      }

      setLoading(true);
      try {
        // Add recent blockhash for legacy transactions
        if (tx instanceof Transaction) {
          tx.recentBlockhash = (
            await connection.getLatestBlockhash()
          ).blockhash;
          tx.feePayer = publicKey;
        }

        const sig = await sendTransaction(tx, connection, {
          preflightCommitment: "confirmed",
          skipPreflight: false,
        });

        setSignature(sig);

        // Wait for confirmation
        const { blockhash, lastValidBlockHeight } =
          await connection.getLatestBlockhash();
        await connection.confirmTransaction(
          { signature: sig, blockhash, lastValidBlockHeight },
          "confirmed"
        );

        addToast({
          type: "success",
          title: "Transaction confirmed",
          description: `${sig.slice(0, 8)}…`,
        });
        options?.onSuccess?.(sig);
        return sig;
      } catch (err) {
        const error = err instanceof Error ? err : new Error(String(err));
        addToast({
          type: "error",
          title: "Transaction failed",
          description: error.message,
        });
        options?.onError?.(error);
        throw error;
      } finally {
        setLoading(false);
      }
    },
    [connection, publicKey, sendTransaction, addToast, options]
  );

  return { execute, loading, signature };
}
