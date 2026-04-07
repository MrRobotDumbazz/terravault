import { useState } from "react";
import { useParams } from "react-router-dom";
import { useWallet } from "@solana/wallet-adapter-react";
import { useProjectDetail, useProjectMilestones } from "@/hooks/useProject";
import FundraiseProgress from "@/components/project/FundraiseProgress";
import MilestoneTimeline from "@/components/project/MilestoneTimeline";
import InvestFlow from "@/components/invest/InvestFlow";

type Tab = "overview" | "milestones" | "documents";

export default function ProjectDetail() {
  const { id } = useParams<{ id: string }>();
  const { publicKey } = useWallet();
  const [tab, setTab] = useState<Tab>("overview");
  const [investOpen, setInvestOpen] = useState(false);

  const { data: project, isLoading } = useProjectDetail(id!);
  const { data: milestones } = useProjectMilestones(id!);

  if (isLoading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-10">
        <div className="h-96 card animate-pulse" />
      </div>
    );
  }

  if (!project) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-10 text-center text-gray-500">
        Project not found.
      </div>
    );
  }

  const canInvest =
    project.State === "Fundraising" && !!publicKey;

  const priceUsd = (project.TokenPrice / 1_000_000).toFixed(6);
  const available = project.TotalTokens - project.TokensSold;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Main content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Hero */}
          <div className="card">
            <div className="flex items-start justify-between mb-4">
              <div>
                <h1 className="text-2xl font-bold text-white">
                  {project.name ?? project.MetadataURI ?? `Project #${project.ID}`}
                </h1>
                <p className="text-gray-400 text-sm mt-1">
                  {project.ProjectType} ·{" "}
                  {project.metadata?.location?.city},{" "}
                  {project.metadata?.location?.country}
                </p>
              </div>
              <span
                className={`badge ${
                  project.State === "Fundraising"
                    ? "badge-green"
                    : "badge-blue"
                }`}
              >
                {project.State}
              </span>
            </div>

            {/* Tabs */}
            <div className="flex gap-1 border-b border-gray-800 mb-6">
              {(["overview", "milestones", "documents"] as Tab[]).map((t) => (
                <button
                  key={t}
                  onClick={() => setTab(t)}
                  className={`px-4 py-2 text-sm capitalize transition-colors ${
                    tab === t
                      ? "border-b-2 border-brand-500 text-white"
                      : "text-gray-400 hover:text-white"
                  }`}
                >
                  {t}
                </button>
              ))}
            </div>

            {tab === "overview" && (
              <p className="text-gray-300 leading-relaxed">
                {project.metadata?.description ??
                  "Project description will appear here once the developer has uploaded metadata."}
              </p>
            )}

            {tab === "milestones" && milestones && (
              <MilestoneTimeline
                milestones={milestones}
                currentIndex={0}
              />
            )}

            {tab === "documents" && (
              <div className="space-y-3">
                {(project.metadata?.documents ?? []).map(
                  (doc: any, i: number) => (
                    <a
                      key={i}
                      href={doc.uri}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-3 p-3 bg-gray-800 rounded-xl hover:bg-gray-700 transition-colors"
                    >
                      <span className="text-xl">📄</span>
                      <div>
                        <div className="text-white text-sm font-medium">
                          {doc.title}
                        </div>
                        <div className="text-gray-400 text-xs">{doc.type}</div>
                      </div>
                    </a>
                  )
                )}
                {!project.metadata?.documents?.length && (
                  <p className="text-gray-500 text-sm">No documents uploaded yet.</p>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Fundraise progress */}
          {(project.State === "Fundraising" || project.State === "Active") && (
            <FundraiseProgress project={project} />
          )}

          {/* Investment panel */}
          <div className="card">
            <div className="flex justify-between mb-4">
              <div>
                <div className="text-xs text-gray-500">Token price</div>
                <div className="text-2xl font-bold text-white">
                  ${priceUsd}
                  <span className="text-sm text-gray-400 font-normal ml-1">
                    USDC
                  </span>
                </div>
              </div>
              <div className="text-right">
                <div className="text-xs text-gray-500">Available</div>
                <div className="font-semibold text-white">
                  {available.toLocaleString()}
                </div>
              </div>
            </div>

            <button
              className="btn-primary w-full"
              disabled={!canInvest}
              onClick={() => setInvestOpen(true)}
            >
              {!publicKey
                ? "Connect Wallet"
                : project.State !== "Fundraising"
                ? `Not open (${project.State})`
                : "Invest Now"}
            </button>
          </div>
        </div>
      </div>

      {investOpen && (
        <InvestFlow
          project={project as any}
          onClose={() => setInvestOpen(false)}
        />
      )}
    </div>
  );
}
