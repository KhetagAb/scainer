import { useCallback, useEffect, useMemo, useRef, useState } from "react";
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
import { problemDisplay } from "@/features/findings/reportModel";
import { isPendingReview } from "@/features/review/reviewVerdicts";
import { ensureProblemComments } from "@/features/review/reviewPrefetch";
import { scrollToReviewPanel } from "@/features/review/reviewScroll";
import ReviewSubmissionPanel from "@/features/review/ReviewSubmissionPanel";
import { useReviewActivePanel } from "@/features/review/useReviewActivePanel";

function ReviewShowAllSubmissions({
  onClick,
}: {
  onClick: () => void;
}) {
  return (
    <div className="review-show-all">
      <button type="button" className="review-show-all__btn" onClick={onClick}>
        <span className="review-show-all__label">Показать все посылки</span>
        <svg
          width="28"
          height="28"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden
        >
          <path d="m6 9 6 6 6-6" />
        </svg>
      </button>
    </div>
  );
}

type Props = {
  submissionsQuery: UseQueryResult<SubmissionListItem[]>;
  findingsQuery: UseQueryResult<unknown>;
  problems: ProblemInfo[];
  onUnauthorized: () => void;
  prOnly: boolean;
  setReviewPrOnly: (value: boolean) => void;
};

export default function ReviewPage({
  submissionsQuery,
  findingsQuery,
  problems,
  onUnauthorized,
  prOnly,
  setReviewPrOnly,
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
  const prevPrOnlyRef = useRef(prOnly);

  useEffect(() => {
    if (!problemId) {
      setSessionProblemId(null);
      setSessionQueue([]);
      return;
    }
    if (submissionsQuery.isLoading) return;

    const problemChanged = sessionProblemId !== problemId;
    const prOnlyChanged = prevPrOnlyRef.current !== prOnly;
    prevPrOnlyRef.current = prOnly;

    if (!problemChanged && !prOnlyChanged && sessionProblemId === problemId) return;

    setSessionQueue(submissionsForProblemReview(items, problemId, prOnly));
    setSessionProblemId(problemId);
  }, [problemId, sessionProblemId, items, submissionsQuery.isLoading, prOnly]);

  const queue = useMemo(() => {
    if (!problemId) return [];
    if (sessionProblemId === problemId) return sessionQueue;
    return submissionsForProblemReview(items, problemId, prOnly);
  }, [problemId, sessionProblemId, sessionQueue, items, prOnly]);

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

  const nextProblemId = useMemo(
    () => (problemId ? nextProblemWithPr(problems, items, problemId) : null),
    [problemId, problems, items],
  );

  const [isProblemTransitioning, setIsProblemTransitioning] = useState(false);

  const goToNextProblem = useCallback(async () => {
    if (!nextProblemId || !contestId || !problemId || isProblemTransitioning) return;

    setIsProblemTransitioning(true);
    try {
      const targetQueue = submissionsForProblemReview(items, nextProblemId, prOnly);
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
    prOnly,
    queryClient,
    navigate,
  ]);

  const hasHiddenSubmissions = useMemo(() => {
    if (!problemId || !prOnly) return false;
    return items.some(
      (s) => s.problem === problemId && !isPendingReview(s.verdict),
    );
  }, [problemId, prOnly, items]);

  const hasHiddenForProblem = (pid: string) => {
    if (!prOnly) return false;
    return items.some((s) => s.problem === pid && !isPendingReview(s.verdict));
  };

  const showAllSubmissions = () => setReviewPrOnly(false);

  useEffect(() => {
    if (problemId) return;
    if (submissionsQuery.isLoading) return;
    if (!contestId || !problems.length) return;

    const firstWithPr = firstProblemWithPr(problems, items);
    const target = firstWithPr ?? firstProblem(problems);
    if (!target) return;

    if (!firstWithPr && prOnly) {
      setReviewPrOnly(false);
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
    prOnly,
    setReviewPrOnly,
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
        <p>
          {prOnly ? (
            <>
              По задаче <strong>{problemId}</strong> нет посылок со статусом PR.
            </>
          ) : (
            <>
              По задаче <strong>{problemId}</strong> нет посылок.
            </>
          )}
        </p>
        {hasHiddenSubmissions ? (
          <ReviewShowAllSubmissions onClick={showAllSubmissions} />
        ) : null}
      </div>
    );
  }

  return (
    <div className="review-stack">
      {queue.map((s, i) => {
        const isLast = i === queue.length - 1;
        return (
          <ReviewSubmissionPanel
            key={s.id}
            contestId={contestId}
            submission={s}
            allSubmissions={items}
            findingsReport={report}
            problemLabel={problemLabel}
            nextSubmissionId={queue[i + 1]?.id ?? null}
            nextProblemId={nextProblemId}
            onUnauthorized={onUnauthorized}
            showAllSubmissions={
              isLast && hasHiddenForProblem(problemId) && Boolean(nextProblemId)
            }
            onShowAllSubmissions={showAllSubmissions}
            onGoToNextProblem={goToNextProblem}
            problemTransitionPending={isProblemTransitioning}
            isActive={activePanelId === submissionPanelId(s.id)}
          />
        );
      })}
    </div>
  );
}
