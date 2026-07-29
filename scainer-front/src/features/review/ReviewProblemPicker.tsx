import { useEffect, useRef, useState } from "react";
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
  participantQuery?: string;
};

export default function ReviewProblemPicker({
  contestId,
  contestName,
  problems,
  submissions,
  activeProblemId,
  statementAvailable,
  participantQuery = "",
}: Props) {
  const [statementOpen, setStatementOpen] = useState(false);
  const [statementModalLabel, setStatementModalLabel] = useState<string | null>(null);
  const pickerRef = useRef<HTMLUListElement>(null);
  const prCounts = prCountByProblem(submissions, participantQuery);
  const sorted = problems.slice().sort((a, b) => {
    const la = problemDisplay(a.id, a.name);
    const lb = problemDisplay(b.id, b.name);
    return la < lb ? -1 : la > lb ? 1 : 0;
  });

  const activeIndex = activeProblemId
    ? sorted.findIndex((p) => p.id === activeProblemId)
    : -1;

  const activeProblem = activeProblemId
    ? sorted.find((p) => p.id === activeProblemId)
    : undefined;
  const activeProblemLabel = activeProblem?.name ?? activeProblemId ?? null;

  useEffect(() => {
    const el = pickerRef.current;
    if (!el || activeIndex < 0) return;

    const activeItem = el.querySelector<HTMLElement>(
      ".review-picker__item:has(.review-picker__chip.is-active)",
    );
    if (!activeItem) return;

    if (activeIndex === sorted.length - 1) {
      activeItem.scrollIntoView({ inline: "end", block: "nearest", behavior: "smooth" });
    } else {
      activeItem.scrollIntoView({ inline: "start", block: "nearest", behavior: "smooth" });
    }
  }, [activeIndex, sorted.length]);

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
          <ul ref={pickerRef} className="review-picker" aria-label="Задачи для ревью">
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
