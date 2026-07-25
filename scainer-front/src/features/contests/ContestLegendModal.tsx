import { useEffect, useId, useRef, useState } from "react";
import { createPortal } from "react-dom";
import ContestCard from "@/features/contests/ContestCard";
import {
  ContestFindingsIcon,
  ContestReviewIcon,
} from "@/features/contests/ContestSectionNav";

const SEEN_KEY = "scainer.contestLegend.seen";

export function wasContestLegendSeen(): boolean {
  try {
    return localStorage.getItem(SEEN_KEY) === "1";
  } catch {
    return true;
  }
}

export function markContestLegendSeen(): void {
  try {
    localStorage.setItem(SEEN_KEY, "1");
  } catch {
    /* ignore quota / private mode */
  }
}

type Props = {
  open: boolean;
  onClose: () => void;
};

const COLOR_LEVELS: Array<{ level: string; label: string; hint: string }> = [
  { level: "level-0", label: "A", hint: "нет" },
  { level: "level-low", label: "B", hint: ">10%" },
  { level: "level-high", label: "C", hint: "≤10%" },
];

function HeaderIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" aria-hidden>
      <rect x="3" y="4" width="18" height="16" rx="2.5" stroke="currentColor" strokeWidth="2" />
      <path d="M7 9h6M7 13h10M7 17h4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  );
}

export default function ContestLegendModal({ open, onClose }: Props) {
  const titleId = useId();
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
        "modal-overlay modal-overlay--legend" +
        (visible ? " modal-overlay--legend-visible" : "")
      }
      role="presentation"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className={
          "form-card modal-card contest-legend-modal" +
          (visible ? " contest-legend-modal--visible" : "")
        }
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
      >
        <header className="contest-legend-modal__header">
          <div className="contest-legend-modal__header-main">
            <span className="contest-legend-modal__icon" aria-hidden>
              <HeaderIcon />
            </span>
            <div>
              <h2 id={titleId} className="contest-legend-modal__title">
                Карточка контеста
              </h2>
              <p className="contest-legend-modal__subtitle">
                Разделы и цвета задач
              </p>
            </div>
          </div>
        </header>

        <div className="modal-card__scroll contest-legend-modal__body">
          <div className="legend-diagram">
            <div className="legend-diagram__card-wrap">
              <ContestCard
                className="legend-diagram__card"
                aria-hidden
                staticActions
                id="legend"
                name="День 01"
                displayName="День 01"
                pendingCount={67}
                hasStrongSignals
                stats={{ submissionCount: 69 }}
                problemStats={[
                  { id: "a", name: "A", suspiciousSharePercent: 0, pendingCount: 0 },
                  { id: "b", name: "B", suspiciousSharePercent: 20, pendingCount: 0 },
                  { id: "c", name: "C", suspiciousSharePercent: 10, pendingCount: 0 },
                ]}
              />
            </div>

            <div className="legend-diagram__notes">
              <aside className="legend-callout legend-callout--sections">
                <ul className="legend-callout__sections">
                  <li>
                    <span className="legend-callout__section-icon legend-callout__section-icon--review" aria-hidden>
                      <ContestReviewIcon size={16} />
                    </span>
                    <p lang="ru">
                      <strong>Ревью</strong> — очередь посылок на ручную проверку.
                      {" "}
                      <span className="contest-card__nav-key contest-card__nav-key--violet">
                        <span className="contest-card__nav-key__dot" />
                        67 PR
                      </span>
                      {" "}— число ожидающих.
                    </p>
                  </li>
                  <li>
                    <span className="legend-callout__section-icon legend-callout__section-icon--findings" aria-hidden>
                      <ContestFindingsIcon size={16} />
                    </span>
                    <p lang="ru">
                      <strong>Детект</strong> — автоматические находки похожести и сигналов.
                    </p>
                  </li>
                </ul>
              </aside>
              <aside className="legend-callout legend-callout--colors">
                <p lang="ru">
                  <strong>Цвет</strong> — доля сигналов задачи: чем больше сигналов по задаче — тем вероятнее ложноположительные срабатывания.
                </p>
                <ul className="legend-callout__swatches">
                  {COLOR_LEVELS.map((c) => (
                    <li key={c.level}>
                      <span className={`contest-problem-chip ${c.level}`}>{c.label}</span>
                      <span>{c.hint}</span>
                    </li>
                  ))}
                </ul>
              </aside>
            </div>
          </div>
        </div>

        <footer className="contest-legend-modal__footer">
          <button
            ref={closeBtnRef}
            type="button"
            className="contest-legend-modal__confirm"
            onClick={onClose}
          >
            Понятно
          </button>
        </footer>
      </div>
    </div>,
    document.body,
  );
}
