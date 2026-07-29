import { Fragment, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { useQueryClient, type UseQueryResult } from "@tanstack/react-query";
import type { ProblemInfo, ReportData, SubmissionListItem } from "@/client/types.gen";
import {
  firstProblem,
  firstProblemWithPr,
  nextProblemWithPr,
  submissionPanelId,
  submissionsForProblemReview,
} from "@/features/review/reviewFindings";
import {
  matchesParticipantQuery,
  matchesVerdictFilter,
  isPrOnlyVerdictFilter,
  verdictFiltersEqual,
  type ReviewFiltersInput,
  type ReviewVerdictFilter,
} from "@/features/review/reviewFilterUtils";
import { problemDisplay } from "@/features/findings/reportModel";
import { ensureProblemComments } from "@/features/review/reviewPrefetch";
import { scrollToReviewPanel } from "@/features/review/reviewScroll";
import ReviewSubmissionPanel from "@/features/review/ReviewSubmissionPanel";
import ReviewShowAllSubmissions from "@/features/review/ReviewShowAllSubmissions";
import { useReviewActivePanel } from "@/features/review/useReviewActivePanel";

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
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const problemId = searchParams.get("problem");

  const items = submissionsQuery.data ?? [];
  const report = findingsQuery.data as ReportData | undefined;

  /** Очередь задачи на сессию: OK/RJ не убирают посылку до перезагрузки / смены задачи. */
  const [sessionProblemId, setSessionProblemId] = useState<string | null>(null);
  const [sessionQueue, setSessionQueue] = useState<SubmissionListItem[]>([]);
  const prevFiltersRef = useRef(reviewFilters);

  useEffect(() => {
    if (!problemId) {
      setSessionProblemId(null);
      setSessionQueue([]);
      return;
    }
    if (submissionsQuery.isLoading) return;

    const problemChanged = sessionProblemId !== problemId;
    const filtersChanged =
      !verdictFiltersEqual(
        prevFiltersRef.current.verdictFilter,
        reviewFilters.verdictFilter,
      ) ||
      prevFiltersRef.current.participantQuery !== reviewFilters.participantQuery;
    prevFiltersRef.current = reviewFilters;

    if (!problemChanged && !filtersChanged && sessionProblemId === problemId) return;

    setSessionQueue(submissionsForProblemReview(items, problemId, reviewFilters));
    setSessionProblemId(problemId);
  }, [problemId, sessionProblemId, items, submissionsQuery.isLoading, reviewFilters]);

  const queue = useMemo(() => {
    if (!problemId) return [];
    if (sessionProblemId === problemId) return sessionQueue;
    return submissionsForProblemReview(items, problemId, reviewFilters);
  }, [problemId, sessionProblemId, sessionQueue, items, reviewFilters]);

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

  const participantQuery = reviewFilters.participantQuery;

  const nextProblemId = useMemo(
    () =>
      problemId ? nextProblemWithPr(problems, items, problemId, participantQuery) : null,
    [problemId, problems, items, participantQuery],
  );

  const [isProblemTransitioning, setIsProblemTransitioning] = useState(false);

  const goToNextProblem = useCallback(async () => {
    if (!nextProblemId || !contestId || !problemId || isProblemTransitioning) return;

    setIsProblemTransitioning(true);
    try {
      const targetQueue = submissionsForProblemReview(items, nextProblemId, reviewFilters);
      await ensureProblemComments(queryClient, contestId, targetQueue);

      navigate(
        `/contests/${encodeURIComponent(contestId)}/review?problem=${encodeURIComponent(nextProblemId)}`,
        { replace: true },
      );
      if (window.location.hash) {
        window.history.replaceState(
          null,
          "",
          window.location.pathname + window.location.search,
        );
      }
      window.scrollTo(0, 0);
    } finally {
      setIsProblemTransitioning(false);
    }
  }, [
    nextProblemId,
    contestId,
    problemId,
    isProblemTransitioning,
    items,
    reviewFilters,
    queryClient,
    navigate,
  ]);

  const hasHiddenSubmissions = useMemo(() => {
    if (!problemId || !reviewFilters.verdictFilter.active) return false;
    return items.some(
      (s) =>
        s.problem === problemId &&
        !matchesVerdictFilter(s.verdict, reviewFilters.verdictFilter) &&
        matchesParticipantQuery(s.participant, reviewFilters.participantQuery),
    );
  }, [problemId, reviewFilters, items]);

  const hasHiddenForProblem = (pid: string) => {
    if (!reviewFilters.verdictFilter.active) return false;
    return items.some(
      (s) =>
        s.problem === pid &&
        !matchesVerdictFilter(s.verdict, reviewFilters.verdictFilter) &&
        matchesParticipantQuery(s.participant, reviewFilters.participantQuery),
    );
  };

  const showAllSubmissions = () =>
    setVerdictFilter({ ...reviewFilters.verdictFilter, active: false });

  useEffect(() => {
    if (problemId) return;
    if (submissionsQuery.isLoading) return;
    if (!contestId || !problems.length) return;

    const firstWithPr = firstProblemWithPr(problems, items, participantQuery);
    const target = firstWithPr ?? firstProblem(problems);
    if (!target) return;

    if (!firstWithPr && isPrOnlyVerdictFilter(reviewFilters.verdictFilter)) {
      setVerdictFilter({ ...reviewFilters.verdictFilter, active: false });
    }

    navigate(
      `/contests/${encodeURIComponent(contestId)}/review?problem=${encodeURIComponent(target)}`,
      { replace: true },
    );
  }, [
    problemId,
    submissionsQuery.isLoading,
    problems,
    items,
    contestId,
    navigate,
    reviewFilters,
    participantQuery,
    setVerdictFilter,
  ]);

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
        {hasHiddenSubmissions ? (
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
            <ReviewSubmissionPanel
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
