import { useEffect, useRef, useState } from "react";
import {
  formatVerdictFilterChipLabel,
  REVIEW_VERDICT_OPTIONS,
  sortVerdicts,
  VERDICT_FILTER_CHIP_MAX_LABEL,
  verdictFilterShortCode,
  type ReviewVerdictFilter,
} from "@/features/review/reviewFilterUtils";
import { formatVerdictLabel, verdictChipTone } from "@/features/review/reviewVerdicts";

type Props = {
  value: ReviewVerdictFilter;
  onChange: (value: ReviewVerdictFilter) => void;
};

export default function ReviewVerdictFilterChip({ value, onChange }: Props) {
  const [touchOpen, setTouchOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const chipTone =
    value.active && value.verdicts.length === 1
      ? verdictChipTone(value.verdicts[0])
      : null;
  const label = formatVerdictFilterChipLabel(value.verdicts);

  useEffect(() => {
    if (!touchOpen) return;
    const onDocClick = (e: MouseEvent) => {
      if (rootRef.current?.contains(e.target as Node)) return;
      setTouchOpen(false);
    };
    const id = window.setTimeout(() => {
      document.addEventListener("mousedown", onDocClick);
    }, 0);
    return () => {
      window.clearTimeout(id);
      document.removeEventListener("mousedown", onDocClick);
    };
  }, [touchOpen]);

  const toggleActive = () => onChange({ ...value, active: !value.active });

  const selectVerdict = (verdict: string, multi: boolean) => {
    if (multi) {
      const selected = new Set(value.verdicts.map((v) => v.toUpperCase()));
      const code = verdict.toUpperCase();
      if (selected.has(code)) selected.delete(code);
      else selected.add(code);

      const next = sortVerdicts(
        REVIEW_VERDICT_OPTIONS.filter((v) => selected.has(v)),
      );
      if (!next.length) {
        onChange({ active: false, verdicts: [verdict] });
        return;
      }
      onChange({ active: true, verdicts: next });
      return;
    }

    onChange({ active: true, verdicts: [verdict] });
    setTouchOpen(false);
  };

  const onPickerClick = () => {
    if (window.matchMedia("(hover: hover)").matches) return;
    setTouchOpen((open) => !open);
  };

  return (
    <div
      ref={rootRef}
      className={[
        "review-filter-chip",
        "review-filter-chip--verdict",
        chipTone ? `review-filter-chip--${chipTone}` : "",
        value.active ? "review-filter-chip--active" : "review-filter-chip--inactive",
        touchOpen ? "review-filter-chip--menu-open" : "",
      ]
        .filter(Boolean)
        .join(" ")}
    >
      <button
        type="button"
        className="review-filter-chip__body review-filter-chip__body--static"
        aria-pressed={value.active}
        title={value.active ? "Отключить фильтр по вердикту" : "Включить фильтр по вердикту"}
        onClick={toggleActive}
      >
        Статус
      </button>
      <div className="review-filter-chip__menu-anchor">
        <span
          className="review-filter-chip__picker"
          role="button"
          tabIndex={0}
          aria-haspopup="listbox"
          aria-expanded={touchOpen}
          aria-label="Выбрать вердикт"
          onClick={onPickerClick}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              onPickerClick();
            }
          }}
        >
          <span className="review-filter-chip__picker-label-stack">
            <span className="review-filter-chip__picker-label-sizer" aria-hidden>
              {VERDICT_FILTER_CHIP_MAX_LABEL}
            </span>
            <span className="review-filter-chip__picker-label">{label}</span>
          </span>
          <span className="review-filter-chip__chevron" aria-hidden />
        </span>
        <ul
          className="review-filter-chip__menu"
          role="listbox"
          aria-label="Вердикт посылки"
          aria-multiselectable="true"
        >
          {REVIEW_VERDICT_OPTIONS.map((v) => {
            const optionTone = verdictChipTone(v);
            const optionLabel = formatVerdictLabel(v);
            const selected = value.verdicts.some(
              (code) => code.trim().toUpperCase() === v,
            );
            return (
              <li key={v} role="presentation">
                <button
                  type="button"
                  role="option"
                  aria-selected={selected}
                  className={[
                    "review-filter-chip__option",
                    `review-filter-chip__option--${optionTone}`,
                    selected ? "review-filter-chip__option--selected" : "",
                  ]
                    .filter(Boolean)
                    .join(" ")}
                  onClick={(e) => selectVerdict(v, e.metaKey || e.ctrlKey)}
                >
                  <span className="review-filter-chip__option-code">
                    {verdictFilterShortCode(v)}
                  </span>
                  <span className="review-filter-chip__option-label">{optionLabel}</span>
                </button>
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}
