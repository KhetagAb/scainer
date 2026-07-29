import { Fragment, useEffect, useMemo } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { useQueryClient, type UseQueryResult } from "@tanstack/react-query";
import type { ProblemInfo, ReportData, SubmissionListItem } from "@/client/types.gen";
import { problemDisplay } from "@/features/findings/reportModel";
import { submissionPanelId } from "@/features/review/reviewFindings";
import { indexSubmissionsByProblem } from "@/features/review/reviewModel";
import type { ReviewFiltersInput, ReviewVerdictFilter } from "@/features/review/reviewFilterUtils";
import { scrollToReviewPanel } from "@/features/review/reviewScroll";
import ReviewSubmissionPanelGate from "@/features/review/ReviewSubmissionPanelGate";
import ReviewShowAllSubmissions from "@/features/review/ReviewShowAllSubmissions";
import { useReviewActivePanel } from "@/features/review/useReviewActivePanel";
import { useReviewProblemNav } from "@/features/review/useReviewProblemNav";
import { useReviewProblemQueue } from "@/features/review/useReviewProblemQueue";

type Props = {
  submissionsQuery: UseQueryResult<SubmissionListItem[]>;
  findingsQuery: UseQueryResult<unknown>;
  problems: ProblemInfo[];
  onUnauthorized: () => void;
  reviewFilters: ReviewFiltersInput;
  setVerdictFilter: (value: ReviewVerdictFilter) => void;
};

export default function ReviewPage({
  submissionsQuery,
  findingsQuery,
  problems,
  onUnauthorized,
  reviewFilters,
  setVerdictFilter,
}: Props) {
  const { id: contestId } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const queryClient = useQueryClient();
  const problemId = searchParams.get("problem");

  const items = submissionsQuery.data ?? [];
  const report = findingsQuery.data as ReportData | undefined;

  const itemsByProblem = useMemo(() => indexSubmissionsByProblem(items), [items]);
  const problemItems = useMemo(
    () => (problemId ? itemsByProblem.get(problemId) ?? [] : []),
    [problemId, itemsByProblem],
  );

  const { queue, hiddenOnProblem } = useReviewProblemQueue({
    problemId,
    problemItems,
    reviewFilters,
    isLoading: submissionsQuery.isLoading,
  });

  const {
    nextProblemId,
    goToNextProblem,
    isProblemTransitioning,
    hasHiddenForProblem,
  } = useReviewProblemNav({
    contestId,
    problemId,
    problems,
    items,
    reviewFilters,
    submissionsLoading: submissionsQuery.isLoading,
    queryClient,
  });

  const panelIds = useMemo(
    () => queue.map((s) => submissionPanelId(s.id)),
    [queue],
  );
  const activePanelId = useReviewActivePanel(panelIds);
  const problemLabel = useMemo(() => {
    if (!problemId) return "";
    const p = problems.find((x) => x.id === problemId);
    return problemDisplay(problemId, p?.name);
  }, [problemId, problems]);

  const showAllSubmissions = () =>
    setVerdictFilter({ ...reviewFilters.verdictFilter, active: false });

  useEffect(() => {
    const hash = window.location.hash.replace(/^#/, "");
    if (!hash || !queue.length) return;
    const t = window.setTimeout(() => {
      const el = document.getElementById(hash);
      if (el) scrollToReviewPanel(el);
    }, 80);
    return () => window.clearTimeout(t);
  }, [queue, problemId]);

  const participantFilterActive = reviewFilters.participantQuery.trim().length > 0;

  if (submissionsQuery.isLoading) {
    return <div className="page-center">Загрузка посылок…</div>;
  }
  if (submissionsQuery.isError) {
    if ((submissionsQuery.error as { status?: number })?.status === 401) onUnauthorized();
    return <div className="page-center login-error">Не удалось загрузить посылки</div>;
  }

  if (!contestId) return null;

  if (!problemId) {
    if (!problems.length) {
      return (
        <div className="review-empty">
          <p>Задач пока нет.</p>
        </div>
      );
    }
    return <div className="page-center">Выбор задачи…</div>;
  }

  if (!queue.length) {
    return (
      <div className="review-empty">
        {hiddenOnProblem ? (
          <ReviewShowAllSubmissions
            className="review-show-all--empty"
            onClick={showAllSubmissions}
          />
        ) : participantFilterActive ? (
          <p>
            По задаче <strong>{problemId}</strong> нет посылок для участника «
            {reviewFilters.participantQuery.trim()}».
          </p>
        ) : (
          <p>
            По задаче <strong>{problemId}</strong> нет посылок.
          </p>
        )}
      </div>
    );
  }

  return (
    <div className="review-stack">
      {queue.map((s, i) => {
        const isLast = i === queue.length - 1;
        const showAllAfter = isLast && hasHiddenForProblem(problemId);
        return (
          <Fragment key={s.id}>
            <ReviewSubmissionPanelGate
              index={i}
              contestId={contestId}
              submission={s}
              allSubmissions={items}
              findingsReport={report}
              problemLabel={problemLabel}
              nextSubmissionId={queue[i + 1]?.id ?? null}
              nextProblemId={nextProblemId}
              onUnauthorized={onUnauthorized}
              onGoToNextProblem={goToNextProblem}
              problemTransitionPending={isProblemTransitioning}
              isActive={activePanelId === submissionPanelId(s.id)}
            />
            {showAllAfter ? (
              <ReviewShowAllSubmissions onClick={showAllSubmissions} />
            ) : null}
          </Fragment>
        );
      })}
    </div>
  );
}
