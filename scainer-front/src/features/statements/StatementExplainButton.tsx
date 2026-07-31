import { Sparkles } from "lucide-react";

type Props = {
  active: boolean;
  onClick: () => void;
};

export default function StatementExplainButton({ active, onClick }: Props) {
  return (
    <button
      type="button"
      className={"statement-explain-btn" + (active ? " statement-explain-btn--active" : "")}
      aria-label="Разобрать условие"
      aria-pressed={active}
      onClick={onClick}
    >
      <Sparkles className="statement-explain-btn__icon" size={20} strokeWidth={2} aria-hidden />
    </button>
  );
}
