import { useState } from "react";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
import { problemDisplay } from "@/features/findings/reportModel";
import { prCountByProblem } from "@/features/review/reviewFindings";
import ReviewProblemChip from "@/features/review/ReviewProblemChip";
import ProblemStatementModal from "@/features/statements/ProblemStatementModal";

type Props = {
  contestId: string;
  contestName?: string;
  problems: ProblemInfo[];
  submissions: SubmissionListItem[];
  activeProblemId: string | null;
  statementAvailable: boolean;
  prOnly: boolean;
  onPrOnlyChange: (value: boolean) => void;
};

export default function ReviewProblemPicker({
  contestId,
  contestName,
  problems,
  submissions,
  activeProblemId,
  statementAvailable,
  prOnly,
  onPrOnlyChange,
}: Props) {
  const [statementOpen, setStatementOpen] = useState(false);
  const [statementModalLabel, setStatementModalLabel] = useState<string | null>(null);
  const prCounts = prCountByProblem(submissions);
  const sorted = problems.slice().sort((a, b) => {
    const la = problemDisplay(a.id, a.name);
    const lb = problemDisplay(b.id, b.name);
    return la < lb ? -1 : la > lb ? 1 : 0;
  });

  const activeProblem = activeProblemId
    ? sorted.find((p) => p.id === activeProblemId)
    : undefined;
  const activeProblemLabel = activeProblem?.name ?? activeProblemId ?? null;

  if (!sorted.length) {
    return <p className="review-picker__empty">Задач пока нет</p>;
  }

  const base = `/contests/${encodeURIComponent(contestId)}`;

  const openStatement = (statementLabel: string) => {
    setStatementModalLabel(statementLabel);
    setStatementOpen(true);
  };

  const closeStatement = () => {
    setStatementOpen(false);
    setStatementModalLabel(null);
  };

  return (
    <>
      <div
        className={
          "review-picker-row" + (statementAvailable ? " review-picker-row--statements" : "")
        }
      >
        <div className="review-picker-scroll">
          <ul className="review-picker" aria-label="Задачи для ревью">
            {sorted.map((p) => {
              const pr = prCounts.get(p.id) ?? 0;
              const active = activeProblemId === p.id;
              const label = problemDisplay(p.id, p.name);
              return (
                <li key={p.id} className="review-picker__item">
                  <ReviewProblemChip
                    to={`${base}/review?problem=${encodeURIComponent(p.id)}`}
                    label={label}
                    statementLabel={p.name || p.id}
                    pr={pr}
                    active={active}
                    statementAvailable={statementAvailable}
                    onStatementOpen={openStatement}
                  />
                </li>
              );
            })}
          </ul>
        </div>

        <label className="review-picker__filter">
          <input
            type="checkbox"
            className="review-picker__filter-input"
            checked={prOnly}
            onChange={(e) => onPrOnlyChange(e.target.checked)}
          />
          <span className="review-picker__filter-label">PR only</span>
        </label>
      </div>

      {statementAvailable ? (
        <ProblemStatementModal
          open={statementOpen}
          contestId={contestId}
          title={contestName}
          problemLabel={statementModalLabel ?? activeProblemLabel}
          onClose={closeStatement}
        />
      ) : null}
    </>
  );
}
