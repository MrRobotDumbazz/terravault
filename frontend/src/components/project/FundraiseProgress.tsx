import type { ApiProject } from "@/types/project";

interface Props {
  project: ApiProject;
}

export default function FundraiseProgress({ project }: Props) {
  const raised = project.TotalRaised / 1_000_000;
  const target = project.FundraiseTarget / 1_000_000;
  const hardCap = project.FundraiseHardCap / 1_000_000 || 1;
  const pct = Math.min((raised / hardCap) * 100, 100);
  const softPct = (target / hardCap) * 100;

  const deadline = new Date(project.FundraiseDeadline);
  const daysLeft = Math.max(
    0,
    Math.ceil((deadline.getTime() - Date.now()) / 86_400_000)
  );

  return (
    <div className="card">
      <h3 className="font-semibold text-white mb-4">Fundraising Progress</h3>

      {/* Bar */}
      <div className="relative h-3 bg-gray-800 rounded-full overflow-hidden mb-2">
        <div
          className="absolute h-full bg-brand-500 rounded-full transition-all"
          style={{ width: `${pct}%` }}
        />
        {/* Soft cap marker */}
        <div
          className="absolute top-0 h-full w-0.5 bg-yellow-400"
          style={{ left: `${softPct}%` }}
        />
      </div>

      <div className="flex justify-between text-xs text-gray-400 mb-4">
        <span>0</span>
        <span className="text-yellow-400">
          Soft cap ${target.toLocaleString()}
        </span>
        <span>Hard cap ${hardCap.toLocaleString()}</span>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <div>
          <div className="text-xs text-gray-500">Raised</div>
          <div className="font-semibold text-white">
            ${raised.toLocaleString()}
          </div>
        </div>
        <div>
          <div className="text-xs text-gray-500">Tokens sold</div>
          <div className="font-semibold text-white">
            {project.TokensSold.toLocaleString()}
          </div>
        </div>
        <div>
          <div className="text-xs text-gray-500">Deadline</div>
          <div className="font-semibold text-white">
            {daysLeft > 0 ? `${daysLeft}d left` : "Ended"}
          </div>
        </div>
      </div>
    </div>
  );
}
