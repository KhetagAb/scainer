import { type ContestStatsFields } from "@/features/contests/contestHelpers";

type Props = {
  stats: ContestStatsFields;
};

export default function ContestStats({ stats }: Props) {
  const subs = stats.submissionCount ?? 0;

  return (
    <div className="contest-stats contest-stats--meta-only">
      <div className="contest-stats__left">
        <span className="contest-stats__meta contest-stats__muted">{subs} посылок</span>
      </div>
    </div>
  );
}
