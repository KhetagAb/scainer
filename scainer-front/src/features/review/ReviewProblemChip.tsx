import { FileText } from "lucide-react";
import { NavLink } from "react-router-dom";

type Props = {
  to: string;
  label: string;
  pr: number;
  active: boolean;
  statementAvailable: boolean;
  onStatementOpen: (label: string) => void;
};

export default function ReviewProblemChip({
  to,
  label,
  pr,
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
        (pr === 0 ? " is-empty" : "") +
        (statementAvailable ? " is-statement" : "")
      }
      title={active && statementAvailable ? `Условие — ${label}` : label}
      aria-label={active && statementAvailable ? `Условие — ${label}` : label}
      onClick={(e) => {
        if (active && statementAvailable) {
          e.preventDefault();
          onStatementOpen(label);
        }
      }}
    >
      <span className="review-picker__chip-title">{label}</span>
      <span className="review-picker__chip-pr" aria-label={`${pr} pending review`}>
        {pr}
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
              onStatementOpen(label);
            }}
          >
            Условие
          </button>
        </span>
      ) : null}
    </NavLink>
  );
}
