import { Link, useLocation } from "react-router-dom";
import { useWallet } from "@solana/wallet-adapter-react";
import { useWalletStore } from "@/store/wallet";
import WalletButton from "@/components/ui/WalletButton";

const NAV_LINKS = [
  { to: "/projects", label: "Projects" },
  { to: "/dashboard", label: "Portfolio" },
  { to: "/developer", label: "Developer" },
];

export default function Navbar() {
  const location = useLocation();
  const { publicKey } = useWallet();
  const kycVerified = useWalletStore((s) => s.kycVerified);

  return (
    <nav className="sticky top-0 z-50 border-b border-gray-800 bg-gray-950/90 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          {/* Logo */}
          <Link to="/" className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-brand-400 to-earth-500 flex items-center justify-center">
              <span className="text-white font-bold text-sm">TV</span>
            </div>
            <span className="font-bold text-lg text-white tracking-tight">
              TerraVault
            </span>
          </Link>

          {/* Nav links */}
          <div className="hidden md:flex items-center gap-1">
            {NAV_LINKS.map(({ to, label }) => (
              <Link
                key={to}
                to={to}
                className={`px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                  location.pathname.startsWith(to)
                    ? "bg-gray-800 text-white"
                    : "text-gray-400 hover:text-white hover:bg-gray-800/50"
                }`}
              >
                {label}
              </Link>
            ))}
          </div>

          {/* Wallet + KYC */}
          <div className="flex items-center gap-3">
            {publicKey && !kycVerified && (
              <Link
                to="/kyc"
                className="text-xs text-yellow-400 border border-yellow-800 px-3 py-1 rounded-lg hover:bg-yellow-900/20 transition-colors"
              >
                Complete KYC
              </Link>
            )}
            {publicKey && kycVerified && (
              <span className="badge-green text-xs">KYC ✓</span>
            )}
            <WalletButton />
          </div>
        </div>
      </div>
    </nav>
  );
}
