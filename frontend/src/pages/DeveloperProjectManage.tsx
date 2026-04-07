import { useParams } from "react-router-dom";
import { useProjectDetail, useProjectMilestones } from "@/hooks/useProject";
import MilestoneTimeline from "@/components/project/MilestoneTimeline";

export default function DeveloperProjectManage() {
  const { id } = useParams<{ id: string }>();
  const { data: project, isLoading } = useProjectDetail(id!);
  const { data: milestones } = useProjectMilestones(id!);

  if (isLoading) return <div className="max-w-7xl mx-auto px-4 py-10"><div className="card h-96 animate-pulse" /></div>;
  if (!project) return <div className="max-w-7xl mx-auto px-4 py-10 text-center text-gray-500">Project not found.</div>;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">
            {(project as any).name ?? "Project"}
          </h1>
          <p className="text-gray-400 text-sm mt-1">
            State: <span className="text-brand-400">{project.State}</span>
          </p>
        </div>
        <div className="flex gap-3">
          <button className="btn-outline text-sm">Pause</button>
          <button className="btn-secondary text-sm">Update Oracle</button>
        </div>
      </div>

      {/* Milestones management */}
      <div className="card">
        <h2 className="font-semibold text-white mb-6">Milestones</h2>
        {milestones ? (
          <MilestoneTimeline
            milestones={milestones}
            currentIndex={0}
          />
        ) : (
          <p className="text-gray-500 text-sm">No milestones yet.</p>
        )}

        {project.State === "Active" || project.State === "InMilestones" ? (
          <div className="mt-6 pt-6 border-t border-gray-800">
            <h3 className="text-sm font-semibold text-white mb-4">
              Submit Milestone Proof
            </h3>
            <div className="space-y-3">
              <input
                type="text"
                className="input"
                placeholder="IPFS URI of proof documents (e.g. ipfs://Qm…)"
              />
              <p className="text-xs text-gray-500">
                Upload proof docs to IPFS via the API, then paste the URI here.
                The oracle will sign and submit the on-chain proof.
              </p>
              <button className="btn-primary">Submit to Oracle</button>
            </div>
          </div>
        ) : null}
      </div>

      {/* Income deposit */}
      {(project.State === "Completed" || project.State === "Distributing") && (
        <div className="card">
          <h2 className="font-semibold text-white mb-4">Deposit Income</h2>
          <div className="space-y-3">
            <input type="number" className="input" placeholder="Amount (USDC)" />
            <select className="input">
              <option value="RentalIncome">Rental Income</option>
              <option value="SaleProceeds">Sale Proceeds</option>
              <option value="Refinancing">Refinancing</option>
              <option value="Other">Other</option>
            </select>
            <button className="btn-primary">Deposit for Distribution</button>
          </div>
        </div>
      )}
    </div>
  );
}
