type Props = {
  groupBy: "problem" | "participant";
  onChange: (value: "problem" | "participant") => void;
  className?: string;
};

export default function FindingsGroupTabs({ groupBy, onChange, className }: Props) {
  return (
    <div className={className ? `header-control ${className}` : "header-control"}>
      <span className="control-group__label">Группировка</span>
      <nav className="group-tabs" role="tablist" aria-label="Группировка">
        <button
          type="button"
          className="tab"
          role="tab"
          aria-selected={groupBy === "problem"}
          onClick={() => onChange("problem")}
        >
          По задачам
        </button>
        <button
          type="button"
          className="tab"
          role="tab"
          aria-selected={groupBy === "participant"}
          onClick={() => onChange("participant")}
        >
          По участникам
        </button>
      </nav>
    </div>
  );
}
