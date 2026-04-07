import { Link } from "react-router-dom";

export default function Landing() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      {/* Hero */}
      <section className="pt-20 pb-16 text-center">
        <div className="inline-flex items-center gap-2 bg-brand-900/30 border border-brand-800 text-brand-400 text-sm px-4 py-1.5 rounded-full mb-6">
          <span className="w-2 h-2 rounded-full bg-brand-400 animate-pulse" />
          Built on Solana · SPL Token-2022
        </div>

        <h1 className="text-5xl sm:text-6xl font-bold text-white mb-6 leading-tight">
          Real Estate Ownership,
          <br />
          <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-400 to-earth-400">
            Tokenized on Solana
          </span>
        </h1>

        <p className="text-gray-400 text-xl max-w-2xl mx-auto mb-10">
          Invest fractionally in pre-construction real estate worldwide.
          Milestone-based escrow protects your capital. Automatic income
          distribution rewards token holders.
        </p>

        <div className="flex gap-4 justify-center flex-wrap">
          <Link to="/projects" className="btn-primary text-base px-8 py-3">
            Browse Projects
          </Link>
          <Link to="/developer" className="btn-outline text-base px-8 py-3">
            List a Property
          </Link>
        </div>
      </section>

      {/* Features */}
      <section className="py-16 grid grid-cols-1 md:grid-cols-3 gap-6">
        {[
          {
            icon: "🔒",
            title: "Milestone Escrow",
            desc: "USDC is released to developers only after on-chain milestone verification. 48-hour dispute window on every release.",
          },
          {
            icon: "🏗",
            title: "Any Property Type",
            desc: "Residential, commercial, and agricultural pre-construction projects. Each project is a unique SPL Token-2022 issuance.",
          },
          {
            icon: "💰",
            title: "Automatic Income Distribution",
            desc: "Rental income and sale proceeds are distributed proportionally to all token holders. Claim on-chain, trustlessly.",
          },
        ].map((f) => (
          <div key={f.title} className="card">
            <div className="text-3xl mb-4">{f.icon}</div>
            <h3 className="font-semibold text-white text-lg mb-2">{f.title}</h3>
            <p className="text-gray-400 text-sm leading-relaxed">{f.desc}</p>
          </div>
        ))}
      </section>

      {/* Stats placeholder */}
      <section className="py-12 border-t border-gray-800">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
          {[
            { label: "Total Value Locked", value: "$0" },
            { label: "Active Projects", value: "0" },
            { label: "Token Holders", value: "0" },
            { label: "Income Distributed", value: "$0" },
          ].map((s) => (
            <div key={s.label}>
              <div className="text-3xl font-bold text-white mb-1">{s.value}</div>
              <div className="text-sm text-gray-500">{s.label}</div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
