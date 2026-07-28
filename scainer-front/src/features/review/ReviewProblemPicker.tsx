import { FileText } from "lucide-react";
import { useState } from "react";
import { NavLink } from "react-router-dom";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
import { problemDisplay } from "@/features/findings/reportModel";
import { prCountByProblem } from "@/features/review/reviewFindings";
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
              const showStatement = active && statementAvailable;
              return (
                <li key={p.id} className="review-picker__item">
                  <NavLink
                    to={`${base}/review?problem=${encodeURIComponent(p.id)}`}
                    className={
                      "chip problem-chip review-picker__chip" +
                      (active ? " is-active" : "") +
                      (pr === 0 ? " is-empty" : "") +
                      (showStatement ? " is-statement" : "")
                    }
                    title={showStatement ? `Условие задачи — ${label}` : label}
                    aria-label={showStatement ? `Условие задачи — ${label}` : label}
                    onClick={(e) => {
                      if (showStatement) {
                        e.preventDefault();
                        setStatementOpen(true);
                      }
                    }}
                  >
                    <span className="review-picker__chip-main">
                      <span className="review-picker__chip-label">{label}</span>
                      {showStatement ? (
                        <span className="review-picker__chip-statement" aria-hidden>
                          <FileText size={14} strokeWidth={2} />
                          <span className="review-picker__chip-statement-label">
                            Условие задачи
                          </span>
                        </span>
                      ) : null}
                    </span>
                    <span className="review-picker__pr" aria-label={`${pr} pending review`}>
                      {pr}
                    </span>
                  </NavLink>
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
          problemLabel={activeProblemLabel}
          onClose={() => setStatementOpen(false)}
        />
      ) : null}
    </>
  );
}
