import type { KeyboardEvent } from "react";
import { formatSubmittedAt } from "@/features/review/reviewFindings";

export type ReviewCommentItem = {
  id: string;
  from: string;
  time: string;
  text: string;
  subject?: string | null;
};

type Props = {
  thread: ReviewCommentItem[];
  commentsOpen: boolean;
  onExpand: () => void;
};

export default function ReviewCommentsThread({ thread, commentsOpen, onExpand }: Props) {
  if (!thread.length) return null;

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

  const onCommentsKeyDown = (e: KeyboardEvent<HTMLElement>) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onExpand();
    }
  };

  const commentsCollapsed = !commentsOpen && thread.length > 1;

  return (
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
          onClick={onExpand}
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
  );
}
