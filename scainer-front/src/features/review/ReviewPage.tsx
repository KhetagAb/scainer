import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import type { UseQueryResult } from "@tanstack/react-query";
import type { ProblemInfo, ReportData, SubmissionListItem } from "@/client/types.gen";
import {
  firstProblemWithPr,
  nextProblemWithPr,
  submissionsForProblemReview,
} from "@/features/review/reviewFindings";
import { problemDisplay } from "@/features/findings/reportModel";
import { isPendingReview } from "@/features/review/reviewVerdicts";
import ReviewSubmissionPanel from "@/features/review/ReviewSubmissionPanel";
import { submissionPanelId } from "@/features/review/reviewFindings";
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

  const queue = sessionProblemId === problemId ? sessionQueue : [];

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

  const hasHiddenSubmissions = useMemo(() => {
    if (!problemId || !prOnly) return false;
    return items.some(
      (s) => s.problem === problemId && !isPendingReview(s.verdict),
    );
  }, [problemId, prOnly, items]);

  const showAllSubmissions = () => setReviewPrOnly(false);

  useEffect(() => {
    if (problemId) return;
    if (submissionsQuery.isLoading) return;
    const first = firstProblemWithPr(problems, items);
    if (!first || !contestId) return;
    navigate(
      `/contests/${encodeURIComponent(contestId)}/review?problem=${encodeURIComponent(first)}`,
      { replace: true },
    );
  }, [problemId, submissionsQuery.isLoading, problems, items, contestId, navigate]);

  useEffect(() => {
    const hash = window.location.hash.replace(/^#/, "");
    if (!hash || !queue.length) return;
    const t = window.setTimeout(() => {
      document.getElementById(hash)?.scrollIntoView({ block: "start" });
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
    if (firstProblemWithPr(problems, items) == null) {
      return (
        <div className="review-empty">
          <p>Нет посылок со статусом PR.</p>
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
            findingsReport={report}
            problemLabel={problemLabel}
            nextSubmissionId={queue[i + 1]?.id ?? null}
            nextProblemId={nextProblemId}
            onUnauthorized={onUnauthorized}
            showAllSubmissions={isLast && hasHiddenSubmissions && Boolean(nextProblemId)}
            onShowAllSubmissions={showAllSubmissions}
            isLastInStack={isLast}
            isActive={activePanelId === submissionPanelId(s.id)}
          />
        );
      })}
    </div>
  );
}
