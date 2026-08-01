import { useEffect, useState } from "react";
import { extractRawStatementFromPrompt } from "@/features/statements/explainReadingReduction";
import { fetchProblemStatement } from "@/features/statements/fetchProblemStatement";

export function useExplainRawStatement(
  contestId: string,
  problemId: string,
  prompt: string | undefined,
): string | null {
  const [rawStatement, setRawStatement] = useState<string | null>(() =>
    prompt ? extractRawStatementFromPrompt(prompt) : null,
  );

  useEffect(() => {
    const fromPrompt = prompt ? extractRawStatementFromPrompt(prompt) : null;
    if (fromPrompt) {
      setRawStatement(fromPrompt);
      return;
    }

    let cancelled = false;
    void fetchProblemStatement(contestId, problemId)
      .then((ps) => {
        if (!cancelled) setRawStatement(ps.rawStatement.trim() || null);
      })
      .catch(() => {
        if (!cancelled) setRawStatement(null);
      });

    return () => {
      cancelled = true;
    };
  }, [prompt, contestId, problemId]);

  return rawStatement;
}
