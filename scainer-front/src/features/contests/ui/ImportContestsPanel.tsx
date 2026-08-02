import { Radar } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useMutation } from "@tanstack/react-query";
import type { ImportContestsResponse } from "@/client/types.gen";
import { postParallelsImportMutation } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { UNGROUPED_PARALLEL } from "@/features/contests/shared/contestHelpers";
import JobProgressBar from "@/features/contests/jobs/JobProgressBar";
import { parallelLabel } from "@/features/contests/shared/parallels";

const parallelsImportMutationOptions = postParallelsImportMutation();

type Props = {
  onSuccess: () => void;
};

type Step = "loading" | "done" | "error";

export default function ImportContestsPanel({ onSuccess }: Props) {
  const [isOpen, setIsOpen] = useState(false);
  const [step, setStep] = useState<Step>("loading");
  const [result, setResult] = useState<ImportContestsResponse | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const cancelledRef = useRef(false);
  const pendingRefreshRef = useRef(false);
  const importStartedRef = useRef(false);
  const onSuccessRef = useRef(onSuccess);

  useEffect(() => {
    onSuccessRef.current = onSuccess;
  }, [onSuccess]);

  const { mutateAsync } = useMutation(parallelsImportMutationOptions);

  const runImport = useCallback(async () => {
    cancelledRef.current = false;
    setStep("loading");
    setErrorMessage(null);
    setResult(null);
    try {
      const data = await mutateAsync({
        body: {},
        headers: authHeaders(),
      });
      if (cancelledRef.current) return;
      pendingRefreshRef.current = true;
      setResult(data);
      setStep("done");
    } catch {
      if (cancelledRef.current) return;
      setErrorMessage("Не удалось импортировать контесты из ejudge");
      setStep("error");
    }
  }, [mutateAsync]);

  const close = useCallback(() => {
    cancelledRef.current = true;
    const shouldRefresh = pendingRefreshRef.current;
    pendingRefreshRef.current = false;
    setIsOpen(false);
    setStep("loading");
    setResult(null);
    setErrorMessage(null);
    if (shouldRefresh) {
      onSuccessRef.current();
    }
  }, []);

  useEffect(() => {
    if (!isOpen) {
      importStartedRef.current = false;
      return;
    }
    if (importStartedRef.current) return;
    importStartedRef.current = true;
    void runImport();
  }, [isOpen, runImport]);

  useEffect(() => {
    if (!isOpen) return;
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = prevOverflow;
    };
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen || step !== "loading") return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") close();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [isOpen, step, close]);

  if (!isOpen) {
    return (
      <button
        type="button"
        className="app-topbar-trailing-btn chrome-segment-btn--tip"
        aria-label="Импорт контестов"
        onClick={() => setIsOpen(true)}
      >
        <span className="chrome-segment-tip app-topbar-trailing-btn__tip" role="tooltip">
          Импорт контестов из ejudge
        </span>
        <Radar className="app-topbar-trailing-btn__glyph" strokeWidth={2} aria-hidden />
      </button>
    );
  }

  return createPortal(
    <div
      className="modal-overlay"
      onClick={(e) => e.target === e.currentTarget && step !== "loading" && close()}
    >
      <div className="form-card modal-card import-panel__card">
        {step === "loading" && (
          <div className="import-panel__loading">
            <JobProgressBar
              progress={null}
              label="Загружаем из ejudge…"
              className="import-progress-bar--modal"
            />
          </div>
        )}

        {step === "error" && (
          <>
            <p className="import-panel__summary login-error">{errorMessage}</p>
            <div className="form-actions">
              <button
                type="button"
                className="btn btn--primary"
                onClick={() => {
                  importStartedRef.current = true;
                  void runImport();
                }}
              >
                Повторить
              </button>
              <button type="button" className="btn btn--ghost" onClick={close}>
                Закрыть
              </button>
            </div>
          </>
        )}

        {step === "done" && result && (
          <>
            <p className="import-panel__summary">
              Импортировано: {result.added.length}
              {result.duplicates.length > 0 ? `, уже были добавлены: ${result.duplicates.length}` : ""}
              {result.skipped.length > 0 ? `, пропущено: ${result.skipped.length}` : ""}
            </p>
            <div className="modal-card__scroll">
              {result.added.length > 0 ? (
                <ul className="contest-list">
                  {result.added.map((row) => (
                    <li key={row.id} className="contest-row">
                      <div>
                        <div className="contest-row__id">{row.name}</div>
                        <div className="contest-row__meta">ID: {row.id}</div>
                      </div>
                      <span className="chip contest-chip">
                        {parallelLabel(row.parallelId, UNGROUPED_PARALLEL)}
                      </span>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="muted" style={{ margin: 0 }}>
                  Новых контестов нет.
                </p>
              )}
            </div>
            <div className="form-actions">
              <button type="button" className="btn" onClick={close}>
                Закрыть
              </button>
            </div>
          </>
        )}
      </div>
    </div>,
    document.body,
  );
}
