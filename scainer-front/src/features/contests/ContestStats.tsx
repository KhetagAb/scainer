import {
  type ContestStatsFields,
  formatSuspicionPercent,
} from "@/features/contests/contestHelpers";
import {
  type ProblemSignalStat,
  PROBLEM_SUSPICION_TOOLTIP,
  problemSuspicionLevel,
  problemSuspicionLevelHint,
} from "@/features/contests/problemSignalStats";
import ImportProgressBar from "@/features/contests/ImportProgressBar";
import type { ImportProgress } from "@/features/contests/importJobShared";

type Props = {
  stats: ContestStatsFields;
  problemStats?: ProblemSignalStat[];
  /** Пока идёт import — бар вместо чипов задач. */
  importProgress?: ImportProgress;
};

export default function ContestStats({
  stats,
  problemStats,
  importProgress,
}: Props) {
  const subs = stats.submissionCount ?? 0;
  const problemCount = problemStats?.length ?? 0;
  const problemsDensity =
    problemCount >= 12 ? "dense" : problemCount >= 8 ? "compact" : null;

  return (
    <div className="contest-stats contest-stats--meta-only">
      <div className="contest-stats__left">
        <span className="contest-stats__meta contest-stats__muted">{subs} посылок</span>
      </div>
      {importProgress !== undefined ? (
        <ImportProgressBar progress={importProgress} className="import-progress-bar--card" />
      ) : problemStats && problemStats.length > 0 ? (
        <span
          className={`contest-stats__problems${problemsDensity ? ` contest-stats__problems--${problemsDensity}` : ""}`}
          aria-label="Доля подозрительных по задачам"
        >
          {problemStats.map((p) => {
            const label = p.name || p.id;
            const heat = problemSuspicionLevel(p.suspiciousSharePercent);
            return (
              <span
                key={p.id}
                className={`contest-problem-chip ${heat}`}
                title={`${formatSuspicionPercent(p.suspiciousSharePercent)} — ${problemSuspicionLevelHint(heat)}. ${PROBLEM_SUSPICION_TOOLTIP}`}
              >
                {label}
              </span>
            );
          })}
        </span>
      ) : null}
    </div>
  );
}
