import {
  type ContestStatsFields,
  formatSuspicionPercent,
  suspicionLevel,
} from "@/features/contests/contestHelpers";
import {
  type ProblemSignalStat,
  problemSignalLevel,
} from "@/features/contests/problemSignalStats";
import ImportProgressBar from "@/features/contests/ImportProgressBar";
import type { ImportProgress } from "@/features/contests/importJobShared";
import { formatSignalCount } from "@/features/findings/reportModel";

type Props = {
  stats: ContestStatsFields;
  problemStats?: ProblemSignalStat[];
  /** Показать % badge (в шапке контеста — да; на карточке badge уже в top-row). */
  showBadge?: boolean;
  /** Пока идёт import — бар вместо чипов задач. */
  importProgress?: ImportProgress;
};

export default function ContestStats({
  stats,
  problemStats,
  showBadge = true,
  importProgress,
}: Props) {
  const level = suspicionLevel(stats.weightedSuspicionPercent);
  const subs = stats.submissionCount ?? 0;
  const maxSignals = problemStats?.reduce((m, p) => Math.max(m, p.signalCount), 0) ?? 0;
  const problemCount = problemStats?.length ?? 0;
  const problemsDensity =
    problemCount >= 12 ? "dense" : problemCount >= 8 ? "compact" : null;

  return (
    <div className={`contest-stats${showBadge ? "" : " contest-stats--meta-only"}`}>
      <div className="contest-stats__left">
        {showBadge && (
          <span
            className={`score-badge contest-stats__badge ${level}`}
            title="Взвешенный процент подозрительности"
          >
            {formatSuspicionPercent(stats.weightedSuspicionPercent)}
          </span>
        )}
        <span className="contest-stats__meta contest-stats__muted">{subs} посылок</span>
      </div>
      {importProgress !== undefined ? (
        <ImportProgressBar progress={importProgress} className="import-progress-bar--card" />
      ) : problemStats && problemStats.length > 0 ? (
        <span
          className={`contest-stats__problems${problemsDensity ? ` contest-stats__problems--${problemsDensity}` : ""}`}
          aria-label="Сигналы по задачам"
        >
          {problemStats.map((p) => {
            const label = p.name || p.id;
            const heat = problemSignalLevel(p.signalCount, maxSignals);
            return (
              <span
                key={p.id}
                className={`contest-problem-chip ${heat}`}
                title={formatSignalCount(p.signalCount)}
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
