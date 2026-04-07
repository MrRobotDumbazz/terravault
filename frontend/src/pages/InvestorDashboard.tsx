import { useWallet } from "@solana/wallet-adapter-react";
import { useMyPositions } from "@/hooks/useInvestorPosition";
import { Link } from "react-router-dom";

export default function InvestorDashboard() {
  const { publicKey } = useWallet();
  const { data: positions, isLoading } = useMyPositions();

  if (!publicKey) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-20 text-center">
        <div className="text-5xl mb-4">🔐</div>
        <h2 className="text-2xl font-bold text-white mb-2">Connect Wallet</h2>
        <p className="text-gray-400">Connect your Phantom wallet to view your portfolio.</p>
      </div>
    );
  }

  const totalInvested = positions?.reduce(
    (sum: number, p: any) => sum + Number(p.usdcInvested) / 1_000_000,
    0
  ) ?? 0;

  const pendingDistributions = positions?.reduce(
    (sum: number, p: any) => sum + (p.pendingClaimUsdc ?? 0) / 1_000_000,
    0
  ) ?? 0;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <h1 className="text-3xl font-bold text-white mb-8">My Portfolio</h1>

      {/* Summary cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-10">
        {[
          { label: "Total Invested", value: `$${totalInvested.toLocaleString()}` },
          { label: "Active Positions", value: positions?.length ?? 0 },
          { label: "Pending Income", value: `$${pendingDistributions.toFixed(2)}` },
        ].map((s) => (
          <div key={s.label} className="card">
            <div className="text-xs text-gray-500 mb-1">{s.label}</div>
            <div className="text-2xl font-bold text-white">{s.value}</div>
          </div>
        ))}
      </div>

      {/* Positions list */}
      {isLoading && (
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="card h-24 animate-pulse" />
          ))}
        </div>
      )}

      {positions && positions.length === 0 && (
        <div className="card text-center py-16">
          <div className="text-4xl mb-4">🌱</div>
          <h3 className="font-semibold text-white mb-2">No positions yet</h3>
          <p className="text-gray-400 text-sm mb-6">
            Start investing in real estate projects to build your portfolio.
          </p>
          <Link to="/projects" className="btn-primary">
            Browse Projects
          </Link>
        </div>
      )}

      {positions?.map((pos: any) => (
        <div key={pos.project} className="card mb-4 flex items-center gap-6">
          <div className="flex-1">
            <Link
              to={`/projects/${pos.project}`}
              className="font-semibold text-white hover:text-brand-400"
            >
              {pos.projectName ?? pos.project.slice(0, 8) + "…"}
            </Link>
            <div className="text-sm text-gray-400 mt-1">
              {Number(pos.tokensHeld).toLocaleString()} tokens ·{" "}
              ${(Number(pos.usdcInvested) / 1_000_000).toFixed(2)} invested
            </div>
          </div>
          {pos.pendingClaimUsdc > 0 && (
            <div className="text-right">
              <div className="text-xs text-gray-500">Pending income</div>
              <div className="font-semibold text-green-400">
                +${(Number(pos.pendingClaimUsdc) / 1_000_000).toFixed(4)}
              </div>
              <button className="btn-primary mt-2 text-sm py-1 px-3">
                Claim
              </button>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
