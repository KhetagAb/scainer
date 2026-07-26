import { useCallback, useEffect, useRef, useState } from "react";
import { getJob, postContestImport } from "@/client/sdk.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { subscribeJobEvents } from "@/features/contests/jobProgress";
import {
  clearStoredJobId,
  extractJobId,
  isAbortError,
  progressFromState,
  readStoredJobId,
  writeStoredJobId,
  type ImportProgress,
} from "@/features/contests/importJobShared";

export type { ImportProgress } from "@/features/contests/importJobShared";

type UseImportJobOptions = {
  /** Нужен для resume сохранённого jobId после remount. */
  contestId?: string | null;
  /** Вызывается по завершению job (start или resume), не при AbortError. */
  onSettled?: (ok: boolean) => void;
  /** Один раз при переходе в фазу analyzing — посылки уже в store. */
  onImportDone?: () => void;
};

/**
 * useImportJob — POST …/import → jobId + SSE-прогресс.
 * jobId кладётся в localStorage, чтобы после F5 подписка возобновлялась.
 */
export function useImportJob(
  onUnauthorized?: () => void,
  options?: UseImportJobOptions,
) {
  const contestId = options?.contestId ?? null;
  const onSettledRef = useRef(options?.onSettled);
  onSettledRef.current = options?.onSettled;
  const onImportDoneRef = useRef(options?.onImportDone);
  onImportDoneRef.current = options?.onImportDone;

  const [isRunning, setIsRunning] = useState(false);
  const [progress, setProgress] = useState<ImportProgress>(null);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);
  /** Не дать start и resume работать параллельно на одном contest. */
  const activeContestRef = useRef<string | null>(null);

  const watchJob = useCallback(
    async (id: string, jobId: string, signal: AbortSignal): Promise<boolean> => {
      let notifiedImportDone = false;
      const final = await subscribeJobEvents(
        jobId,
        (state) => {
          setProgress(progressFromState(state));
          if (!notifiedImportDone && state.progress.phase === "analyzing") {
            notifiedImportDone = true;
            onImportDoneRef.current?.();
          }
        },
        signal,
      );

      if (final.status === "succeeded" || final.status === "failed") {
        clearStoredJobId(id);
      }

      if (final.status === "failed") {
        setError(final.error || "импорт завершился с ошибкой");
        return false;
      }
      return true;
    },
    [],
  );

  const start = useCallback(
    async (id: string): Promise<boolean> => {
      if (activeContestRef.current) return false;

      setIsRunning(true);
      setError(null);
      setProgress(null);
      activeContestRef.current = id;

      const controller = new AbortController();
      abortRef.current = controller;
      let jobIdWritten = false;
      try {
        const res = await postContestImport({
          path: { id },
          headers: authHeaders(),
        });

        if (res.error) {
          if (res.response.status === 401) onUnauthorized?.();
          if (res.response.status === 404) clearStoredJobId(id);
          setError(res.error.error);
          onSettledRef.current?.(false);
          return false;
        }

        const jobId = extractJobId(res.data);
        if (!jobId) {
          setError("сервер не вернул jobId");
          onSettledRef.current?.(false);
          return false;
        }

        writeStoredJobId(id, jobId);
        jobIdWritten = true;
        const ok = await watchJob(id, jobId, controller.signal);
        onSettledRef.current?.(ok);
        return ok;
      } catch (e) {
        if (isAbortError(e)) return false;
        if (!jobIdWritten) clearStoredJobId(id);
        setError(e instanceof Error ? e.message : String(e));
        onSettledRef.current?.(false);
        return false;
      } finally {
        setIsRunning(false);
        abortRef.current = null;
        activeContestRef.current = null;
      }
    },
    [onUnauthorized, watchJob],
  );

  // Resume сохранённого job при входе на страницу контеста.
  useEffect(() => {
    if (!contestId) return;
    if (activeContestRef.current) return;

    const jobId = readStoredJobId(contestId);
    if (!jobId) return;

    let cancelled = false;
    const controller = new AbortController();
    abortRef.current = controller;
    activeContestRef.current = contestId;
    setIsRunning(true);
    setError(null);

    void (async () => {
      try {
        const snap = await getJob({
          path: { jobId },
          headers: authHeaders(),
        });

        if (cancelled) return;

        if (snap.error) {
          if (snap.response.status === 401) onUnauthorized?.();
          if (snap.response.status === 404) clearStoredJobId(contestId);
          else setError(snap.error.error);
          onSettledRef.current?.(false);
          return;
        }

        if (snap.response.status === 404 || !snap.data) {
          clearStoredJobId(contestId);
          return;
        }

        const status = snap.data.status;
        if (status === "succeeded" || status === "failed") {
          clearStoredJobId(contestId);
          if (status === "failed") {
            setError(snap.data.error || "импорт завершился с ошибкой");
            onSettledRef.current?.(false);
          } else {
            onSettledRef.current?.(true);
          }
          return;
        }

        setProgress(progressFromState(snap.data));
        const ok = await watchJob(contestId, jobId, controller.signal);
        if (!cancelled) onSettledRef.current?.(ok);
      } catch (e) {
        if (cancelled || isAbortError(e)) return;
        setError(e instanceof Error ? e.message : String(e));
        onSettledRef.current?.(false);
      } finally {
        if (!cancelled) {
          setIsRunning(false);
          abortRef.current = null;
          activeContestRef.current = null;
        }
      }
    })();

    return () => {
      cancelled = true;
      controller.abort();
      if (abortRef.current === controller) abortRef.current = null;
      if (activeContestRef.current === contestId) activeContestRef.current = null;
    };
  }, [contestId, onUnauthorized, watchJob]);

  const cancel = useCallback(() => {
    abortRef.current?.abort();
  }, []);

  return { start, cancel, isRunning, progress, error };
}
