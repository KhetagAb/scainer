import { useMemo } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import type { UseQueryResult } from "@tanstack/react-query";
import type { ReportData, SubmissionListItem } from "@/client/types.gen";
import {
  formatSubmittedAt,
  prQueueForProblem,
  submissionIdsInFindings,
  findingsForSubmission,
  topFindingKey,
} from "@/features/review/reviewFindings";

type Props = {
  submissionsQuery: UseQueryResult<SubmissionListItem[]>;
  findingsQuery: UseQueryResult<unknown>;
  onUnauthorized: () => void;
};

export default function ReviewPage({
  submissionsQuery,
  findingsQuery,
  onUnauthorized,
}: Props) {
  const { id: contestId } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const problemId = searchParams.get("problem");

  const items = submissionsQuery.data ?? [];
  const report = findingsQuery.data as ReportData | undefined;
  const inFindings = useMemo(() => submissionIdsInFindings(report), [report]);

  const queue = useMemo(
    () => (problemId ? prQueueForProblem(items, problemId) : []),
    [items, problemId],
  );

  if (submissionsQuery.isLoading) {
    return <div className="page-center">Загрузка посылок…</div>;
  }
  if (submissionsQuery.isError) {
    if ((submissionsQuery.error as { status?: number })?.status === 401) onUnauthorized();
    return <div className="page-center login-error">Не удалось загрузить посылки</div>;
  }

  if (!contestId) return null;

  const base = `/contests/${encodeURIComponent(contestId)}`;

  if (!problemId) {
    return (
      <p className="review-empty">
        Выберите задачу сверху, чтобы открыть очередь PR-посылок.
      </p>
    );
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
    <div className="review-queue">
      <h2 className="review-queue__title">
        Очередь PR · {problemId}
        <span className="review-queue__count">{queue.length}</span>
      </h2>
      <ul className="review-queue__list">
        {queue.map((s) => {
          const hasFinding = inFindings.has(s.id);
          const findingKey = hasFinding
            ? topFindingKey(findingsForSubmission(report, s.id))
            : undefined;
          return (
            <li key={s.id} className="review-queue__item">
              <Link
                to={`${base}/review/${encodeURIComponent(s.id)}?problem=${encodeURIComponent(problemId)}`}
                className={`review-queue__row${hasFinding ? " has-finding" : ""}`}
              >
                <span className="review-queue__participant">{s.participant}</span>
                <span className="review-queue__lang">{s.lang}</span>
                <span className="review-queue__time">{formatSubmittedAt(s.submitted_at)}</span>
              </Link>
              {findingKey ? (
                <Link
                  to={`${base}/findings?finding=${encodeURIComponent(findingKey)}`}
                  className="review-queue__badge"
                  title="К подозрению"
                >
                  finding
                </Link>
              ) : null}
            </li>
          );
        })}
      </ul>
    </div>
  );
}
