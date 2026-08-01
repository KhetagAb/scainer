import type { ReviewParticipantFilter } from "@/features/review/reviewTypes";

const INPUT_HINT =
  "Поиск по части логина. Регулярное выражение: оберните в /…/, например /^6B/";

type Props = {
  value: ReviewParticipantFilter;
  onChange: (value: ReviewParticipantFilter) => void;
};

export default function ReviewParticipantFilter({ value, onChange }: Props) {
  const toggleActive = () => onChange({ ...value, active: !value.active });

  return (
    <div
      className={
        "review-filter-chip review-filter-chip--participant-input" +
        (!value.active ? " review-filter-chip--inactive" : "")
      }
    >
      <button
        type="button"
        className="review-filter-chip__body review-filter-chip__body--static"
        aria-pressed={value.active}
        title={value.active ? "Отключить фильтр по участнику" : "Включить фильтр по участнику"}
        onClick={toggleActive}
      >
        Участник
      </button>
      <input
        type="text"
        className="review-filter-chip__inline-input"
        value={value.query}
        onChange={(e) => onChange({ ...value, query: e.target.value })}
        placeholder="например, 6B"
        title={INPUT_HINT}
        spellCheck={false}
        aria-label="Фильтр по участнику"
      />
    </div>
  );
}
