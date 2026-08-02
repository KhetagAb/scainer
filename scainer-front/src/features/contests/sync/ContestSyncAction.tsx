import {
  RESYNC_CONFIRM,
  SYNC_WARN_LOAD_LABEL,
  syncWarnReasonLabel,
  type SyncActionModel,
} from "@/features/contests/sync/contestDataStatus";
import { ContestSyncIcon } from "@/features/contests/sync/ContestSyncIcon";
import {
  formatJobProgress,
  SYNC_PROGRESS_SIZER_LABEL,
  type JobProgress,
} from "@/features/contests/jobs/jobShared";
import JobProgressBar from "@/features/contests/jobs/JobProgressBar";

type Props = {
  model: SyncActionModel;
  progress?: JobProgress;
  batchProgress?: { done: number; total: number } | null;
  disabled?: boolean;
  fillWidth?: boolean;
  inTopbar?: boolean;
  onClick: () => void;
  onResync?: () => void;
};

function isResyncClick(e: React.MouseEvent): boolean {
  return e.metaKey || e.ctrlKey;
}

function SyncActionSizer() {
  return (
    <div className="contest-sync-action-slot__sizer" aria-hidden="true">
      <button
        type="button"
        className="btn btn--primary contest-sync-action contest-sync-action--labeled"
        tabIndex={-1}
        disabled
      >
        <ContestSyncIcon size={14} className="btn--icon__glyph contest-sync-action__glyph" />
        <span className="contest-sync-action__label">{SYNC_WARN_LOAD_LABEL}</span>
      </button>
      <div className="import-progress-bar import-progress-bar--sync">
        <span className="import-progress-bar__label">{SYNC_PROGRESS_SIZER_LABEL}</span>
      </div>
    </div>
  );
}

export default function ContestSyncAction({
  model,
  progress,
  batchProgress,
  disabled,
  fillWidth = false,
  inTopbar = false,
  onClick,
  onResync,
}: Props) {
  const isRunning = model.state === "running";
  const isWarn = model.state === "warn";
  const isDisabled = disabled || model.state === "disabled";

  const resyncShortcutHint = onResync ? "⌘+клик — полный пересбор." : null;
  const warnChipLabel =
    isWarn && model.warnReason ? syncWarnReasonLabel(model.warnReason) : null;
  const tipLines = [model.warnDetail, model.hint, resyncShortcutHint].filter(
    (line): line is string => Boolean(line),
  );
  const tipText = tipLines.join("\n");
  const ariaLabel = warnChipLabel
    ? `${warnChipLabel}. ${tipLines.join(" ")}`
    : `${model.label}. ${tipLines.join(" ")}`;

  const handleClick = (e: React.MouseEvent<HTMLButtonElement>) => {
    if (onResync && isResyncClick(e)) {
      if (window.confirm(RESYNC_CONFIRM)) onResync();
      return;
    }
    onClick();
  };

  const actionIcon = (size: number | undefined, className: string) => (
    <ContestSyncIcon size={size} className={className} />
  );

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
      <div
        className={
          "contest-sync-action-slot" +
          (fillWidth || inTopbar ? " contest-sync-action-slot--fill-width" : "")
        }
      >
        {!fillWidth && !inTopbar ? <SyncActionSizer /> : null}
        <div className="contest-sync-action contest-sync-action--running">
          <JobProgressBar
            progress={barProgress ?? { phase: "importing", done: 0, total: 0 }}
            label={text}
            className="import-progress-bar--sync"
          />
        </div>
      </div>
    );
  }

  const isIconOnly =
    inTopbar ||
    (model.label === "Обновить" &&
      (model.state === "idle" || model.state === "disabled"));

  const useTopbarChrome = inTopbar && isIconOnly;
  const showTopbarWarnChip = useTopbarChrome && warnChipLabel != null;

  const topbarBtnClass =
    "app-topbar-trailing-btn app-topbar-trailing-btn--primary contest-sync-action" +
    (showTopbarWarnChip ? " app-topbar-trailing-btn--labeled chrome-segment-btn--tip" : "") +
    (isWarn ? " app-topbar-trailing-btn--warn" : "") +
    (isDisabled ? " contest-sync-action--disabled" : "");

  return (
    <button
      type="button"
      className={
        (useTopbarChrome
          ? topbarBtnClass
          : "btn btn--primary contest-sync-action") +
        (isIconOnly && !useTopbarChrome ? " contest-sync-action--icon-only" : "") +
        (!isIconOnly ? " contest-sync-action--labeled" : "") +
        (isWarn && !useTopbarChrome ? " contest-sync-action--warn" : "") +
        (isDisabled ? " contest-sync-action--disabled" : "") +
        (fillWidth && !isIconOnly ? " contest-sync-action--fill-width" : "")
      }
      onClick={handleClick}
      disabled={isDisabled}
      title={useTopbarChrome && !showTopbarWarnChip ? model.hint : undefined}
      aria-label={ariaLabel}
    >
      {showTopbarWarnChip ? (
        <span className="chrome-segment-tip app-topbar-trailing-btn__tip" role="tooltip">
          {tipText}
        </span>
      ) : null}
      {actionIcon(
        isIconOnly ? undefined : 14,
        useTopbarChrome
          ? "app-topbar-trailing-btn__glyph"
          : "btn--icon__glyph contest-sync-action__glyph",
      )}
      {showTopbarWarnChip ? (
        <span className="app-topbar-trailing-btn__label">{warnChipLabel}</span>
      ) : null}
      {!isIconOnly ? (
        <span className="contest-sync-action__label">{model.label}</span>
      ) : null}
    </button>
  );
}
