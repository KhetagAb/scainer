import { useEffect, useId, useMemo, useRef, useState, type CSSProperties } from "react";
import temp50Url from "@/assets/temp-50.png";
import temp75Url from "@/assets/temp-75.png";
import temp90Url from "@/assets/temp-90.png";

type Props = {
  threshold: number;
  onChange: (threshold: number) => void;
};

/** Сверху вниз: Сильная → Средняя → Слабая. */
const LEVELS = [
  { value: 0.9, label: "Сильная", icon: temp90Url },
  { value: 0.75, label: "Средняя", icon: temp75Url },
  { value: 0.5, label: "Слабая", icon: temp50Url },
] as const;

function levelIndex(threshold: number): number {
  let best = 0;
  for (let i = 1; i < LEVELS.length; i++) {
    if (Math.abs(LEVELS[i].value - threshold) < Math.abs(LEVELS[best].value - threshold)) {
      best = i;
    }
  }
  return best;
}

function TempIcon({ src }: { src: string }) {
  return (
    <span
      className="sensitivity-root__icon"
      style={{ maskImage: `url(${src})`, WebkitMaskImage: `url(${src})` }}
      aria-hidden
    />
  );
}

/** Speed dial: уровни выезжают столбиком прямо над триггером. */
export default function SensitivitySlider({ threshold, onChange }: Props) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const closeTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const dialId = useId();
  const index = useMemo(() => levelIndex(threshold), [threshold]);
  const current = LEVELS[index];

  const clearCloseTimer = () => {
    if (closeTimer.current != null) {
      clearTimeout(closeTimer.current);
      closeTimer.current = null;
    }
  };

  const openDial = () => {
    clearCloseTimer();
    setOpen(true);
  };

  const scheduleClose = () => {
    clearCloseTimer();
    closeTimer.current = setTimeout(() => setOpen(false), 120);
  };

  useEffect(() => {
    return () => clearCloseTimer();
  }, []);

  useEffect(() => {
    if (!open) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [open]);

  return (
    <div
      className={`sensitivity-root${open ? " is-open" : ""}`}
      ref={rootRef}
      onMouseEnter={openDial}
      onMouseLeave={scheduleClose}
      onFocus={openDial}
      onBlur={(e) => {
        if (!rootRef.current?.contains(e.relatedTarget as Node)) {
          scheduleClose();
        }
      }}
    >
      {open ? (
        <div
          className="sensitivity-dial"
          id={dialId}
          role="listbox"
          aria-label="Чувствительность"
        >
          <span className="sensitivity-dial__title">Чувствительность</span>
          {LEVELS.map((level, i) => {
            const active = level.value === current.value;
            return (
              <button
                key={level.label}
                type="button"
                role="option"
                aria-selected={active}
                className={`sensitivity-dial__action${active ? " is-active" : ""}`}
                style={{ "--dial-i": String(LEVELS.length - i) } as CSSProperties}
                onClick={() => onChange(level.value)}
              >
                <span className="sensitivity-dial__action-label">{level.label}</span>
                <span className="sensitivity-dial__action-face">
                  <TempIcon src={level.icon} />
                </span>
              </button>
            );
          })}
        </div>
      ) : null}

      <button
        type="button"
        className={`sensitivity-root__btn${open ? " is-open" : ""}`}
        aria-label={`Чувствительность: ${current.label}`}
        aria-expanded={open}
        aria-controls={open ? dialId : undefined}
        title={`Чувствительность · ${current.label}`}
      >
        <TempIcon src={current.icon} />
      </button>
    </div>
  );
}
