import { useMemo, useState, type FormEvent } from "react";
import { Link, useNavigate, useParams, useSearchParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient, type UseQueryResult } from "@tanstack/react-query";
import {
  getContestSubmissionsOptions,
  getSubmissionCommentsOptions,
  getSubmissionCommentsQueryKey,
  postSubmissionVerdictMutation,
} from "@/client/@tanstack/react-query.gen";
import type { ReportData, SubmissionListItem } from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";
import {
  findingsForSubmission,
  formatSubmittedAt,
  nextPrInQueue,
  prQueueForProblem,
  topFindingKey,
} from "@/features/review/reviewFindings";

type Props = {
  submissionsQuery: UseQueryResult<SubmissionListItem[]>;
  findingsQuery: UseQueryResult<unknown>;
  onUnauthorized: () => void;
};

export default function ReviewSubmissionPage({
  submissionsQuery,
  findingsQuery,
  onUnauthorized,
}: Props) {
  const { id: contestId, submissionId: rawSubmissionId } = useParams<{
    id: string;
    submissionId: string;
  }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [comment, setComment] = useState("");
  const [actionError, setActionError] = useState<string | null>(null);

  const submissionId = rawSubmissionId ? decodeURIComponent(rawSubmissionId) : "";
  const items = submissionsQuery.data ?? [];
  const fromList = items.find((s) => s.id === submissionId);
  const problemId = searchParams.get("problem") || fromList?.problem || "";

  const commentsQuery = useQuery({
    ...getSubmissionCommentsOptions({
      path: { id: contestId ?? "", submissionId },
      headers: authHeaders(),
    }),
    enabled: Boolean(contestId && submissionId),
  });

  const verdictMutation = useMutation(postSubmissionVerdictMutation());

  const report = findingsQuery.data as ReportData | undefined;
  const findingKey = useMemo(
    () => topFindingKey(findingsForSubmission(report, submissionId)),
    [report, submissionId],
  );

  const queue = useMemo(
    () => (problemId ? prQueueForProblem(items, problemId) : []),
    [items, problemId],
  );

  if (!contestId || !submissionId) return null;

  const base = `/contests/${encodeURIComponent(contestId)}`;
  const backTo = problemId
    ? `${base}/review?problem=${encodeURIComponent(problemId)}`
    : `${base}/review`;

  if (commentsQuery.isError) {
    if ((commentsQuery.error as { status?: number })?.status === 401) onUnauthorized();
  }

  const meta = fromList;
  const liveVerdict = commentsQuery.data?.verdict ?? meta?.verdict ?? "—";

  const goNext = (freshItems: SubmissionListItem[]) => {
    const nextQueue = problemId ? prQueueForProblem(freshItems, problemId) : [];
    const next = nextPrInQueue(nextQueue, submissionId);
    if (next) {
      navigate(
        `${base}/review/${encodeURIComponent(next.id)}?problem=${encodeURIComponent(problemId)}`,
      );
      setComment("");
      setActionError(null);
      return;
    }
    navigate(backTo);
  };

  const submitVerdict = async (verdict: "OK" | "RJ") => {
    setActionError(null);
    try {
      await verdictMutation.mutateAsync({
        path: { id: contestId, submissionId },
        body: {
          verdict,
          ...(comment.trim() ? { comment: comment.trim() } : {}),
        },
        headers: authHeaders(),
      });
      await queryClient.invalidateQueries({
        queryKey: getSubmissionCommentsQueryKey({
          path: { id: contestId, submissionId },
          headers: authHeaders(),
        }),
      });
      const refreshed = await queryClient.fetchQuery({
        ...getContestSubmissionsOptions({
          path: { id: contestId },
          headers: authHeaders(),
        }),
      });
      goNext(refreshed ?? []);
    } catch (err) {
      if ((err as { status?: number })?.status === 401) onUnauthorized();
      const detail =
        (err as { error?: string })?.error ||
        (err as { message?: string })?.message ||
        "Не удалось выставить вердикт";
      setActionError(detail);
    }
  };

  const onManualNext = (e: FormEvent) => {
    e.preventDefault();
    goNext(items);
  };

  return (
    <div className="review-detail">
      <Link to={backTo} className="back-link">
        ← К очереди{problemId ? ` · ${problemId}` : ""}
      </Link>

      <header className="review-detail__head">
        <h2 className="review-detail__title">
          {meta?.participant ?? "Посылка"}
          <span className="review-detail__verdict">{liveVerdict}</span>
        </h2>
        <div className="review-detail__meta">
          {problemId ? <span>задача {problemId}</span> : null}
          {meta?.lang ? <span>{meta.lang}</span> : null}
          {meta?.submitted_at ? (
            <span>{formatSubmittedAt(meta.submitted_at)}</span>
          ) : null}
          {findingKey ? (
            <Link
              to={`${base}/findings?finding=${encodeURIComponent(findingKey)}`}
              className="review-detail__finding"
            >
              К подозрению
            </Link>
          ) : null}
        </div>
      </header>

      {commentsQuery.data?.status_stale ? (
        <p className="review-detail__warn">
          Статус в ejudge не обновлён
          {commentsQuery.data.status_error
            ? `: ${commentsQuery.data.status_error}`
            : ""}
        </p>
      ) : null}
      {commentsQuery.data?.comments_error ? (
        <p className="review-detail__warn">
          Комментарии: {commentsQuery.data.comments_error}
        </p>
      ) : null}

      <section className="review-source" aria-label="Исходный код">
        <h3 className="review-source__title">Код</h3>
        {commentsQuery.isLoading ? (
          <p className="review-comments__empty">Загрузка…</p>
        ) : (
          <pre className="code-lines review-source__code">
            {(commentsQuery.data?.source?.length
              ? commentsQuery.data.source
              : ["(нет исходника)"]
            ).map((line, idx) => (
              <code key={idx} className="line">
                {line}
              </code>
            ))}
          </pre>
        )}
      </section>

      <section className="review-comments" aria-label="Комментарии">
        <h3 className="review-comments__title">Комментарии</h3>
        {commentsQuery.isLoading ? (
          <p className="review-comments__empty">Загрузка…</p>
        ) : (commentsQuery.data?.comments?.length ?? 0) === 0 ? (
          <p className="review-comments__empty">Пока нет комментариев</p>
        ) : (
          <ul className="review-comments__list">
            {(commentsQuery.data?.comments ?? []).map((c) => (
              <li key={c.id} className="review-comments__item">
                <div className="review-comments__meta">
                  <span>{c.from}</span>
                  <time dateTime={c.time}>{formatSubmittedAt(c.time)}</time>
                </div>
                {c.subject ? (
                  <div className="review-comments__subject">{c.subject}</div>
                ) : null}
                <pre className="review-comments__text">{c.text}</pre>
              </li>
            ))}
          </ul>
        )}
      </section>

      <form
        className="review-verdict"
        onSubmit={(e) => {
          e.preventDefault();
        }}
      >
        <label className="review-verdict__label" htmlFor="review-comment">
          Комментарий к вердикту (необязательно)
        </label>
        <textarea
          id="review-comment"
          className="review-verdict__input"
          rows={4}
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          disabled={verdictMutation.isPending}
        />
        {actionError ? <p className="review-verdict__error">{actionError}</p> : null}
        <div className="review-verdict__actions">
          <button
            type="button"
            className="btn btn--primary"
            disabled={verdictMutation.isPending}
            onClick={() => void submitVerdict("OK")}
          >
            OK
          </button>
          <button
            type="button"
            className="btn btn--ghost review-verdict__rj"
            disabled={verdictMutation.isPending}
            onClick={() => void submitVerdict("RJ")}
          >
            RJ
          </button>
          <button
            type="button"
            className="btn btn--ghost"
            disabled={verdictMutation.isPending || !nextPrInQueue(queue, submissionId)}
            onClick={onManualNext}
          >
            Следующая
          </button>
        </div>
      </form>
    </div>
  );
}
