import type { Milestone } from "@/types/milestone";

const STATUS_COLORS: Record<string, string> = {
  Pending: "bg-gray-700 text-gray-400",
  UnderReview: "bg-yellow-900/40 text-yellow-400 border border-yellow-800",
  Approved: "bg-blue-900/40 text-blue-400 border border-blue-800",
  Released: "bg-green-900/40 text-green-400 border border-green-800",
  Disputed: "bg-red-900/40 text-red-400 border border-red-800",
};

const STATUS_DOT: Record<string, string> = {
  Pending: "bg-gray-600",
  UnderReview: "bg-yellow-400 animate-pulse",
  Approved: "bg-blue-400",
  Released: "bg-green-400",
  Disputed: "bg-red-400 animate-pulse",
};

interface Props {
  milestones: Milestone[];
  currentIndex: number;
}

export default function MilestoneTimeline({ milestones, currentIndex }: Props) {
  return (
    <div className="space-y-4">
      {milestones.map((m, i) => (
        <div key={i} className="flex gap-4">
          {/* Connector */}
          <div className="flex flex-col items-center">
            <div
              className={`w-3 h-3 rounded-full mt-1.5 flex-shrink-0 ${
                STATUS_DOT[m.status]
              }`}
            />
            {i < milestones.length - 1 && (
              <div className="w-0.5 flex-1 bg-gray-800 mt-1" />
            )}
          </div>

          {/* Content */}
          <div className="flex-1 pb-4">
            <div className="flex items-center gap-2 mb-1">
              <span className="font-medium text-white text-sm">
                {m.milestoneType} — Milestone {m.milestoneIndex + 1}
              </span>
              <span
                className={`badge text-xs ${STATUS_COLORS[m.status] ?? "badge-gray"}`}
              >
                {m.status}
              </span>
              {i === currentIndex && (
                <span className="badge badge-blue text-xs">Current</span>
              )}
            </div>

            <p className="text-xs text-gray-400 mb-2">{m.description}</p>

            <div className="flex items-center gap-4 text-xs text-gray-500">
              <span>Release: {(m.releaseBps / 100).toFixed(0)}%</span>
              {!!m.releasedAmountUsdc && (
                <span className="text-green-400">
                  ${(Number(m.releasedAmountUsdc) / 1_000_000).toLocaleString()} released
                </span>
              )}
              {m.proofUri && (
                <a
                  href={m.proofUri}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-brand-400 hover:underline"
                >
                  View proof ↗
                </a>
              )}
              {m.disputeDeadline && m.status === "UnderReview" && (
                <span className="text-yellow-400">
                  Dispute deadline:{" "}
                  {new Date(m.disputeDeadline * 1000).toLocaleString()}
                </span>
              )}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
