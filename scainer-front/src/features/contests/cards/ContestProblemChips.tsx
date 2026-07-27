import { formatSuspicionPercent } from "@/features/contests/shared/contestHelpers";
import {
  type ProblemSignalStat,
  PROBLEM_SUSPICION_TOOLTIP,
  problemSuspicionLevel,
  problemSuspicionLevelHint,
} from "@/features/contests/shared/problemSignalStats";

type Props = {
  problemStats: ProblemSignalStat[];
  className?: string;
  /** Компактная однострочная раскладка в карточке контеста. */
  variant?: "card";
};

export default function ContestProblemChips({
  problemStats,
  className,
  variant,
}: Props) {
  if (problemStats.length === 0) return null;

  const problemCount = problemStats.length;
  const compactAt = variant === "card" ? 6 : 8;
  const denseAt = variant === "card" ? 10 : 12;
  const density =
    problemCount >= denseAt ? "dense" : problemCount >= compactAt ? "compact" : null;

  return (
    <span
      className={
        `contest-stats__problems${density ? ` contest-stats__problems--${density}` : ""}` +
        (className ? ` ${className}` : "")
      }
      aria-label="Доля сигналов по задачам"
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
  );
}
