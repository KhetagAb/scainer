import { useMemo } from "react";
import { SENSITIVITY_MARKS } from "@/features/contests/shared/useUserSensitivity";

type Props = {
  threshold: number;
  onChange: (threshold: number) => void;
};

/** Слева направо: 50% → 75% → 90% (порог score). */
const LEVELS: ReadonlyArray<{
  value: number;
  label: string;
  hint: string;
}> = [
  {
    value: 0.5,
    label: SENSITIVITY_MARKS[50],
    hint: "Все подозрения. Часть срабатываний будет ложной",
  },
  {
    value: 0.75,
    label: SENSITIVITY_MARKS[75],
    hint: "Баланс находок и точности",
  },
  {
    value: 0.9,
    label: SENSITIVITY_MARKS[90],
    hint: "Только явные случаи. Меньше ложных срабатываний",
  },
];

function levelIndex(threshold: number): number {
  let best = 0;
  for (let i = 1; i < LEVELS.length; i++) {
    if (Math.abs(LEVELS[i].value - threshold) < Math.abs(LEVELS[best].value - threshold)) {
      best = i;
    }
  }
  return best;
}

export default function SensitivitySlider({ threshold, onChange }: Props) {
  const index = useMemo(() => levelIndex(threshold), [threshold]);
  const current = LEVELS[index];

  return (
    <nav
      className="chrome-segment-group chrome-segment-group--sensitivity contest-head__group contest-head__group--sensitivity"
      role="radiogroup"
      aria-label="Чувствительность"
    >
      {LEVELS.map((level, tipIndex) => {
        const active = level.value === current.value;
        return (
          <button
            key={level.value}
            type="button"
            role="radio"
            aria-checked={active}
            className={`chrome-segment-group-btn sensitivity-btn chrome-segment-btn--tip contest-head__group-btn${active ? " is-active" : ""}`}
            aria-label={level.hint}
            onClick={() => onChange(level.value)}
          >
            <span className={`chrome-segment-tip sensitivity-tip sensitivity-tip--${tipIndex}`} role="tooltip">
              {level.hint}
            </span>
            <span className="sensitivity-label" aria-hidden>
              {level.label}
            </span>
          </button>
        );
      })}
    </nav>
  );
}
