import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import {
  formatVerdictFilterChipLabel,
  REVIEW_VERDICT_OPTIONS,
  sortVerdicts,
  verdictFilterShortCode,
  type ReviewVerdictFilter,
} from "@/features/review/reviewFilterUtils";
import { formatVerdictLabel, verdictChipTone } from "@/features/review/reviewVerdicts";

type Props = {
  value: ReviewVerdictFilter;
  onChange: (value: ReviewVerdictFilter) => void;
};

type MenuPosition = {
  top: number;
  left: number;
  minWidth: number;
};

export default function ReviewVerdictFilterChip({ value, onChange }: Props) {
  const [open, setOpen] = useState(false);
  const [menuPosition, setMenuPosition] = useState<MenuPosition | null>(null);
  const rootRef = useRef<HTMLDivElement>(null);
  const chipTone =
    value.active && value.verdicts.length === 1
      ? verdictChipTone(value.verdicts[0])
      : null;
  const label = formatVerdictFilterChipLabel(value.verdicts);

  useLayoutEffect(() => {
    if (!open || !rootRef.current) {
      setMenuPosition(null);
      return;
    }

    const updatePosition = () => {
      const rect = rootRef.current?.getBoundingClientRect();
      if (!rect) return;
      setMenuPosition({
        top: rect.bottom + 4,
        left: rect.left,
        minWidth: rect.width,
      });
    };

    updatePosition();
    window.addEventListener("resize", updatePosition);
    window.addEventListener("scroll", updatePosition, true);
    return () => {
      window.removeEventListener("resize", updatePosition);
      window.removeEventListener("scroll", updatePosition, true);
    };
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onDocClick = (e: MouseEvent) => {
      const target = e.target as Node;
      if (rootRef.current?.contains(target)) return;
      if (target instanceof Element && target.closest(".review-filter-chip__menu")) return;
      setOpen(false);
    };
    const id = window.setTimeout(() => {
      document.addEventListener("mousedown", onDocClick);
    }, 0);
    return () => {
      window.clearTimeout(id);
      document.removeEventListener("mousedown", onDocClick);
    };
  }, [open]);

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
    setOpen(false);
  };

  const menu =
    open && menuPosition
      ? createPortal(
          <ul
            className="review-filter-chip__menu review-filter-chip__menu--portal"
            role="listbox"
            aria-label="Вердикт посылки"
            aria-multiselectable="true"
            style={{
              top: menuPosition.top,
              left: menuPosition.left,
              minWidth: menuPosition.minWidth,
            }}
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
                    onClick={(e) =>
                      selectVerdict(v, e.metaKey || e.ctrlKey)
                    }
                  >
                    <span className="review-filter-chip__option-code">
                      {verdictFilterShortCode(v)}
                    </span>
                    <span className="review-filter-chip__option-label">{optionLabel}</span>
                  </button>
                </li>
              );
            })}
          </ul>,
          document.body,
        )
      : null;

  return (
    <>
      <div
        ref={rootRef}
        className={[
          "review-filter-chip",
          chipTone ? `review-filter-chip--${chipTone}` : "",
          value.active ? "review-filter-chip--active" : "review-filter-chip--inactive",
          open ? "review-filter-chip--open" : "",
        ]
          .filter(Boolean)
          .join(" ")}
      >
        <button
          type="button"
          className="review-filter-chip__body"
          aria-pressed={value.active}
          title={value.active ? "Отключить фильтр по вердикту" : "Включить фильтр по вердикту"}
          onClick={toggleActive}
        >
          {label}
        </button>
        <button
          type="button"
          className="review-filter-chip__toggle"
          aria-expanded={open}
          aria-haspopup="listbox"
          aria-label="Выбрать вердикт"
          onMouseDown={(e) => e.stopPropagation()}
          onClick={() => setOpen((v) => !v)}
        >
          <span className="review-filter-chip__chevron" aria-hidden />
        </button>
      </div>
      {menu}
    </>
  );
}
