import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import ContestCard from "@/features/contests/cards/ContestCard";
import {
  ContestFindingsIcon,
  ContestReviewIcon,
} from "@/features/contests/ui/ContestSectionNav";

const SEEN_KEY = "scainer.contestLegend.seen";
const DEMO_CONTEST_ID = "50505";

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

export default function ContestLegendModal({ open, onClose }: Props) {
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
          "contest-legend-panel contest-legend-panel--contest" +
          (visible ? " contest-legend-panel--visible" : "")
        }
        role="dialog"
        aria-modal="true"
        aria-labelledby="contest-legend-title"
      >
        <div className="contest-legend-panel__body">
          <h2 id="contest-legend-title" className="contest-legend-panel__title" lang="ru">
            Как читать карточку контеста?
          </h2>
          <div className="legend-diagram">
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
                        <strong>Ревью</strong> — очередь посылок на ручную проверку.
                      </p>
                      <ul className="legend-callout__swatches">
                        <li>
                          <span className="contest-card__nav-key contest-card__nav-key--violet">
                            <span className="contest-card__nav-key__dot" />
                            67 PR
                          </span>
                          <span>число ожидающих к проверке посылок</span>
                        </li>
                      </ul>
                    </div>
                  </li>
                </ul>
              </aside>
            </div>

            <div className="legend-diagram__card-wrap">
              <ContestCard
                className="legend-diagram__card"
                preview
                id={DEMO_CONTEST_ID}
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

            <div className="legend-diagram__notes legend-diagram__notes--right">
              <aside className="legend-callout legend-callout--sections">
                <ul className="legend-callout__sections">
                  <li>
                    <span
                      className="legend-callout__section-icon legend-callout__section-icon--findings"
                      aria-hidden
                    >
                      <ContestFindingsIcon size={16} />
                    </span>
                    <p lang="ru">
                      <strong>Детект</strong> — список подозрительных посылок и обнаруженных нарушений.
                    </p>
                  </li>
                  <li>
                    <span
                      className="legend-callout__section-icon legend-callout__section-icon--spacer"
                      aria-hidden
                    />
                    <div className="legend-callout__section-body">
                      <p lang="ru">
                        <strong>Цвет</strong> — чем больше сигналов по задаче — тем
                        вероятнее ложноположительное срабатывание.
                      </p>
                      <ul className="legend-callout__swatches legend-callout__swatches--colors">
                        {COLOR_LEVELS.map((c) => (
                          <li key={c.level}>
                            <span className="contest-stats__problems">
                              <span className={`contest-problem-chip ${c.level}`}>
                                {c.label}
                              </span>
                            </span>
                            <span>{c.hint}</span>
                          </li>
                        ))}
                      </ul>
                    </div>
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
