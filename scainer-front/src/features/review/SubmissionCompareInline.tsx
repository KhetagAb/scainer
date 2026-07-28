import { FileDiff, X } from "lucide-react";
import { useEffect, useRef, useState, type KeyboardEvent } from "react";

type Props = {
  currentRunId: string | null;
  targetRunId: string;
  defaultTargetRunId: string | null;
  onTargetRunIdChange: (value: string) => void;
  open: boolean;
  manualEntry: boolean;
  onOpenDefault: () => void;
  onStartManual: () => void;
  onClose: () => void;
  loading?: boolean;
  hasDefaultTarget?: boolean;
  langWarning?: string | null;
};

export default function SubmissionCompareInline({
  currentRunId,
  targetRunId,
  defaultTargetRunId,
  onTargetRunIdChange,
  open,
  manualEntry,
  onOpenDefault,
  onStartManual,
  onClose,
  loading = false,
  hasDefaultTarget = false,
  langWarning,
}: Props) {
  const [draftRunId, setDraftRunId] = useState(targetRunId);
  const runIdRef = useRef<HTMLInputElement>(null);

  const showIcon = !open;
  const iconViolet = hasDefaultTarget && !open;
  const showDefaultChip =
    hasDefaultTarget && open && !manualEntry && Boolean(defaultTargetRunId);
  const showManualChip = open && (!hasDefaultTarget || manualEntry);

  useEffect(() => {
    if (showManualChip) setDraftRunId(targetRunId);
  }, [showManualChip, targetRunId]);

  useEffect(() => {
    if (!showManualChip || targetRunId || loading) return;
    const id = window.requestAnimationFrame(() => {
      runIdRef.current?.focus();
    });
    return () => window.cancelAnimationFrame(id);
  }, [showManualChip, targetRunId, loading]);

  useEffect(() => {
    if (!showManualChip) return;
    const t = window.setTimeout(() => {
      if (draftRunId !== targetRunId) onTargetRunIdChange(draftRunId);
    }, 300);
    return () => window.clearTimeout(t);
  }, [draftRunId, targetRunId, onTargetRunIdChange, showManualChip]);

  const onRunIdKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      onTargetRunIdChange(draftRunId);
    }
  };

  const onIconClick = () => {
    if (hasDefaultTarget) onOpenDefault();
    else onStartManual();
  };

  return (
    <div
      className={
        "submission-compare-inline" +
        (open ? " submission-compare-inline--open" : "") +
        (loading ? " submission-compare-inline--loading" : "")
      }
    >
      {showIcon ? (
        <button
          type="button"
          className={
            "submission-compare-inline__toggle" +
            (loading ? " submission-compare-inline__toggle--loading" : "") +
            (iconViolet ? " submission-compare-inline__toggle--has-default" : "")
          }
          onClick={onIconClick}
          disabled={loading}
          title={
            loading
              ? "Загрузка посылки…"
              : hasDefaultTarget
                ? "Сравнить с предыдущей посылкой"
                : "Сравнить с другой посылкой"
          }
          aria-label={
            loading
              ? "Загрузка посылки"
              : hasDefaultTarget
                ? "Сравнить с предыдущей посылкой"
                : "Сравнить с другой посылкой"
          }
          aria-expanded={open}
          aria-busy={loading}
        >
          <FileDiff size={14} aria-hidden />
        </button>
      ) : null}

      {showDefaultChip ? (
        <div className="submission-compare-inline__chip">
          <span className="submission-compare-inline__label" aria-hidden>
            {currentRunId ?? "?"} ↔
          </span>
          <button
            type="button"
            className="submission-compare-inline__default-link"
            onClick={onStartManual}
            disabled={loading}
            title="Указать другой run_id"
          >
            {defaultTargetRunId}
          </button>
          {langWarning ? (
            <span className="submission-compare-inline__warn" title={langWarning}>
              lang
            </span>
          ) : null}
          <button
            type="button"
            className="submission-compare-inline__close"
            onClick={onClose}
            aria-label="Закрыть сравнение"
          >
            <X size={14} aria-hidden />
          </button>
        </div>
      ) : null}

      {showManualChip ? (
        <div className="submission-compare-inline__chip">
          <span className="submission-compare-inline__label" aria-hidden>
            {currentRunId ?? "?"} ↔
          </span>
          <input
            ref={runIdRef}
            type="text"
            inputMode="numeric"
            className="submission-compare-inline__run-id"
            value={draftRunId}
            onChange={(e) => setDraftRunId(e.target.value)}
            onKeyDown={onRunIdKeyDown}
            disabled={loading}
            size={3}
            maxLength={6}
            aria-label="Run ID для сравнения"
          />
          {langWarning ? (
            <span className="submission-compare-inline__warn" title={langWarning}>
              lang
            </span>
          ) : null}
          <button
            type="button"
            className="submission-compare-inline__close"
            onClick={onClose}
            aria-label="Закрыть сравнение"
          >
            <X size={14} aria-hidden />
          </button>
        </div>
      ) : null}
    </div>
  );
}
