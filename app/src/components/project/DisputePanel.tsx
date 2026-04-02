import { useEffect, useRef, useState } from "react";
import { useWalletStore } from "@/store/wallet";

interface DisputeData {
  project_pubkey: string;
  raised_by: string;
  evidence_hash: string;
  status: "open" | "resolved";
  decision: string;
  deadline: number; // unix timestamp
}

interface Props {
  projectPubkey: string;
  onResolved?: () => void;
}

function useCountdown(deadlineUnix: number) {
  const [remaining, setRemaining] = useState(deadlineUnix - Math.floor(Date.now() / 1000));

  useEffect(() => {
    const id = setInterval(() => {
      setRemaining(deadlineUnix - Math.floor(Date.now() / 1000));
    }, 1000);
    return () => clearInterval(id);
  }, [deadlineUnix]);

  if (remaining <= 0) return "Expired";
  const h = Math.floor(remaining / 3600);
  const m = Math.floor((remaining % 3600) / 60);
  const s = remaining % 60;
  return `${h}h ${m}m ${s}s`;
}

const DECISIONS = ["PayInvestors", "RefundAndExtend", "ForceClose"] as const;

export default function DisputePanel({ projectPubkey, onResolved }: Props) {
  const { role, token } = useWalletStore((s) => ({ role: s.role, token: s.token }));
  const [dispute, setDispute] = useState<DisputeData | null>(null);
  const [loading, setLoading] = useState(true);
  const [decision, setDecision] = useState<(typeof DECISIONS)[number]>("PayInvestors");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const fileRef = useRef<HTMLInputElement>(null);

  const countdown = useCountdown(dispute?.deadline ?? 0);

  useEffect(() => {
    fetch(`/api/v1/disputes/${projectPubkey}`, {
      headers: { Authorization: `Bearer ${token ?? ""}` },
    })
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => setDispute(data))
      .finally(() => setLoading(false));
  }, [projectPubkey, token]);

  async function submitEvidence() {
    const file = fileRef.current?.files?.[0];
    if (!file) return;
    setSubmitting(true);
    setError("");
    const form = new FormData();
    form.append("project_pubkey", projectPubkey);
    form.append("file", file);
    const res = await fetch("/api/v1/disputes/evidence", {
      method: "POST",
      headers: { Authorization: `Bearer ${token ?? ""}` },
      body: form,
    });
    if (!res.ok) setError("Failed to submit evidence");
    else {
      const data = await res.json();
      setDispute((prev) => prev ? { ...prev, evidence_hash: data.evidence_cid } : prev);
    }
    setSubmitting(false);
  }

  async function resolveDispute() {
    setSubmitting(true);
    setError("");
    const res = await fetch("/api/v1/disputes/resolve", {
      method: "POST",
      headers: { Authorization: `Bearer ${token ?? ""}`, "Content-Type": "application/json" },
      body: JSON.stringify({ project_pubkey: projectPubkey, decision }),
    });
    if (!res.ok) setError("Failed to resolve dispute");
    else {
      setDispute((prev) => prev ? { ...prev, status: "resolved", decision } : prev);
      onResolved?.();
    }
    setSubmitting(false);
  }

  if (loading) return <div className="text-gray-400 text-sm">Loading dispute…</div>;
  if (!dispute) return null;

  const windowOpen = dispute.status === "open";

  return (
    <div className="card border border-red-800/60 space-y-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="w-2.5 h-2.5 rounded-full bg-red-500 animate-pulse" />
          <h2 className="font-semibold text-white">Dispute Active</h2>
        </div>
        {windowOpen && (
          <div className="text-sm font-mono text-red-400">{countdown}</div>
        )}
        {!windowOpen && (
          <span className="text-xs bg-gray-700 text-gray-300 px-2 py-0.5 rounded-full">
            Resolved · {dispute.decision}
          </span>
        )}
      </div>

      {/* Evidence hashes */}
      <div className="space-y-1">
        <div className="text-xs text-gray-500 uppercase tracking-wide">Evidence hash</div>
        <div className="font-mono text-xs text-gray-300 break-all">
          {dispute.evidence_hash || "—"}
        </div>
      </div>

      {/* Submit evidence (window must be open) */}
      {windowOpen && (
        <div className="space-y-2">
          <div className="text-xs text-gray-500 uppercase tracking-wide">Submit evidence</div>
          <div className="flex gap-2">
            <input
              ref={fileRef}
              type="file"
              accept=".pdf,.png,.jpg,.zip"
              className="block text-sm text-gray-400 file:mr-3 file:py-1.5 file:px-4 file:rounded-lg file:border-0 file:bg-gray-700 file:text-white hover:file:bg-gray-600"
            />
            <button
              onClick={submitEvidence}
              disabled={submitting}
              className="btn-outline text-sm py-1.5 px-4"
            >
              Upload
            </button>
          </div>
        </div>
      )}

      {/* Admin resolve */}
      {windowOpen && role === "admin" && (
        <div className="border-t border-gray-700 pt-4 space-y-3">
          <div className="text-xs text-gray-500 uppercase tracking-wide">Admin decision</div>
          <select
            value={decision}
            onChange={(e) => setDecision(e.target.value as typeof decision)}
            className="w-full bg-gray-800 border border-gray-700 text-white text-sm rounded-lg px-3 py-2"
          >
            {DECISIONS.map((d) => (
              <option key={d} value={d}>{d}</option>
            ))}
          </select>
          <button
            onClick={resolveDispute}
            disabled={submitting}
            className="btn-primary w-full text-sm py-2"
          >
            Resolve Dispute
          </button>
        </div>
      )}

      {error && <div className="text-red-400 text-sm">{error}</div>}
    </div>
  );
}
