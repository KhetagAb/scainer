import {
  formatJobProgress,
  jobProgressPercent,
  type JobProgress,
} from "@/features/contests/jobs/jobShared";

type Props = {
  progress: JobProgress;
  /** Переопределить подпись (например «3/12» для пачки). */
  label?: string;
  className?: string;
};

/** Высокий прогресс-бар с текстом по центру. */
export default function JobProgressBar({ progress, label, className }: Props) {
  const text = label ?? formatJobProgress(progress);
  const pct = jobProgressPercent(progress);

  return (
    <div
      className={className ? `import-progress-bar ${className}` : "import-progress-bar"}
      role="progressbar"
      aria-label={text}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={pct ?? undefined}
      aria-valuetext={text}
    >
      <span
        className={`import-progress-bar__fill${pct == null ? " is-indeterminate" : ""}`}
        style={pct != null ? { width: `${pct}%` } : undefined}
      />
      <span className="import-progress-bar__label">{text}</span>
    </div>
  );
}
