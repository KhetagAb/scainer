import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState } from "react";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
import { problemDisplay } from "@/features/findings/reportModel";
import { countByProblem, sortProblems } from "@/features/review/reviewModel";
import type { ReviewFiltersInput } from "@/features/review/reviewFilterUtils";
import ReviewParticipantFilter from "@/features/review/ReviewParticipantFilter";
import ReviewProblemChip from "@/features/review/ReviewProblemChip";
import ReviewVerdictFilterChip from "@/features/review/ReviewVerdictFilterChip";
import type {
  ReviewParticipantFilter as ReviewParticipantFilterValue,
  ReviewVerdictFilter,
} from "@/features/review/reviewTypes";
import { prefetchProblemStatementExplain } from "@/features/statements/explainQuery";
import ProblemStatementModal from "@/features/statements/ProblemStatementModal";

type Props = {
  contestId: string;
  contestName?: string;
  problems: ProblemInfo[];
  submissions: SubmissionListItem[];
  activeProblemId: string | null;
  statementAvailable: boolean;
  reviewFilters: ReviewFiltersInput;
  verdictFilter: ReviewVerdictFilter;
  onVerdictFilterChange: (value: ReviewVerdictFilter) => void;
  participantFilter: ReviewParticipantFilterValue;
  onParticipantFilterChange: (value: ReviewParticipantFilterValue) => void;
};

export default function ReviewProblemPicker({
  contestId,
  contestName,
  problems,
  submissions,
  activeProblemId,
  statementAvailable,
  reviewFilters,
  verdictFilter,
  onVerdictFilterChange,
  participantFilter,
  onParticipantFilterChange,
}: Props) {
  const queryClient = useQueryClient();
  const [statementOpen, setStatementOpen] = useState(false);
  const [statementModalProblemId, setStatementModalProblemId] = useState<string | null>(null);
  const [statementModalLabel, setStatementModalLabel] = useState<string | null>(null);
  const pickerRef = useRef<HTMLUListElement>(null);
  const [fadeStart, setFadeStart] = useState(false);
  const [fadeEnd, setFadeEnd] = useState(false);
  const submissionCounts = useMemo(
    () => countByProblem(submissions, reviewFilters),
    [submissions, reviewFilters],
  );
  const sorted = useMemo(() => sortProblems(problems), [problems]);

  const activeIndex = activeProblemId
    ? sorted.findIndex((p) => p.id === activeProblemId)
    : -1;

  const activeProblem = activeProblemId
    ? sorted.find((p) => p.id === activeProblemId)
    : undefined;
  const activeProblemLabel = activeProblem?.name ?? activeProblemId ?? null;

  useEffect(() => {
    const scrollEl = pickerRef.current;
    if (!scrollEl || activeIndex < 0) return;

    const activeItem = scrollEl.querySelector<HTMLElement>(
      ".review-picker__item:has(.review-picker__chip.is-active)",
    );
    if (!activeItem) return;

    // Минимальный сдвиг: не тянуть чип к краю, если он уже почти виден.
    activeItem.scrollIntoView({ inline: "nearest", block: "nearest", behavior: "smooth" });
  }, [activeIndex, activeProblemId, sorted.length]);

  useEffect(() => {
    const scrollEl = pickerRef.current;
    if (!scrollEl) return;

    const updateFade = () => {
      const { scrollLeft, scrollWidth, clientWidth } = scrollEl;
      setFadeStart(scrollLeft > 0);
      setFadeEnd(scrollLeft + clientWidth < scrollWidth - 1);
    };

    updateFade();
    scrollEl.addEventListener("scroll", updateFade, { passive: true });
    const ro = new ResizeObserver(updateFade);
    ro.observe(scrollEl);

    return () => {
      scrollEl.removeEventListener("scroll", updateFade);
      ro.disconnect();
    };
  }, [sorted.length]);

  if (!sorted.length) {
    return <p className="review-picker__empty">Задач пока нет</p>;
  }

  const base = `/contests/${encodeURIComponent(contestId)}`;

  const openStatement = (problemId: string, statementLabel: string) => {
    prefetchProblemStatementExplain(queryClient, contestId, problemId);
    setStatementModalProblemId(problemId);
    setStatementModalLabel(statementLabel);
    setStatementOpen(true);
  };

  const closeStatement = () => {
    setStatementOpen(false);
    setStatementModalProblemId(null);
    setStatementModalLabel(null);
  };

  return (
    <>
      <div
        className={
          "review-picker-row" + (statementAvailable ? " review-picker-row--statements" : "")
        }
      >
        <div
          className={[
            "review-picker-scroll",
            fadeStart ? "review-picker-scroll--fade-start" : "",
            fadeEnd ? "review-picker-scroll--fade-end" : "",
          ]
            .filter(Boolean)
            .join(" ")}
        >
          <div className="review-picker-scroll__viewport">
            <ul ref={pickerRef} className="review-picker" aria-label="Задачи для ревью">
              {sorted.map((p) => {
                const count = submissionCounts.get(p.id) ?? 0;
                const active = activeProblemId === p.id;
                const label = problemDisplay(p.id, p.name);
                return (
                  <li key={p.id} className="review-picker__item">
                    <ReviewProblemChip
                      to={`${base}/review?problem=${encodeURIComponent(p.id)}`}
                      label={label}
                      problemId={p.id}
                      statementLabel={p.name || p.id}
                      count={count}
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
        <div className="review-picker-filters">
          <ReviewParticipantFilter
            value={participantFilter}
            onChange={onParticipantFilterChange}
          />
          <ReviewVerdictFilterChip value={verdictFilter} onChange={onVerdictFilterChange} />
        </div>
      </div>

      {statementAvailable ? (
        <ProblemStatementModal
          open={statementOpen}
          contestId={contestId}
          title={contestName}
          problemId={statementModalProblemId ?? activeProblemId}
          problemLabel={statementModalLabel ?? activeProblemLabel}
          onClose={closeStatement}
        />
      ) : null}
    </>
  );
}
