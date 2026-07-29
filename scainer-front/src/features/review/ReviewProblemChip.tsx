import { FileText } from "lucide-react";
import { NavLink } from "react-router-dom";

type Props = {
  to: string;
  label: string;
  /** Буква задачи из ejudge (ProblemInfo.name) для фильтрации страниц PDF. */
  statementLabel: string;
  count: number;
  active: boolean;
  statementAvailable: boolean;
  onStatementOpen: (statementLabel: string) => void;
};

export default function ReviewProblemChip({
  to,
  label,
  statementLabel,
  count,
  active,
  statementAvailable,
  onStatementOpen,
}: Props) {
  return (
    <NavLink
      to={to}
      className={
        "chip problem-chip review-picker__chip" +
        (active ? " is-active" : "") +
        (count === 0 ? " is-empty" : "") +
        (statementAvailable ? " is-statement" : "")
      }
      title={active && statementAvailable ? `Условие — ${label}` : label}
      aria-label={active && statementAvailable ? `Условие — ${label}` : label}
      onClick={(e) => {
        if (active && statementAvailable) {
          e.preventDefault();
          onStatementOpen(statementLabel);
        }
      }}
    >
      <span className="review-picker__chip-title">{label}</span>
      <span className="review-picker__chip-count" aria-label={`${count} посылок по фильтру`}>
        {count}
      </span>
      {statementAvailable ? (
        <span className="review-picker__chip-statement">
          <FileText size={12} strokeWidth={2} aria-hidden />
          <button
            type="button"
            className="review-picker__chip-statement-label"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              onStatementOpen(statementLabel);
            }}
          >
            Условие
          </button>
        </span>
      ) : null}
    </NavLink>
  );
}
