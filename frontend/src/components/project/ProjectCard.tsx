import { Link } from "react-router-dom";
import type { ApiProject } from "@/types/project";

const STATUS_BADGE: Record<string, string> = {
  Fundraising: "badge-green",
  Active: "badge-blue",
  InMilestones: "badge-yellow",
  Completed: "badge-gray",
  Distributing: "badge-blue",
  Cancelled: "badge-red",
  Paused: "badge-yellow",
  Draft: "badge-gray",
};

const TYPE_LABEL: Record<string, string> = {
  Residential: "🏠 Residential",
  Commercial: "🏢 Commercial",
  Agricultural: "🌾 Agricultural",
  Mixed: "🏗 Mixed",
};

interface Props {
  project: ApiProject;
}

export default function ProjectCard({ project }: Props) {
  const hardCap = project.FundraiseHardCap || 1;
  const raisedPct = Math.round((project.TotalRaised / hardCap) * 100);

  const priceUsd = (project.TokenPrice / 1_000_000).toFixed(2);
  const raisedUsd = (project.TotalRaised / 1_000_000).toLocaleString();
  const targetUsd = (project.FundraiseTarget / 1_000_000).toLocaleString();

  const name = project.name ?? project.MetadataURI ?? `Project #${project.ID}`;

  return (
    <Link to={`/projects/${project.OnChainPubkey}`}>
      <div className="card hover:border-brand-600 transition-colors cursor-pointer group">
        {/* Image */}
        <div className="relative h-44 rounded-xl overflow-hidden bg-gray-800 mb-4">
          {project.imageUrl ? (
            <img
              src={project.imageUrl}
              alt={name}
              className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-4xl">
              🏗
            </div>
          )}
          <div className="absolute top-3 left-3">
            <span className={STATUS_BADGE[project.State] ?? "badge-gray"}>
              {project.State}
            </span>
          </div>
          <div className="absolute top-3 right-3 bg-gray-900/80 px-2 py-1 rounded-lg text-xs text-gray-300">
            {TYPE_LABEL[project.ProjectType] ?? project.ProjectType}
          </div>
        </div>

        {/* Info */}
        <h3 className="font-semibold text-white text-lg mb-1 truncate">
          {name}
        </h3>
        <p className="text-sm text-gray-400 mb-4">
          {project.MilestoneCount} milestones ·{" "}
          {project.MilestonesCompleted}/{project.MilestoneCount} done
        </p>

        {/* Fundraise progress */}
        {(project.State === "Fundraising" || project.State === "Active") && (
          <div className="mb-4">
            <div className="flex justify-between text-xs text-gray-400 mb-1">
              <span>${raisedUsd} raised</span>
              <span>Target: ${targetUsd}</span>
            </div>
            <div className="h-2 bg-gray-800 rounded-full overflow-hidden">
              <div
                className="h-full bg-brand-500 rounded-full transition-all"
                style={{ width: `${Math.min(raisedPct, 100)}%` }}
              />
            </div>
            <div className="text-right text-xs text-gray-500 mt-1">
              {raisedPct}%
            </div>
          </div>
        )}

        {/* Token price */}
        <div className="flex items-center justify-between">
          <div>
            <div className="text-xs text-gray-500">Token price</div>
            <div className="font-semibold text-white">${priceUsd} USDC</div>
          </div>
          <div className="text-right">
            <div className="text-xs text-gray-500">Tokens sold</div>
            <div className="font-semibold text-white">
              {project.TokensSold.toLocaleString()} /{" "}
              {project.TotalTokens.toLocaleString()}
            </div>
          </div>
        </div>
      </div>
    </Link>
  );
}
