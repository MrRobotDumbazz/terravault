interface Props {
  isBlacklisted: boolean;
  developerPubkey: string;
}

export default function DeveloperBadge({ isBlacklisted, developerPubkey }: Props) {
  const short = `${developerPubkey.slice(0, 6)}…${developerPubkey.slice(-4)}`;

  if (isBlacklisted) {
    return (
      <div className="inline-flex flex-col gap-1">
        <span className="inline-flex items-center gap-1.5 text-xs font-medium bg-red-900/40 text-red-400 border border-red-800 px-2.5 py-1 rounded-full">
          <span className="w-1.5 h-1.5 rounded-full bg-red-500" />
          Blacklisted
        </span>
        <p className="text-xs text-red-400/70">
          This developer has been flagged for fraud. Exercise extreme caution.
        </p>
      </div>
    );
  }

  return (
    <span className="inline-flex items-center gap-1.5 text-xs font-medium bg-green-900/30 text-green-400 border border-green-800/50 px-2.5 py-1 rounded-full">
      <span className="w-1.5 h-1.5 rounded-full bg-green-500" />
      {short}
    </span>
  );
}
