import type { SyncActionModel } from "@/features/contests/sync/contestDataStatus";
import { ContestSyncIcon } from "@/features/contests/sync/ContestSyncIcon";
import {
  formatJobProgress,
  type JobProgress,
} from "@/features/contests/jobs/jobShared";
import JobProgressBar from "@/features/contests/jobs/JobProgressBar";

type Props = {
  model: SyncActionModel;
  progress?: JobProgress;
  batchProgress?: { done: number; total: number } | null;
  disabled?: boolean;
  onClick: () => void;
};

export default function ContestSyncAction({
  model,
  progress,
  batchProgress,
  disabled,
  onClick,
}: Props) {
  const isRunning = model.state === "running";
  const isWarn = model.state === "warn";
  const isDisabled = disabled || model.state === "disabled";

  const ariaLabel = `${model.label}. ${model.hint}`;

  if (isRunning) {
    const barProgress =
      progress ??
      (batchProgress
        ? {
            phase: "importing",
            done: batchProgress.done,
            total: batchProgress.total,
          }
        : null);
    const batchLabel = batchProgress
      ? `${batchProgress.done}/${batchProgress.total}`
      : undefined;
    const text =
      batchLabel ??
      (barProgress ? formatJobProgress(barProgress) : "Обновляем…");

    return (
      <div className="contest-sync-action contest-sync-action--running">
        <JobProgressBar
          progress={barProgress ?? { phase: "importing", done: 0, total: 0 }}
          label={text}
          className="import-progress-bar--sync"
        />
      </div>
    );
  }

  const showLabel = isWarn || model.label !== "Обновить";

  return (
    <button
      type="button"
      className={
        "contest-sync-action" +
        (isWarn ? " contest-sync-action--warn" : " contest-sync-action--idle") +
        (isDisabled ? " contest-sync-action--disabled" : "")
      }
      onClick={onClick}
      disabled={isDisabled}
      title={model.hint}
      aria-label={ariaLabel}
    >
      <span
        className={
          "contest-sync-action__icon" + (isWarn ? " contest-sync-action__icon--warn" : "")
        }
      >
        <ContestSyncIcon size={18} />
      </span>
      {showLabel ? (
        <span className="contest-sync-action__label">{model.label}</span>
      ) : null}
    </button>
  );
}
