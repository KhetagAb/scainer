type Props = {
  groupBy: "problem" | "participant";
  onChange: (value: "problem" | "participant") => void;
};

function PersonIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <circle cx="10" cy="8" r="5" />
      <path d="M2 21a8 8 0 0 1 10.434-7.62" />
      <circle cx="18" cy="18" r="3" />
      <path d="m22 22-1.9-1.9" />
    </svg>
  );
}

function ProblemIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="m3 16 4 4 4-4" />
      <path d="M7 20V4" />
      <path d="M20 8h-5" />
      <path d="M15 10V6.5a2.5 2.5 0 0 1 5 0V10" />
      <path d="M15 14h5l-5 6h5" />
    </svg>
  );
}

/** Иконки группировки findings: по задачам / по участникам. */
export default function FindingsGroupTabs({ groupBy, onChange }: Props) {
  return (
    <nav className="contest-head__group" role="tablist" aria-label="Группировка">
      <button
        type="button"
        className={`contest-head__group-btn${groupBy === "problem" ? " is-active" : ""}`}
        role="tab"
        aria-selected={groupBy === "problem"}
        aria-label="По задачам"
        title="По задачам"
        onClick={() => onChange("problem")}
      >
        <ProblemIcon />
      </button>
      <button
        type="button"
        className={`contest-head__group-btn${groupBy === "participant" ? " is-active" : ""}`}
        role="tab"
        aria-selected={groupBy === "participant"}
        aria-label="По участникам"
        title="По участникам"
        onClick={() => onChange("participant")}
      >
        <PersonIcon />
      </button>
    </nav>
  );
}
