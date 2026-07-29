import type { RefObject } from "react";
import { formatVerdictLabel } from "@/features/review/reviewVerdicts";

type Props = {
  participant: string;
  liveVerdict: string;
  verdictTone: string;
  commentRef: RefObject<HTMLTextAreaElement | null>;
  commentFieldId: string;
  comment: string;
  onCommentChange: (value: string) => void;
  actionPending: boolean;
  actionError: string | null;
  pendingVerdictActions: boolean;
  verdictReviewedFromPr: boolean;
  prVerdict: boolean;
  commentText: string;
  onSubmitVerdict: (verdict: "OK" | "RJ") => void;
  onSubmitComment: () => void;
};

export default function ReviewVerdictForm({
  participant,
  liveVerdict,
  verdictTone,
  commentRef,
  commentFieldId,
  comment,
  onCommentChange,
  actionPending,
  actionError,
  pendingVerdictActions,
  verdictReviewedFromPr,
  prVerdict,
  commentText,
  onSubmitVerdict,
  onSubmitComment,
}: Props) {
  return (
    <>
      <div className="review-side-head">
        <span className="review-side-head__name">{participant}</span>
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
          onChange={(e) => onCommentChange(e.target.value)}
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
              className={`btn btn--icon${pendingVerdictActions ? " btn--ok" : ""}`}
              disabled={actionPending}
              aria-label="OK"
              onClick={() => onSubmitVerdict("OK")}
            >
              OK
            </button>
            <button
              type="button"
              className={`btn btn--icon${pendingVerdictActions ? " btn--rj" : ""}`}
              disabled={actionPending}
              aria-label="RJ"
              onClick={() => onSubmitVerdict("RJ")}
            >
              RJ
            </button>
          </div>
          <button
            type="button"
            className={`btn btn--comment${!prVerdict ? " btn--comment--muted" : ""}`}
            disabled={actionPending || !commentText}
            onClick={() => onSubmitComment()}
          >
            Comment
          </button>
        </div>
      </form>
    </>
  );
}
