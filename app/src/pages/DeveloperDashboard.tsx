import { Link } from "react-router-dom";
import { useWallet } from "@solana/wallet-adapter-react";

export default function DeveloperDashboard() {
  const { publicKey } = useWallet();

  if (!publicKey) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-20 text-center">
        <div className="text-5xl mb-4">👷</div>
        <h2 className="text-2xl font-bold text-white mb-2">Developer Portal</h2>
        <p className="text-gray-400">Connect your wallet to manage your projects.</p>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold text-white">My Projects</h1>
        <Link to="/developer/create" className="btn-primary">
          + New Project
        </Link>
      </div>

      <div className="card text-center py-20">
        <div className="text-4xl mb-4">🏗</div>
        <h3 className="font-semibold text-white mb-2">No projects yet</h3>
        <p className="text-gray-400 text-sm mb-6">
          Create your first tokenized real estate project.
        </p>
        <Link to="/developer/create" className="btn-primary">
          Create Project →
        </Link>
      </div>
    </div>
  );
}
