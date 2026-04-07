import { useEffect, useState } from "react";
import { useWalletStore } from "@/store/wallet";
import DisputePanel from "@/components/project/DisputePanel";

interface OpenDispute {
  project_pubkey: string;
  raised_by: string;
  deadline: number;
}

export default function AdminPanel() {
  const { role, token } = useWalletStore((s) => ({ role: s.role, token: s.token }));
  const [disputes, setDisputes] = useState<OpenDispute[]>([]);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<string | null>(null);

  useEffect(() => {
    if (role !== "admin") return;
    // Fetch open disputes via project list — filter DisputeActive state
    fetch("/api/v1/admin/projects", {
      headers: { Authorization: `Bearer ${token ?? ""}` },
    })
      .then((r) => r.json())
      .then((data) => {
        const open: OpenDispute[] = (data.projects ?? [])
          .filter((p: { state: string }) => p.state === "DisputeActive")
          .map((p: { on_chain_pubkey: string; developer_wallet: string }) => ({
            project_pubkey: p.on_chain_pubkey,
            raised_by: p.developer_wallet,
            deadline: 0,
          }));
        setDisputes(open);
      })
      .finally(() => setLoading(false));
  }, [role, token]);

  if (role !== "admin") {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 flex flex-col items-center justify-center min-h-[60vh]">
        <div className="text-6xl font-bold text-gray-700 mb-4">403</div>
        <h1 className="text-2xl font-semibold text-white mb-2">Access Denied</h1>
        <p className="text-gray-400">This page is restricted to administrators.</p>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <h1 className="text-3xl font-bold text-white mb-8">Admin Panel</h1>

      <div className="card">
        <h2 className="font-semibold text-white mb-6">
          Open Disputes ({loading ? "…" : disputes.length})
        </h2>

        {!loading && disputes.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            <div className="text-3xl mb-3">✅</div>
            No open disputes.
          </div>
        )}

        {disputes.length > 0 && (
          <div className="space-y-4">
            {disputes.map((d) => (
              <div key={d.project_pubkey} className="bg-gray-800 rounded-xl p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-mono text-sm text-white">
                      {d.project_pubkey.slice(0, 8)}…{d.project_pubkey.slice(-6)}
                    </div>
                    <div className="text-xs text-gray-400 mt-0.5">
                      Raised by {d.raised_by.slice(0, 8)}…
                    </div>
                  </div>
                  <button
                    onClick={() =>
                      setSelected((prev) =>
                        prev === d.project_pubkey ? null : d.project_pubkey
                      )
                    }
                    className="btn-outline text-sm py-1.5 px-4"
                  >
                    {selected === d.project_pubkey ? "Collapse" : "Resolve"}
                  </button>
                </div>

                {selected === d.project_pubkey && (
                  <DisputePanel
                    projectPubkey={d.project_pubkey}
                    onResolved={() => {
                      setSelected(null);
                      setDisputes((prev) =>
                        prev.filter((x) => x.project_pubkey !== d.project_pubkey)
                      );
                    }}
                  />
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
