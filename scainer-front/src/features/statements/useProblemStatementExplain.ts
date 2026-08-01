import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useRef, useState, type MutableRefObject } from "react";
import type { ProblemStatementExplainView } from "@/client/types.gen";
import {
  clearExplainFormalizationSeen,
  markExplainFormalizationSeen,
  wasExplainFormalizationSeen,
} from "@/features/statements/explainFormalizationSeen";
import {
  fetchProblemStatementExplainQuery,
  getCachedProblemStatementExplain,
  removeProblemStatementExplainQuery,
  setCachedProblemStatementExplain,
} from "@/features/statements/explainQuery";
import type { FormalizationView } from "@/features/statements/formalizationView";
import { hasFormalization } from "@/features/statements/hasFormalization";
import { useTypewriterText } from "@/features/statements/useTypewriterText";

type Params = {
  open: boolean;
  contestId: string;
  problemId: string | null | undefined;
};

function applyPrefetchResult(
  data: ProblemStatementExplainView,
  contestId: string,
  problemKey: string,
  setExplainData: (data: ProblemStatementExplainView) => void,
  setExplainOpen: (open: boolean) => void,
  explainPrefetchRef: MutableRefObject<ProblemStatementExplainView | null>,
): void {
  if (!hasFormalization(data)) return;

  if (wasExplainFormalizationSeen(contestId, problemKey)) {
    setExplainData(data);
    setExplainOpen(true);
    return;
  }

  explainPrefetchRef.current = data;
}

export function useProblemStatementExplain({ open, contestId, problemId }: Params) {
  const queryClient = useQueryClient();
  const explainDataRef = useRef<ProblemStatementExplainView | null>(null);
  const explainPrefetchRef = useRef<ProblemStatementExplainView | null>(null);

  const [explainOpen, setExplainOpen] = useState(false);
  const [explainLoading, setExplainLoading] = useState(false);
  const [explainError, setExplainError] = useState<string | null>(null);
  const [explainData, setExplainData] = useState<ProblemStatementExplainView | null>(null);
  const [explainTypingSource, setExplainTypingSource] = useState<ProblemStatementExplainView | null>(
    null,
  );

  explainDataRef.current = explainData;

  const problemKey = problemId?.trim() ?? "";

  const completeTyping = useCallback(() => {
    if (!explainTypingSource || !problemKey) return;
    setExplainData(explainTypingSource);
    setExplainTypingSource(null);
    setExplainLoading(false);
    markExplainFormalizationSeen(contestId, problemKey);
  }, [contestId, problemKey, explainTypingSource]);

  const { displayed: typingText } = useTypewriterText(explainTypingSource?.statement ?? null, {
    onComplete: completeTyping,
  });

  const revealFormalization = useCallback(
    (data: ProblemStatementExplainView, contest: string, problem: string) => {
      explainPrefetchRef.current = null;
      if (wasExplainFormalizationSeen(contest, problem)) {
        setExplainData(data);
        setExplainLoading(false);
        return;
      }
      setExplainTypingSource(data);
    },
    [],
  );

  const resetExplainState = useCallback(() => {
    setExplainOpen(false);
    setExplainData(null);
    setExplainError(null);
    setExplainLoading(false);
    setExplainTypingSource(null);
    explainPrefetchRef.current = null;
  }, []);

  useEffect(() => {
    if (!open) {
      resetExplainState();
    }
  }, [open, resetExplainState]);

  useEffect(() => {
    if (!open || !problemKey) return;

    let cancelled = false;

    const cached = getCachedProblemStatementExplain(queryClient, contestId, problemKey);
    if (cached) {
      applyPrefetchResult(
        cached,
        contestId,
        problemKey,
        setExplainData,
        setExplainOpen,
        explainPrefetchRef,
      );
      return;
    }

    void fetchProblemStatementExplainQuery(queryClient, contestId, problemKey)
      .then((data) => {
        if (cancelled) return;
        applyPrefetchResult(
          data,
          contestId,
          problemKey,
          setExplainData,
          setExplainOpen,
          explainPrefetchRef,
        );
      })
      .catch(() => {
        /* повтор при ручном открытии панели explain */
      });

    return () => {
      cancelled = true;
    };
  }, [open, contestId, problemKey, queryClient]);

  useEffect(() => {
    setExplainData(null);
    setExplainTypingSource(null);
    setExplainError(null);
    setExplainLoading(false);
    explainPrefetchRef.current = null;
  }, [contestId, problemKey]);

  useEffect(() => {
    if (!open || !explainOpen || !problemKey) return;
    if (hasFormalization(explainDataRef.current)) return;
    if (explainTypingSource) return;

    const prefetched = explainPrefetchRef.current;
    if (prefetched && hasFormalization(prefetched)) {
      revealFormalization(prefetched, contestId, problemKey);
      return;
    }

    const cached = getCachedProblemStatementExplain(queryClient, contestId, problemKey);
    if (cached && hasFormalization(cached)) {
      revealFormalization(cached, contestId, problemKey);
      return;
    }

    let cancelled = false;

    setExplainLoading(true);
    setExplainError(null);

    void fetchProblemStatementExplainQuery(queryClient, contestId, problemKey)
      .then((data) => {
        if (cancelled) return;

        if (hasFormalization(explainDataRef.current)) {
          setExplainLoading(false);
          return;
        }

        if (hasFormalization(data)) {
          revealFormalization(data, contestId, problemKey);
          return;
        }

        setExplainData(data);
        setExplainLoading(false);
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          let message = "Не удалось загрузить формализацию";
          if (err instanceof Error) {
            if (err.name === "TimeoutError" || err.name === "AbortError") {
              message = "Формализация заняла слишком много времени";
            } else {
              message = err.message;
            }
          }
          setExplainError(message);
          setExplainLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [
    open,
    explainOpen,
    contestId,
    problemKey,
    queryClient,
    revealFormalization,
    explainTypingSource,
  ]);

  const toggleExplain = useCallback(() => {
    setExplainOpen((prev) => !prev);
  }, []);

  const closeOnDelete = useCallback(() => {
    resetExplainState();
    if (problemKey) {
      clearExplainFormalizationSeen(contestId, problemKey);
      removeProblemStatementExplainQuery(queryClient, contestId, problemKey);
    }
  }, [contestId, problemKey, queryClient, resetExplainState]);

  const updateExplainData = useCallback(
    (data: ProblemStatementExplainView) => {
      setExplainData(data);
      if (problemKey) {
        setCachedProblemStatementExplain(queryClient, contestId, problemKey, data);
      }
    },
    [contestId, problemKey, queryClient],
  );

  const formalizationView: FormalizationView = (() => {
    if (!explainOpen) return { phase: "hidden" };
    if (explainError) return { phase: "error", message: explainError };
    if (explainLoading && !hasFormalization(explainData) && !explainTypingSource) {
      return { phase: "fetching" };
    }
    if (explainTypingSource) {
      return { phase: "typing", data: explainTypingSource, text: typingText };
    }
    if (explainData && hasFormalization(explainData)) {
      return { phase: "ready", data: explainData };
    }
    return { phase: "fetching" };
  })();

  return {
    explainAvailable: Boolean(problemKey),
    explainOpen,
    formalizationView,
    toggleExplain,
    closeOnDelete,
    setExplainData: updateExplainData,
  };
}
