import { useState } from "react";
import { useWallet, useConnection } from "@solana/wallet-adapter-react";
import { PublicKey, Transaction } from "@solana/web3.js";
import { getAssociatedTokenAddress, TOKEN_2022_PROGRAM_ID } from "@solana/spl-token";
import * as anchor from "@coral-xyz/anchor";
import { useWalletStore } from "@/store/wallet";
import type { ApiProject } from "@/types/project";
import { PROGRAM_ID, USDC_MINT } from "@/lib/constants";
import { getInvestorPositionPDA, getEscrowVaultPDA } from "@/lib/pda";
import { parseTransactionError } from "@/lib/errors";
import IDL from "@/idl/terravault.json";

type Step = "check" | "amount" | "confirm" | "done";

interface Props {
  project: ApiProject;
  onClose: () => void;
}

export default function InvestFlow({ project, onClose }: Props) {
  const { publicKey, sendTransaction, signTransaction } = useWallet();
  const { connection } = useConnection();
  const kycVerified = useWalletStore((s) => s.kycVerified);

  const [step, setStep] = useState<Step>(() => (!publicKey ? "check" : "amount"));
  const [tokenAmount, setTokenAmount] = useState("");
  const [loading, setLoading] = useState(false);
  const [txSig, setTxSig] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const priceUsdc = project.TokenPrice / 1_000_000;
  const totalCost = (Number(tokenAmount) || 0) * priceUsdc;
  const maxTokens = project.TotalTokens - project.TokensSold;
  const name = project.name ?? `Project #${project.ID}`;

  const handleBuy = async () => {
    if (!publicKey || !sendTransaction || !tokenAmount) return;
    setLoading(true);
    setError(null);
    try {
      const projectState = new PublicKey(project.OnChainPubkey);
      const [investorPosition] = getInvestorPositionPDA(projectState, publicKey);
      const [escrowVault] = getEscrowVaultPDA(projectState);
      const investorUsdcAccount = await getAssociatedTokenAddress(
        USDC_MINT,
        publicKey,
        false,
        TOKEN_2022_PROGRAM_ID
      );

      // Build instruction via Anchor
      const provider = new anchor.AnchorProvider(
        connection,
        { publicKey, signTransaction: signTransaction!, signAllTransactions: async (txs: any[]) => txs } as any,
        { commitment: "confirmed" }
      );
      const program = new anchor.Program(IDL as anchor.Idl, provider);

      const ix = await (program.methods as any)
        .buyTokens(new anchor.BN(Number(tokenAmount)))
        .accounts({
          investor: publicKey,
          projectState,
          investorPosition,
          escrowVault,
          investorUsdcAccount,
          usdcMint: USDC_MINT,
          tokenProgram: TOKEN_2022_PROGRAM_ID,
          systemProgram: anchor.web3.SystemProgram.programId,
        })
        .instruction();

      const tx = new Transaction().add(ix);
      const { blockhash, lastValidBlockHeight } = await connection.getLatestBlockhash();
      tx.recentBlockhash = blockhash;
      tx.feePayer = publicKey;

      const sig = await sendTransaction(tx, connection, { preflightCommitment: "confirmed" });
      await connection.confirmTransaction({ signature: sig, blockhash, lastValidBlockHeight }, "confirmed");

      setTxSig(sig);
      setStep("done");
    } catch (e: unknown) {
      setError(parseTransactionError(e));
      setStep("amount");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
      <div className="card w-full max-w-md">
        <div className="flex items-center justify-between mb-6">
          <h2 className="font-bold text-lg text-white">Invest in {name}</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white text-xl">×</button>
        </div>

        {step === "check" && (
          <div className="text-center py-6">
            <div className="text-4xl mb-3">🔐</div>
            <p className="text-gray-300 mb-4">Connect your wallet to invest</p>
          </div>
        )}

        {step === "amount" && (
          <div className="space-y-4">
            <div>
              <label className="text-sm text-gray-400 mb-1 block">
                Number of tokens (max {maxTokens.toLocaleString()})
              </label>
              <input
                type="number"
                className="input"
                placeholder="0"
                min="1"
                max={maxTokens}
                value={tokenAmount}
                onChange={(e) => setTokenAmount(e.target.value)}
              />
            </div>

            <div className="bg-gray-800 rounded-xl p-4 space-y-2 text-sm">
              <div className="flex justify-between text-gray-400">
                <span>Token price</span>
                <span>${priceUsdc.toFixed(2)} USDC</span>
              </div>
              <div className="flex justify-between text-gray-400">
                <span>Quantity</span>
                <span>{Number(tokenAmount) || 0}</span>
              </div>
              <div className="border-t border-gray-700 pt-2 flex justify-between font-semibold text-white">
                <span>Total cost</span>
                <span>${totalCost.toFixed(2)} USDC</span>
              </div>
            </div>

            {error && (
              <div className="bg-red-900/40 border border-red-700 rounded-xl p-3 text-xs text-red-300 break-all">
                {error}
              </div>
            )}

            <button
              className="btn-primary w-full"
              onClick={handleBuy}
              disabled={!tokenAmount || Number(tokenAmount) < 1 || Number(tokenAmount) > maxTokens || loading}
            >
              {loading ? "Sending transaction…" : "Invest Now"}
            </button>
            <p className="text-xs text-gray-500 text-center">
              Your USDC will be held in escrow until milestone completion.
            </p>
          </div>
        )}

        {step === "done" && (
          <div className="text-center py-6">
            <div className="text-5xl mb-3">🎉</div>
            <h3 className="font-semibold text-white text-lg mb-2">Investment confirmed!</h3>
            <p className="text-gray-400 text-sm mb-2">
              You now hold{" "}
              <span className="text-brand-400 font-semibold">{tokenAmount} tokens</span>{" "}
              in {name}.
            </p>
            {txSig && (
              <p className="text-xs text-gray-500 mb-6 break-all">
                Tx: {txSig.slice(0, 20)}…
              </p>
            )}
            <button className="btn-primary" onClick={onClose}>Done</button>
          </div>
        )}
      </div>
    </div>
  );
}
