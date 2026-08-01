import { useQueryClient } from "@tanstack/react-query";
import { X } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { fetchContestStatement } from "@/features/statements/fetchContestStatement";
import { prefetchProblemStatementExplain } from "@/features/statements/explainQuery";
import ProblemStatementExplainPanel from "@/features/statements/ProblemStatementExplainPanel";
import ProblemStatementViewer from "@/features/statements/ProblemStatementViewer";
import StatementExplainButton from "@/features/statements/StatementExplainButton";
import { useProblemStatementExplain } from "@/features/statements/useProblemStatementExplain";

type Props = {
  open: boolean;
  contestId: string;
  title?: string;
  problemId?: string | null;
  problemLabel?: string | null;
  onClose: () => void;
};

export default function ProblemStatementModal({
  open,
  contestId,
  title,
  problemId,
  problemLabel,
  onClose,
}: Props) {
  const queryClient = useQueryClient();
  const closeBtnRef = useRef<HTMLButtonElement>(null);
  const [mounted, setMounted] = useState(open);
  const [visible, setVisible] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pdfData, setPdfData] = useState<ArrayBuffer | null>(null);

  const {
    explainAvailable,
    explainOpen,
    formalizationView,
    toggleExplain,
    closeOnDelete,
    setExplainData,
  } = useProblemStatementExplain({ open, contestId, problemId });

  useEffect(() => {
    if (open) {
      setMounted(true);
      const id = requestAnimationFrame(() => {
        requestAnimationFrame(() => setVisible(true));
      });
      return () => cancelAnimationFrame(id);
    }
    setVisible(false);
    const t = window.setTimeout(() => setMounted(false), 220);
    return () => window.clearTimeout(t);
  }, [open]);

  useEffect(() => {
    if (!open) {
      setPdfData(null);
      setError(null);
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError(null);
    setPdfData(null);

    void fetchContestStatement(contestId)
      .then((data) => {
        if (!cancelled) {
          setPdfData(data);
          setLoading(false);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Не удалось загрузить условие");
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [open, contestId]);

  useEffect(() => {
    const problemKey = problemId?.trim();
    if (!open || !problemKey) return;
    prefetchProblemStatementExplain(queryClient, contestId, problemKey);
  }, [open, contestId, problemId, queryClient]);

  useEffect(() => {
    if (!mounted) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = prev;
    };
  }, [mounted]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose]);

  useEffect(() => {
    if (open && visible) closeBtnRef.current?.focus();
  }, [open, visible]);

  const onBackdropClick = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      if (e.target === e.currentTarget) onClose();
    },
    [onClose],
  );

  if (!mounted) return null;

  const ariaLabel = title ? `Условие — ${title}` : "Условие";
  const showLayout = !error && (loading || pdfData || explainOpen);

  return createPortal(
    <div
      className={
        "problem-statement-overlay" + (visible ? " problem-statement-overlay--visible" : "")
      }
      role="presentation"
      onClick={onBackdropClick}
    >
      {explainAvailable ? (
        <StatementExplainButton active={explainOpen} onClick={toggleExplain} />
      ) : null}
      <button
        ref={closeBtnRef}
        type="button"
        className="problem-statement-overlay__close"
        aria-label="Закрыть"
        onClick={onClose}
      >
        <X size={22} strokeWidth={2} aria-hidden />
      </button>
      <div
        className={
          "problem-statement-panel" +
          (visible ? " problem-statement-panel--visible" : "") +
          (explainOpen ? " problem-statement-panel--split" : "")
        }
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel}
      >
        {error ? <p className="problem-statement-panel__error">{error}</p> : null}
        {showLayout ? (
          <div className="problem-statement-panel__layout">
            <div className="problem-statement-panel__main">
              {pdfData ? (
                <ProblemStatementViewer data={pdfData} problemLabel={problemLabel} />
              ) : (
                <p className="problem-statement-viewer__loading">Загрузка условия…</p>
              )}
            </div>
            {explainOpen && problemId?.trim() ? (
              <ProblemStatementExplainPanel
                contestId={contestId}
                problemId={problemId}
                view={formalizationView}
                onUpdated={setExplainData}
                onDeleted={closeOnDelete}
              />
            ) : null}
          </div>
        ) : null}
      </div>
    </div>,
    document.body,
  );
}
