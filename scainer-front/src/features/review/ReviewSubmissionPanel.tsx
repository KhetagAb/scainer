import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type FormEvent,
  type KeyboardEvent,
} from "react";
import { Link, useNavigate } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  getSubmissionCommentsOptions,
  getSubmissionCommentsQueryKey,
  postSubmissionCommentMutation,
  postSubmissionVerdictMutation,
} from "@/client/@tanstack/react-query.gen";
import type { ReportData, SubmissionListItem } from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";
import {
  findingsForSubmission,
  formatSubmittedAt,
  submissionPanelId,
  topFindingKey,
} from "@/features/review/reviewFindings";
import SourceCode from "@/features/code/SourceCode";
import { CodePaneMoreBar } from "@/features/code/CodePaneMoreBar";
import ReviewCelebrateOverlay from "@/features/review/ReviewCelebrateOverlay";
import { EjudgeContestChip } from "@/features/ejudge/EjudgeContestChip";
import { RadarAttentionIcon } from "@/features/review/RadarAttentionIcon";

function citeLinesLabel(from: number, to: number): string {
  return from === to ? `Строка ${from}: ` : `Строки ${from}–${to}: `;
}

function verdictChipTone(verdict: string): "ok" | "pr" | "fail" | "neutral" {
  const v = verdict.trim().toUpperCase();
  if (!v || v === "—" || v === "-") return "neutral";
  if (v === "OK" || v === "AC") return "ok";
  if (v === "PR" || v === "PD") return "pr";
  return "fail";
}

function formatVerdictLabel(verdict: string): string {
  const v = verdict.trim().toUpperCase();
  if (v === "PR" || v === "PD") return "Pending Review";
  if (v === "OK" || v === "AC") return "Accepted";
  if (v === "RJ") return "Rejected";
  return verdict.trim() || "—";
}

function ChevronDownIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
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
  );
}

function ArrowRightIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M5 12h14" />
      <path d="m12 5 7 7-7 7" />
    </svg>
  );
}

function CheckIcon({ size = 18 }: { size?: number }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M20 6 9 17l-5-5" />
    </svg>
  );
}

function XIcon({ size = 18 }: { size?: number }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M18 6 6 18" />
      <path d="m6 6 12 12" />
    </svg>
  );
}

type Props = {
  contestId: string;
  submission: SubmissionListItem;
  findingsReport: ReportData | undefined;
  problemLabel: string;
  nextSubmissionId: string | null;
  nextProblemId: string | null;
  onUnauthorized: () => void;
};

export default function ReviewSubmissionPanel({
  contestId,
  submission,
  findingsReport,
  problemLabel,
  nextSubmissionId,
  nextProblemId,
  onUnauthorized,
}: Props) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const commentRef = useRef<HTMLTextAreaElement>(null);
  const [comment, setComment] = useState("");
  const [actionError, setActionError] = useState<string | null>(null);
  const [commentsOpen, setCommentsOpen] = useState(true);
  const [celebrate, setCelebrate] = useState(false);

  const submissionId = submission.id;
  const panelId = submissionPanelId(submissionId);
  const commentFieldId = `review-comment-${panelId}`;
  const goNextProblem = !nextSubmissionId && Boolean(nextProblemId);
  const nextEnabled = Boolean(nextSubmissionId || nextProblemId);

  useEffect(() => {
    setCommentsOpen(true);
    setComment("");
    setActionError(null);
  }, [submissionId]);

  const insertCite = useCallback((from: number, to: number) => {
    const el = commentRef.current;
    const label = citeLinesLabel(from, to);
    if (!el) {
      setComment((prev) => {
        if (!prev) return label;
        const needsNl = !prev.endsWith("\n");
        return needsNl ? `${prev}\n${label}` : `${prev}${label}`;
      });
      return;
    }

    const start = el.selectionStart;
    const end = el.selectionEnd;
    const value = el.value;
    const atLineStart = start === 0 || value[start - 1] === "\n";
    const prefix = atLineStart ? "" : "\n";
    const insert = prefix + label;
    const next = value.slice(0, start) + insert + value.slice(end);
    const caret = start + insert.length;
    setComment(next);
    requestAnimationFrame(() => {
      el.focus();
      el.setSelectionRange(caret, caret);
    });
  }, []);

  const commentsQuery = useQuery({
    ...getSubmissionCommentsOptions({
      path: { id: contestId, submissionId },
      headers: authHeaders(),
    }),
  });

  const verdictMutation = useMutation(postSubmissionVerdictMutation());
  const commentMutation = useMutation(postSubmissionCommentMutation());
  const actionPending = verdictMutation.isPending || commentMutation.isPending;

  const findingKey = useMemo(
    () => topFindingKey(findingsForSubmission(findingsReport, submissionId)),
    [findingsReport, submissionId],
  );

  if (commentsQuery.isError) {
    if ((commentsQuery.error as { status?: number })?.status === 401) onUnauthorized();
  }

  const liveVerdict = commentsQuery.data?.verdict ?? submission.verdict ?? "—";
  const verdictTone = verdictChipTone(liveVerdict);
  const commentText = comment.trim();
  const thread = useMemo(() => {
    const comments = commentsQuery.data?.comments ?? [];
    return [...comments].sort(
      (a, b) => new Date(b.time).getTime() - new Date(a.time).getTime(),
    );
  }, [commentsQuery.data?.comments]);
  const commentsCollapsed = !commentsOpen && thread.length > 1;
  const base = `/contests/${encodeURIComponent(contestId)}`;

  const refreshComments = () =>
    queryClient.invalidateQueries({
      queryKey: getSubmissionCommentsQueryKey({
        path: { id: contestId, submissionId },
        headers: authHeaders(),
      }),
    });

  const scrollToNext = () => {
    if (!nextSubmissionId) return;
    const el = document.getElementById(submissionPanelId(nextSubmissionId));
    el?.scrollIntoView({ behavior: "smooth", block: "start" });
  };

  const goNext = () => {
    if (nextSubmissionId) {
      scrollToNext();
      return;
    }
    if (nextProblemId) {
      navigate(
        `/contests/${encodeURIComponent(contestId)}/review?problem=${encodeURIComponent(nextProblemId)}`,
      );
    }
  };

  const submitComment = async () => {
    if (!commentText) return;
    setActionError(null);
    try {
      await commentMutation.mutateAsync({
        path: { id: contestId, submissionId },
        body: { text: commentText },
        headers: authHeaders(),
      });
      setComment("");
      await refreshComments();
    } catch (err) {
      if ((err as { status?: number })?.status === 401) onUnauthorized();
      const detail =
        (err as { error?: string })?.error ||
        (err as { message?: string })?.message ||
        "Не удалось отправить комментарий";
      setActionError(detail);
    }
  };

  const submitVerdict = async (verdict: "OK" | "RJ") => {
    setActionError(null);
    try {
      await verdictMutation.mutateAsync({
        path: { id: contestId, submissionId },
        body: {
          verdict,
          ...(commentText ? { comment: commentText } : {}),
        },
        headers: authHeaders(),
      });
      setComment("");
      await refreshComments();
      if (nextSubmissionId) scrollToNext();
      else setCelebrate(true);
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
    goNext();
  };

  const expandComments = () => setCommentsOpen(true);

  const onCommentsKeyDown = (e: KeyboardEvent<HTMLElement>) => {
    if (!commentsCollapsed) return;
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      expandComments();
    }
  };

  const commentsList = (
    <ul className="review-comments__list">
      {thread.map((c) => (
        <li key={c.id} className="review-comments__item">
          <div className="review-comments__meta">
            <span className="review-comments__author">{c.from}:</span>
            <time className="review-comments__time" dateTime={c.time}>
              {formatSubmittedAt(c.time)}
            </time>
          </div>
          {c.subject ? <div className="review-comments__subject">{c.subject}</div> : null}
          <pre className="review-comments__text">{c.text}</pre>
        </li>
      ))}
    </ul>
  );

  return (
    <article id={panelId} className="review-detail review-detail--panel">
      <ReviewCelebrateOverlay
        open={celebrate}
        onClose={() => setCelebrate(false)}
        problemLabel={problemLabel}
        nextProblemId={nextProblemId}
        onNextProblem={goNext}
      />
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

      <div className="review-workspace">
        <div className="review-workspace__label-row">
          <h3 className="review-workspace__label">Код</h3>
          <CodePaneMoreBar key={submissionId} className="review-workspace__more">
            <EjudgeContestChip title={submissionId}>{submissionId}</EjudgeContestChip>
            {submission.problem ? (
              <span className="chip problem-chip" title={submission.problem}>
                {submission.problem}
              </span>
            ) : null}
            {submission.lang ? <span className="chip">{submission.lang}</span> : null}
          </CodePaneMoreBar>
        </div>
        <div className="review-workspace__label-spacer" aria-hidden />

        <div className="review-workspace__code" aria-label="Исходный код">
          {findingKey ? (
            <Link
              to={`${base}/findings?finding=${encodeURIComponent(findingKey)}`}
              className="review-findings-link review-findings-link--overlay"
              aria-label="К подозрениям"
            >
              <RadarAttentionIcon />
              <span className="review-findings-link__label">
                <span className="review-findings-link__text">ATTENTION!</span>
                <span className="review-findings-link__subtext">К подозрениям →</span>
              </span>
            </Link>
          ) : null}
          {commentsQuery.isLoading ? (
            <p className="review-comments__empty">Загрузка…</p>
          ) : (
            <SourceCode
              className={`review-source__code${
                verdictTone === "ok"
                  ? " review-source__code--ok"
                  : verdictTone === "fail"
                    ? " review-source__code--rj"
                    : findingKey
                      ? " review-source__code--attention"
                      : ""
              }`}
              lang={submission.lang}
              lines={
                commentsQuery.data?.source?.length
                  ? commentsQuery.data.source
                  : ["(нет исходника)"]
              }
              onLineNumberClick={(lineNo) => insertCite(lineNo, lineNo)}
              onCiteLines={insertCite}
            />
          )}
        </div>

        <aside className="review-workspace__side">
          <div className="review-side-head">
            <span className="review-side-head__name">{submission.participant}</span>
            <span className={`review-verdict-chip review-verdict-chip--${verdictTone}`}>
              {formatVerdictLabel(liveVerdict)}
            </span>
          </div>
          <form
            className="review-verdict"
            onSubmit={(e) => {
              e.preventDefault();
            }}
          >
            <textarea
              ref={commentRef}
              id={commentFieldId}
              className="review-verdict__input"
              rows={5}
              placeholder="Комментарий"
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              disabled={actionPending}
              aria-label="Комментарий"
            />
            {actionError ? <p className="review-verdict__error">{actionError}</p> : null}
            <div className="review-verdict__actions">
              <div className="review-verdict__ok-rj">
                <button
                  type="button"
                  className="btn btn--ok btn--icon"
                  disabled={actionPending}
                  aria-label="OK"
                  onClick={() => void submitVerdict("OK")}
                >
                  <CheckIcon />
                </button>
                <button
                  type="button"
                  className="btn btn--rj btn--icon"
                  disabled={actionPending}
                  aria-label="RJ"
                  onClick={() => void submitVerdict("RJ")}
                >
                  <XIcon />
                </button>
              </div>
              <button
                type="button"
                className="btn btn--comment"
                disabled={actionPending || !commentText}
                onClick={() => void submitComment()}
              >
                Comment
              </button>
              {nextEnabled ? (
                <button
                  type="button"
                  className="review-verdict__next"
                  disabled={actionPending}
                  onClick={onManualNext}
                >
                  {goNextProblem ? "Следующая задача" : "Следующая посылка"}
                  {goNextProblem ? <ArrowRightIcon /> : <ChevronDownIcon />}
                </button>
              ) : null}
            </div>
          </form>

          {thread.length > 0 ? (
            <section className="review-comments" aria-label="Комментарии">
              <h3 className="review-comments__title">
                Комментарии
                <span className="review-comments__count">{thread.length}</span>
              </h3>
              {commentsCollapsed ? (
                <div
                  className="review-comments__preview"
                  role="button"
                  tabIndex={0}
                  aria-expanded={false}
                  aria-label={`Показать все комментарии, ещё ${thread.length - 1}`}
                  onClick={expandComments}
                  onKeyDown={onCommentsKeyDown}
                >
                  <div className="review-comments__clip">{commentsList}</div>
                  <span className="review-comments__more">
                    Ещё {thread.length - 1}
                    <svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden>
                      <path
                        d="M2.5 4.5 6 8l3.5-3.5"
                        stroke="currentColor"
                        strokeWidth="1.4"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  </span>
                </div>
              ) : (
                commentsList
              )}
            </section>
          ) : null}
        </aside>
      </div>
    </article>
  );
}
