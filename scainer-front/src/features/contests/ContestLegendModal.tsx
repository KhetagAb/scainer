import { useEffect, useId, useRef } from "react";
import { createPortal } from "react-dom";

type Props = {
  open: boolean;
  onClose: () => void;
};

const COLOR_LEVELS: Array<{ level: string; label: string; hint: string }> = [
  { level: "level-0", label: "A", hint: "нет" },
  { level: "level-low", label: "B", hint: "мало" },
  { level: "level-high", label: "C", hint: "много" },
  { level: "level-crit", label: "D", hint: "crit" },
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

  if (!open) return null;

  return createPortal(
    <div
      className="modal-overlay"
      role="presentation"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className="form-card modal-card contest-legend-modal"
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
                Как читать процент подозрительности и цвета задач
              </p>
            </div>
          </div>
        </header>

        <div className="modal-card__scroll contest-legend-modal__body">
          <div className="legend-diagram">
            <div className="legend-diagram__card-wrap">
              <div className="contest-card contest-card--heat-mid legend-diagram__card" aria-hidden>
                <div className="contest-card__top">
                  <span className="contest-card__name">День 01</span>
                  <span className="score-badge contest-card__heat-badge level-mid">12%</span>
                </div>
                <div className="contest-stats contest-stats--meta-only">
                  <div className="contest-stats__left">
                    <span className="contest-stats__meta contest-stats__muted">69 посылок</span>
                  </div>
                  <span className="contest-stats__problems">
                    {COLOR_LEVELS.map((c) => (
                      <span key={c.level} className={`contest-problem-chip ${c.level}`}>
                        {c.label}
                      </span>
                    ))}
                  </span>
                </div>
              </div>
            </div>

            <div className="legend-diagram__notes">
              <aside className="legend-callout legend-callout--pct">
                <p>
                  <strong>%</strong> — взвешенный (<code>100 × Σ score / посылки</code>) процент
                </p>
                <p>подозрительности</p>
              </aside>

              <aside className="legend-callout legend-callout--colors">
                <p>
                  <strong>Цвет</strong> — относительная (<code>count / max</code>) жара
                  сигналов.
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
