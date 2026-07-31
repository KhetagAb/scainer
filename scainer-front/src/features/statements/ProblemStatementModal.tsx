import { X } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import type { ProblemStatementExplainView } from "@/client/types.gen";
import { fetchContestStatement } from "@/features/statements/fetchContestStatement";
import { fetchProblemStatementExplain } from "@/features/statements/fetchProblemStatementExplain";
import ProblemStatementExplainPanel from "@/features/statements/ProblemStatementExplainPanel";
import ProblemStatementViewer from "@/features/statements/ProblemStatementViewer";
import StatementExplainButton from "@/features/statements/StatementExplainButton";

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
  const closeBtnRef = useRef<HTMLButtonElement>(null);
  const [mounted, setMounted] = useState(open);
  const [visible, setVisible] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pdfData, setPdfData] = useState<ArrayBuffer | null>(null);
  const [explainOpen, setExplainOpen] = useState(false);
  const [explainLoading, setExplainLoading] = useState(false);
  const [explainError, setExplainError] = useState<string | null>(null);
  const [explainData, setExplainData] = useState<ProblemStatementExplainView | null>(null);

  const explainAvailable = Boolean(problemId?.trim());

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
      setExplainOpen(false);
      setExplainData(null);
      setExplainError(null);
      setExplainLoading(false);
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
    if (!open || !explainOpen || !problemId?.trim()) return;
    let cancelled = false;
    setExplainLoading(true);
    setExplainError(null);
    setExplainData(null);
    void fetchProblemStatementExplain(contestId, problemId)
      .then((data) => {
        if (!cancelled) {
          setExplainData(data);
          setExplainLoading(false);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setExplainError(err instanceof Error ? err.message : "Не удалось загрузить разбор");
          setExplainLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [open, explainOpen, contestId, problemId]);

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

  const toggleExplain = useCallback(() => {
    setExplainOpen((prev) => !prev);
  }, []);

  if (!mounted) return null;

  const ariaLabel = title ? `Условие — ${title}` : "Условие";

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
        {loading ? <p className="problem-statement-panel__status">Загрузка…</p> : null}
        {error ? <p className="problem-statement-panel__error">{error}</p> : null}
        {pdfData && !error ? (
          <div className="problem-statement-panel__layout">
            <div className="problem-statement-panel__main">
              <ProblemStatementViewer data={pdfData} problemLabel={problemLabel} />
            </div>
            {explainOpen ? (
              <ProblemStatementExplainPanel
                data={explainData}
                loading={explainLoading}
                error={explainError}
              />
            ) : null}
          </div>
        ) : null}
      </div>
    </div>,
    document.body,
  );
}
