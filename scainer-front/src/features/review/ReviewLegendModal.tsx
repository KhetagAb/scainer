import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { ContestReviewIcon } from "@/features/contests/ContestSectionNav";
import ReviewLegendPreview from "@/features/review/ReviewLegendPreview";
import {
  REVIEW_LEGEND_DEMO_PROBLEMS,
  reviewLegendProblemLabel,
} from "@/features/review/reviewLegendDemo";

const SEEN_KEY = "scainer.reviewLegend.seen";

export function wasReviewLegendSeen(): boolean {
  try {
    return localStorage.getItem(SEEN_KEY) === "1";
  } catch {
    return true;
  }
}

export function markReviewLegendSeen(): void {
  try {
    localStorage.setItem(SEEN_KEY, "1");
  } catch {
    /* ignore quota / private mode */
  }
}

function CodeCiteIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={16}
      height={16}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
      <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
    </svg>
  );
}

function NavDownIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={16}
      height={16}
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

type Props = {
  open: boolean;
  onClose: () => void;
};

export default function ReviewLegendModal({ open, onClose }: Props) {
  const closeBtnRef = useRef<HTMLButtonElement>(null);
  const [mounted, setMounted] = useState(open);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (open) {
      setMounted(true);
      const id = requestAnimationFrame(() => {
        requestAnimationFrame(() => setVisible(true));
      });
      return () => cancelAnimationFrame(id);
    }
    setVisible(false);
    const t = window.setTimeout(() => setMounted(false), 220);
    return () => window.clearTimeout(t);
  }, [open]);

  useEffect(() => {
    if (!open) return;
    closeBtnRef.current?.focus();
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      window.removeEventListener("keydown", onKey);
      document.body.style.overflow = prevOverflow;
    };
  }, [open, onClose]);

  if (!mounted) return null;

  return createPortal(
    <div
      className={
        "contest-legend-overlay" +
        (visible ? " contest-legend-overlay--visible" : "")
      }
      role="presentation"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className={
          "contest-legend-panel contest-legend-panel--review" +
          (visible ? " contest-legend-panel--visible" : "")
        }
        role="dialog"
        aria-modal="true"
        aria-labelledby="review-legend-title"
      >
        <div className="contest-legend-panel__body">
          <h2 id="review-legend-title" className="contest-legend-panel__title" lang="ru">
            Как проводить code review посылок?
          </h2>
          <div className="legend-diagram legend-diagram--review">
            <div className="legend-diagram__notes legend-diagram__notes--left">
              <aside className="legend-callout legend-callout--sections">
                <ul className="legend-callout__sections">
                  <li>
                    <span
                      className="legend-callout__section-icon legend-callout__section-icon--review"
                      aria-hidden
                    >
                      <ContestReviewIcon size={16} />
                    </span>
                    <div className="legend-callout__section-body">
                      <p lang="ru">
                        <strong>Задачи</strong> — переключение между задачами контеста.
                      </p>
                      <ul className="legend-callout__swatches">
                        <li>
                          <span className="chip problem-chip review-picker__chip is-active">
                            {reviewLegendProblemLabel(
                              REVIEW_LEGEND_DEMO_PROBLEMS[0].id,
                              REVIEW_LEGEND_DEMO_PROBLEMS[0].name,
                            )}
                            <span className="review-picker__pr">
                              {REVIEW_LEGEND_DEMO_PROBLEMS[0].pr}
                            </span>
                          </span>
                          <span>с числом посылок Pending Review</span>
                        </li>
                      </ul>
                    </div>
                  </li>
                  <li>
                    <span className="legend-callout__section-icon" aria-hidden>
                      <CodeCiteIcon />
                    </span>
                    <div className="legend-callout__section-body">
                      <p lang="ru">
                        <strong>Цитаты</strong> — привязка комментария к строкам кода.
                      </p>
                      <p className="legend-callout__detail" lang="ru">
                        Двойной клик по строке или выделение в коде добавляет {" "}
                        <code>Строка X:</code>.
                      </p>
                      <p className="legend-callout__detail" lang="ru">
                        Попробуйте зажать Cmd/Ctrl при выделением.
                      </p>
                    </div>
                  </li>
                </ul>
              </aside>
            </div>

            <div className="legend-diagram__card-wrap legend-diagram__card-wrap--review">
              <ReviewLegendPreview className="legend-diagram__card" preview />
            </div>

            <div className="legend-diagram__notes legend-diagram__notes--right">
              <aside className="legend-callout legend-callout--sections">
                <ul className="legend-callout__sections">
                  <li>
                    <span className="legend-callout__section-icon" aria-hidden>
                      <input
                        type="checkbox"
                        className="review-picker__filter-input review-picker__filter-input--legend"
                        checked
                        readOnly
                        tabIndex={-1}
                      />
                    </span>
                    <p lang="ru">
                      <strong>PR only</strong> — показать только посылки на проверку.
                    </p>
                  </li>
                  <li>
                    <span className="legend-callout__section-icon" aria-hidden>
                      <NavDownIcon />
                    </span>
                    <p lang="ru">
                      <strong>Навигация</strong> — клик под формой: следующая посылка или
                      задача.
                    </p>
                  </li>
                </ul>
              </aside>
            </div>
          </div>
        </div>

        <button
          ref={closeBtnRef}
          type="button"
          className="contest-legend-panel__confirm"
          onClick={onClose}
        >
          Понятно
        </button>
      </div>
    </div>,
    document.body,
  );
}
