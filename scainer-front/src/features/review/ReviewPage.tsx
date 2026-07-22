import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import type { UseQueryResult } from "@tanstack/react-query";
import type { ProblemInfo, ReportData, SubmissionListItem } from "@/client/types.gen";
import {
  firstProblemWithPr,
  nextProblemWithPr,
  prQueueForProblem,
} from "@/features/review/reviewFindings";
import { problemDisplay } from "@/features/findings/reportModel";
import ReviewSubmissionPanel from "@/features/review/ReviewSubmissionPanel";

type Props = {
  submissionsQuery: UseQueryResult<SubmissionListItem[]>;
  findingsQuery: UseQueryResult<unknown>;
  problems: ProblemInfo[];
  onUnauthorized: () => void;
};

export default function ReviewPage({
  submissionsQuery,
  findingsQuery,
  problems,
  onUnauthorized,
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

  useEffect(() => {
    if (!problemId) {
      setSessionProblemId(null);
      setSessionQueue([]);
      return;
    }
    if (sessionProblemId === problemId) return;
    if (submissionsQuery.isLoading) return;
    setSessionQueue(prQueueForProblem(items, problemId));
    setSessionProblemId(problemId);
  }, [problemId, sessionProblemId, items, submissionsQuery.isLoading]);

  const queue = sessionProblemId === problemId ? sessionQueue : [];

  const problemLabel = useMemo(() => {
    if (!problemId) return "";
    const p = problems.find((x) => x.id === problemId);
    return problemDisplay(problemId, p?.name);
  }, [problemId, problems]);

  const nextProblemId = useMemo(
    () => (problemId ? nextProblemWithPr(problems, items, problemId) : null),
    [problemId, problems, items],
  );

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
          По задаче <strong>{problemId}</strong> нет посылок со статусом PR.
        </p>
      </div>
    );
  }

  return (
    <div className="review-stack">
      {queue.map((s, i) => (
        <ReviewSubmissionPanel
          key={s.id}
          contestId={contestId}
          submission={s}
          findingsReport={report}
          problemLabel={problemLabel}
          nextSubmissionId={queue[i + 1]?.id ?? null}
          nextProblemId={i === queue.length - 1 ? nextProblemId : null}
          onUnauthorized={onUnauthorized}
        />
      ))}
    </div>
  );
}
