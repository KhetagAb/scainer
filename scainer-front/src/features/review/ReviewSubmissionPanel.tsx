import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
} from "react";
import { Link, useNavigate } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  getSubmissionCommentsOptions,
  getSubmissionCommentsQueryKey,
  getContestProblemsQueryKey,
  getContestSubmissionsQueryKey,
  postSubmissionCommentMutation,
  postSubmissionVerdictMutation,
} from "@/client/@tanstack/react-query.gen";
import type { ProblemInfo, ReportData, SubmissionListItem } from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { ApiError } from "@/lib/apiError";
import {
  findingsForSubmission,
  formatSubmittedAt,
  prCountByProblem,
  submissionPanelId,
  topFindingKey,
} from "@/features/review/reviewFindings";
import { citeLinesLabel, tryMergeCiteAt } from "@/features/review/reviewCite";
import {
  formatVerdictLabel,
  isPendingReview,
  isPrVerdict,
  verdictChipTone,
} from "@/features/review/reviewVerdicts";
import SourceCode from "@/features/code/SourceCode";
import SourceCodeCopyButton from "@/features/code/SourceCodeCopyButton";
import ReviewCelebrateOverlay from "@/features/review/ReviewCelebrateOverlay";
import { EjudgeContestChip } from "@/features/ejudge/EjudgeContestChip";
import ProblemStatementScrollZone from "@/features/statements/ProblemStatementScrollZone";
import { RadarAttentionIcon } from "@/features/review/RadarAttentionIcon";
import { useReviewSideRail } from "@/features/review/sideRail/useReviewSideRail";
import { ReviewSideRailDebugHud } from "@/features/review/sideRail/ReviewSideRailDebug";

function ChevronDownIcon({ size = 16 }: { size?: number }) {
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
      <path d="m6 9 6 6 6-6" />
    </svg>
  );
}

function ChevronRightIcon({ size = 16 }: { size?: number }) {
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
      <path d="m9 6 6 6-6 6" />
    </svg>
  );
}

const NEXT_SUBMISSION_SCROLL_MS = 700;

function smoothScrollToBlockStart(el: HTMLElement, duration = NEXT_SUBMISSION_SCROLL_MS) {
  const startY = window.scrollY;
  const scrollMarginTop = parseFloat(getComputedStyle(el).scrollMarginTop) || 0;
  const targetY = el.getBoundingClientRect().top + window.scrollY - scrollMarginTop;
  const distance = targetY - startY;
  if (Math.abs(distance) < 1) return;

  const start = performance.now();
  const tick = (now: number) => {
    const t = Math.min((now - start) / duration, 1);
    const eased = 1 - (1 - t) ** 3;
    window.scrollTo(0, startY + distance * eased);
    if (t < 1) requestAnimationFrame(tick);
  };
  requestAnimationFrame(tick);
}

type Props = {
  contestId: string;
  submission: SubmissionListItem;
  findingsReport: ReportData | undefined;
  problemLabel: string;
  statementAvailable: boolean;
  contestName: string;
  statementProblemLabel: string | null;
  nextSubmissionId: string | null;
  nextProblemId: string | null;
  onUnauthorized: () => void;
  showAllSubmissions?: boolean;
  onShowAllSubmissions?: () => void;
  isLastInStack?: boolean;
  isActive?: boolean;
};

export default function ReviewSubmissionPanel({
  contestId,
  submission,
  findingsReport,
  problemLabel,
  statementAvailable,
  contestName,
  statementProblemLabel,
  nextSubmissionId,
  nextProblemId,
  onUnauthorized,
  showAllSubmissions = false,
  onShowAllSubmissions,
  isLastInStack: _isLastInStack = false,
  isActive = true,
}: Props) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const commentRef = useRef<HTMLTextAreaElement>(null);
  const commentsScrollRef = useRef<HTMLDivElement>(null);
  const listContentRef = useRef<HTMLUListElement>(null);
  const codeRef = useRef<HTMLDivElement>(null);
  const headRef = useRef<HTMLDivElement>(null);
  const footRef = useRef<HTMLDivElement>(null);
  const bodyRef = useRef<HTMLElement>(null);
  const metaRef = useRef<HTMLDivElement>(null);
  const formRef = useRef<HTMLFormElement>(null);
  const [comment, setComment] = useState("");
  const [actionError, setActionError] = useState<string | null>(null);
  const [celebrate, setCelebrate] = useState(false);
  const [statementOpen, setStatementOpen] = useState(false);

  const submissionId = submission.id;
  const panelId = submissionPanelId(submissionId);
  const commentFieldId = `review-comment-${panelId}`;

  const goNextProblem = !nextSubmissionId && Boolean(nextProblemId);
  const nextEnabled = Boolean(nextSubmissionId || nextProblemId);
  const showAllBesideNext =
    showAllSubmissions && goNextProblem && Boolean(onShowAllSubmissions);

  useEffect(() => {
    setComment("");
    setActionError(null);
  }, [submissionId]);

  const insertCite = useCallback((from: number, to: number) => {
    const el = commentRef.current;
    const label = citeLinesLabel(from, to);

    if (!el) {
      setComment((prev) => {
        const merged = tryMergeCiteAt(prev, prev.length, from, to);
        if (merged) return merged.newValue;
        if (!prev) return label;
        const needsNl = !prev.endsWith("\n");
        return needsNl ? `${prev}\n${label}` : `${prev}${label}`;
      });
      return;
    }

    const start = el.selectionStart;
    const end = el.selectionEnd;
    const value = el.value;

    const merged = tryMergeCiteAt(value, start, from, to);
    if (merged) {
      setComment(merged.newValue);
      requestAnimationFrame(() => {
        el.focus();
        el.setSelectionRange(merged.caret, merged.caret);
      });
      return;
    }

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
    if (commentsQuery.error instanceof ApiError && commentsQuery.error.status === 401) onUnauthorized();
  }

  const liveVerdict = commentsQuery.data?.verdict ?? submission.verdict ?? "—";
  const verdictTone = verdictChipTone(liveVerdict);
  const pendingVerdictActions = isPendingReview(liveVerdict);
  const prVerdict = isPrVerdict(liveVerdict);
  const verdictReviewedFromPr = useMemo(() => {
    const v = liveVerdict.trim().toUpperCase();
    const reviewed = v === "OK" || v === "AC" || v === "RJ";
    return reviewed && isPendingReview(submission.verdict);
  }, [liveVerdict, submission.verdict]);
  const commentText = comment.trim();
  const thread = useMemo(() => {
    const comments = commentsQuery.data?.comments ?? [];
    return [...comments].sort(
      (a, b) => new Date(b.time).getTime() - new Date(a.time).getTime(),
    );
  }, [commentsQuery.data?.comments]);
  const sourceText = useMemo(() => {
    const lines = commentsQuery.data?.source;
    if (!lines?.length) return "";
    if (lines.length === 1 && lines[0] === "(нет исходника)") return "";
    return lines.join("\n");
  }, [commentsQuery.data?.source]);
  const base = `/contests/${encodeURIComponent(contestId)}`;

  const sideRailContentKey = [
    submissionId,
    thread.length > 0 ? "comments" : "no-comments",
    comment.length,
    nextEnabled ? "next" : "no-next",
    showAllBesideNext ? "all" : "no-all",
  ].join("|");

  const {
    railRef,
    ready: railReady,
    mode: railMode,
    scrollComments,
    gapEff,
    debugSnapshot,
  } = useReviewSideRail({
    codeRef,
    headRef,
    footRef,
    bodyRef,
    metaRef,
    formRef,
    listContentRef,
    listScrollRef: commentsScrollRef,
    contentKey: sideRailContentKey,
  });

  const gapsInteractive = railMode === "pinned" && gapEff > 0;

  useEffect(() => {
    const el = commentsScrollRef.current;
    if (!el) return;

    const syncScrollHint = () => {
      const scrollable = el.scrollHeight > el.clientHeight + 1;
      el.classList.toggle("is-scrollable", scrollable);
      el.classList.toggle("is-scrolled-top", scrollable && el.scrollTop > 1);
      el.classList.toggle(
        "is-scrolled-bottom",
        scrollable && el.scrollTop + el.clientHeight < el.scrollHeight - 1,
      );
    };

    syncScrollHint();
    const ro = new ResizeObserver(syncScrollHint);
    ro.observe(el);
    el.addEventListener("scroll", syncScrollHint, { passive: true });
    return () => {
      ro.disconnect();
      el.removeEventListener("scroll", syncScrollHint);
    };
  }, [thread.length, submissionId]);

  const withCommentsScroll = async (fn: () => Promise<void>) => {
    const el = commentsScrollRef.current;
    const scrollTop = el?.scrollTop ?? 0;
    await fn();
    requestAnimationFrame(() => {
      if (el) el.scrollTop = scrollTop;
    });
  };

  const refreshComments = () =>
    queryClient.invalidateQueries({
      queryKey: getSubmissionCommentsQueryKey({
        path: { id: contestId, submissionId },
        headers: authHeaders(),
      }),
    });

  const patchSubmissionVerdict = (verdict: "OK" | "RJ") => {
    const wasPr = isPrVerdict(submission.verdict);

    queryClient.setQueryData<SubmissionListItem[]>(
      getContestSubmissionsQueryKey({
        path: { id: contestId },
        headers: authHeaders(),
      }),
      (items) =>
        items?.map((s) => (s.id === submissionId ? { ...s, verdict } : s)),
    );

    if (wasPr) {
      queryClient.setQueryData<ProblemInfo[]>(
        getContestProblemsQueryKey({
          path: { id: contestId },
          headers: authHeaders(),
        }),
        (problems) =>
          problems?.map((p) =>
            p.id === submission.problem
              ? { ...p, pendingCount: Math.max(0, (p.pendingCount ?? 0) - 1) }
              : p,
          ),
      );
    }
  };

  const prLeftForProblem = (problemId: string) => {
    const items = queryClient.getQueryData<SubmissionListItem[]>(
      getContestSubmissionsQueryKey({
        path: { id: contestId },
        headers: authHeaders(),
      }),
    );
    return prCountByProblem(items ?? []).get(problemId) ?? 0;
  };

  const scrollToNext = () => {
    if (!nextSubmissionId) return;
    const el = document.getElementById(submissionPanelId(nextSubmissionId));
    if (el) smoothScrollToBlockStart(el);
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
      await withCommentsScroll(async () => {
        await commentMutation.mutateAsync({
          path: { id: contestId, submissionId },
          body: { text: commentText },
          headers: authHeaders(),
        });
        setComment("");
        await refreshComments();
      });
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) onUnauthorized();
      setActionError(err instanceof ApiError ? err.error : "Не удалось отправить комментарий");
    }
  };

  const submitVerdict = async (verdict: "OK" | "RJ") => {
    setActionError(null);
    try {
      await withCommentsScroll(async () => {
        await verdictMutation.mutateAsync({
          path: { id: contestId, submissionId },
          body: {
            verdict,
            ...(commentText ? { comment: commentText } : {}),
          },
          headers: authHeaders(),
        });
        setComment("");
        patchSubmissionVerdict(verdict);
        await refreshComments();
        if (prLeftForProblem(submission.problem) === 0) {
          setCelebrate(true);
        } else if (nextSubmissionId) {
          scrollToNext();
        }
      });
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) onUnauthorized();
      setActionError(err instanceof ApiError ? err.error : "Не удалось выставить вердикт");
    }
  };

  const onManualNext = () => {
    goNext();
  };

  const commentsList = (
    <ul ref={listContentRef} className="review-comments__list">
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

  const onScrollZoneKeyDown = (e: KeyboardEvent<HTMLElement>) => {
    if (actionPending) return;
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onManualNext();
    }
  };

  const commentsSection =
    thread.length > 0 ? (
      <section className="review-comments" aria-label="Комментарии">
        <h3 className="review-comments__title">
          Комментарии
          <span className="review-comments__count">{thread.length}</span>
        </h3>
        <div
          ref={commentsScrollRef}
          className="review-side-rail__comments-viewport"
        >
          {commentsList}
        </div>
      </section>
    ) : null;

  const onGapAKeyDown = (e: KeyboardEvent<HTMLButtonElement>) => {
    if (!gapsInteractive || !statementAvailable) return;
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      setStatementOpen(true);
    }
  };

  const onGapBKeyDown = (e: KeyboardEvent<HTMLButtonElement>) => {
    if (!gapsInteractive || actionPending) return;
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onManualNext();
    }
  };

  const footNode = nextEnabled ? (
    showAllBesideNext ? (
      <div
        ref={footRef}
        className="review-side-rail__foot review-side-rail__scroll--duo"
        role="button"
        tabIndex={actionPending ? -1 : 0}
        aria-disabled={actionPending}
        aria-label="Следующая задача"
        onClick={() => {
          if (!actionPending) onManualNext();
        }}
        onKeyDown={onScrollZoneKeyDown}
      >
        <div className="review-scroll-duo__bar">
          <button
            type="button"
            className="review-scroll-duo__show-all"
            aria-label="Показать все посылки"
            onClick={(e) => {
              e.stopPropagation();
              onShowAllSubmissions?.();
            }}
          >
            Показать все посылки
            <ChevronDownIcon size={22} />
          </button>
          <span className="review-scroll-duo__next-cluster" aria-hidden>
            <span className="review-scroll-duo__next-label">Следующая задача</span>
            <ChevronRightIcon size={22} />
          </span>
        </div>
      </div>
    ) : (
      <div
        ref={footRef}
        className="review-side-rail__foot"
        role="button"
        tabIndex={actionPending ? -1 : 0}
        aria-disabled={actionPending}
        aria-label={goNextProblem ? "Следующая задача" : "Следующая посылка"}
        onClick={() => {
          if (!actionPending) onManualNext();
        }}
        onKeyDown={onScrollZoneKeyDown}
      >
        <span
          className={
            "review-scroll-hint" + (goNextProblem ? " review-scroll-hint--next-problem" : "")
          }
          aria-hidden
        >
          <span className="review-scroll-hint__label">
            {goNextProblem ? "Следующая задача" : "Следующая посылка"}
          </span>
          {goNextProblem ? <ChevronRightIcon size={28} /> : <ChevronDownIcon size={28} />}
        </span>
      </div>
    )
  ) : null;

  return (
    <article
      id={panelId}
      className={
        "review-detail review-detail--panel" +
        (isActive ? " review-detail--active" : "")
      }
    >
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
          <div className="review-workspace__more code-pane-bar">
            <div className="code-pane-meta">
              <EjudgeContestChip className="contest-chip" title={submissionId}>
                {submissionId}
              </EjudgeContestChip>
            </div>
          </div>
        </div>
        <div className="review-workspace__label-spacer" aria-hidden />

        <div
          ref={codeRef}
          className="review-workspace__code review-code"
          aria-label="Исходный код"
        >
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
          {!commentsQuery.isLoading && sourceText ? (
            <SourceCodeCopyButton
              text={sourceText}
              className={
                findingKey
                  ? "source-code-copy source-code-copy--with-findings"
                  : "source-code-copy"
              }
            />
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

        <aside className="review-side-rail">
          <div
            ref={railRef}
            className={"review-side-rail__rail" + (railReady ? " is-ready" : "")}
          >
            <ProblemStatementScrollZone
              headRef={headRef}
              contestId={contestId}
              contestName={contestName}
              problemLabel={statementProblemLabel}
              statementAvailable={statementAvailable}
              open={statementOpen}
              onOpen={() => setStatementOpen(true)}
              onClose={() => setStatementOpen(false)}
            />
            <div className="review-side-rail__track">
              <aside
                ref={bodyRef}
                className="review-side-panel review-side-rail__body"
                data-scroll-comments={scrollComments ? "true" : "false"}
              >
                <button
                  type="button"
                  className="review-side-rail__gap review-side-rail__gap--a"
                  aria-label="Открыть условие задачи"
                  aria-disabled={!gapsInteractive || !statementAvailable}
                  tabIndex={gapsInteractive && statementAvailable ? 0 : -1}
                  onClick={(e) => {
                    if (!gapsInteractive || e.detail > 1) return;
                    if (statementAvailable) setStatementOpen(true);
                  }}
                  onKeyDown={onGapAKeyDown}
                />
                <div className="review-side-rail__body-core">
                  <div ref={metaRef} className="review-side-head">
                    <span className="review-side-head__name">{submission.participant}</span>
                    <span className={`review-verdict-chip review-verdict-chip--${verdictTone}`}>
                      {formatVerdictLabel(liveVerdict)}
                    </span>
                  </div>
                  <form
                    ref={formRef}
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
                    <div
                      className={`review-verdict__actions${
                        verdictReviewedFromPr ? " review-verdict__actions--reviewed" : ""
                      }`}
                    >
                      <div
                        className={`review-verdict__ok-rj${
                          pendingVerdictActions ? "" : " review-verdict__ok-rj--neutral"
                        }`}
                      >
                        <button
                          type="button"
                          className={`btn btn--icon${
                            pendingVerdictActions ? " btn--ok" : ""
                          }`}
                          disabled={actionPending}
                          aria-label="AC"
                          onClick={() => void submitVerdict("OK")}
                        >
                          AC
                        </button>
                        <button
                          type="button"
                          className={`btn btn--icon${
                            pendingVerdictActions ? " btn--rj" : ""
                          }`}
                          disabled={actionPending}
                          aria-label="RJ"
                          onClick={() => void submitVerdict("RJ")}
                        >
                          RJ
                        </button>
                      </div>
                      <button
                        type="button"
                        className={`btn btn--comment${!prVerdict ? " btn--comment--muted" : ""}`}
                        disabled={actionPending || !commentText}
                        onClick={() => void submitComment()}
                      >
                        Comment
                      </button>
                    </div>
                  </form>
                  {commentsSection}
                </div>
                <button
                  type="button"
                  className="review-side-rail__gap review-side-rail__gap--b"
                  aria-label={
                    goNextProblem ? "Следующая задача" : "Следующая посылка"
                  }
                  aria-disabled={!gapsInteractive || !nextEnabled}
                  tabIndex={gapsInteractive && nextEnabled ? 0 : -1}
                  onClick={(e) => {
                    if (!gapsInteractive || actionPending || e.detail > 1) return;
                    onManualNext();
                  }}
                  onKeyDown={onGapBKeyDown}
                />
              </aside>
            </div>
            {footNode}
          </div>
          {isActive ? (
            <ReviewSideRailDebugHud
              panelLabel={submission.participant}
              snapshot={debugSnapshot}
            />
          ) : null}
        </aside>
      </div>
    </article>
  );
}
