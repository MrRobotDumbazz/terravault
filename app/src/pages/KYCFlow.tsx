import { useState } from "react";
import { useWallet } from "@solana/wallet-adapter-react";
import { useWalletStore } from "@/store/wallet";
import { API_BASE_URL } from "@/lib/constants";

type Step = "intro" | "pending" | "done" | "rejected";

export default function KYCFlow() {
  const { publicKey } = useWallet();
  const setKycVerified = useWalletStore((s) => s.setKycVerified);
  const [step, setStep] = useState<Step>("intro");
  const [loading, setLoading] = useState(false);

  const startKYC = async () => {
    if (!publicKey) return;
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE_URL}/v1/me/kyc/initiate`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${localStorage.getItem("tv_token") ?? ""}`,
        },
        body: JSON.stringify({ wallet: publicKey.toBase58() }),
      });
      const data = await res.json();
      // Redirect to KYC provider (Persona, etc.)
      if (data.sessionUrl) {
        window.open(data.sessionUrl, "_blank");
        setStep("pending");
      }
    } catch {
      // Handle error
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-lg mx-auto px-4 py-20">
      <div className="card text-center">
        {step === "intro" && (
          <>
            <div className="text-5xl mb-6">📋</div>
            <h1 className="text-2xl font-bold text-white mb-3">
              Identity Verification
            </h1>
            <p className="text-gray-400 text-sm mb-8 leading-relaxed">
              To comply with financial regulations, TerraVault requires KYC
              (Know Your Customer) verification before investing. This is a
              one-time process per wallet.
            </p>

            <div className="text-left space-y-3 mb-8">
              {[
                "Government-issued ID or passport",
                "Selfie photo for liveness check",
                "Takes ~3 minutes",
                "Data handled by our KYC provider (Persona)",
              ].map((item) => (
                <div key={item} className="flex gap-3 text-sm">
                  <span className="text-brand-400">✓</span>
                  <span className="text-gray-300">{item}</span>
                </div>
              ))}
            </div>

            <button
              className="btn-primary w-full"
              onClick={startKYC}
              disabled={!publicKey || loading}
            >
              {loading ? "Initiating…" : "Start Verification →"}
            </button>
          </>
        )}

        {step === "pending" && (
          <>
            <div className="text-5xl mb-6">⏳</div>
            <h2 className="text-xl font-bold text-white mb-3">
              Verification in Progress
            </h2>
            <p className="text-gray-400 text-sm mb-6">
              Complete the verification in the new tab. Once approved, your
              wallet will be automatically verified.
            </p>
            <button
              className="btn-secondary w-full"
              onClick={() => setStep("intro")}
            >
              Restart
            </button>
          </>
        )}

        {step === "done" && (
          <>
            <div className="text-5xl mb-6">🎉</div>
            <h2 className="text-xl font-bold text-white mb-3">
              Verification Complete!
            </h2>
            <p className="text-gray-400 text-sm mb-6">
              Your wallet is now KYC-verified. You can invest in all projects.
            </p>
            <a href="/projects" className="btn-primary inline-block">
              Browse Projects →
            </a>
          </>
        )}
      </div>
    </div>
  );
}
