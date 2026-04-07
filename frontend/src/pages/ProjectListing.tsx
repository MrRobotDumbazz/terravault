import { useState } from "react";
import { useProjects } from "@/hooks/useProject";
import ProjectCard from "@/components/project/ProjectCard";

const TYPES = ["All", "Residential", "Commercial", "Agricultural", "Mixed"];
const STATUSES = ["All", "Fundraising", "Active", "InMilestones", "Completed"];

export default function ProjectListing() {
  const [typeFilter, setTypeFilter] = useState("All");
  const [statusFilter, setStatusFilter] = useState("All");

  const { data, isLoading, error } = useProjects({
    type: typeFilter !== "All" ? typeFilter : undefined,
    status: statusFilter !== "All" ? statusFilter : undefined,
  });

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <h1 className="text-3xl font-bold text-white mb-8">Projects</h1>

      {/* Filters */}
      <div className="flex flex-wrap gap-6 mb-8">
        <div>
          <label className="text-xs text-gray-500 mb-2 block">Type</label>
          <div className="flex gap-2 flex-wrap">
            {TYPES.map((t) => (
              <button
                key={t}
                onClick={() => setTypeFilter(t)}
                className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                  typeFilter === t
                    ? "bg-brand-600 text-white"
                    : "bg-gray-800 text-gray-400 hover:bg-gray-700"
                }`}
              >
                {t}
              </button>
            ))}
          </div>
        </div>
        <div>
          <label className="text-xs text-gray-500 mb-2 block">Status</label>
          <div className="flex gap-2 flex-wrap">
            {STATUSES.map((s) => (
              <button
                key={s}
                onClick={() => setStatusFilter(s)}
                className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                  statusFilter === s
                    ? "bg-brand-600 text-white"
                    : "bg-gray-800 text-gray-400 hover:bg-gray-700"
                }`}
              >
                {s}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Grid */}
      {isLoading && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {Array.from({ length: 6 }).map((_, i) => (
            <div
              key={i}
              className="card h-72 animate-pulse bg-gray-900"
            />
          ))}
        </div>
      )}

      {error && (
        <div className="card border-red-900 text-red-400 text-center py-12">
          Failed to load projects. Make sure the Go API is running.
        </div>
      )}

      {data && (data.projects ?? []).length === 0 && (
        <div className="text-center py-20 text-gray-500">
          No projects found matching your filters.
        </div>
      )}

      {data && (data.projects ?? []).length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {(data.projects ?? []).map((p) => (
            <ProjectCard key={p.OnChainPubkey || p.ID} project={p} />
          ))}
        </div>
      )}
    </div>
  );
}
