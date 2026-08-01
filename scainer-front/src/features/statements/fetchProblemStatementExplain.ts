import type { ProblemStatementExplainView } from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";

const EXPLAIN_TIMEOUT_MS = 130_000;

export async function fetchProblemStatementExplain(
  contestId: string,
  problemId: string,
): Promise<ProblemStatementExplainView> {
  const res = await fetch(
    `/api/contests/${encodeURIComponent(contestId)}/problems/${encodeURIComponent(problemId)}/statement/explain`,
    {
      headers: authHeaders(),
      signal: AbortSignal.timeout(EXPLAIN_TIMEOUT_MS),
    },
  );
  if (res.status === 404) {
    throw new Error("Не удалось разобрать условие для этой задачи");
  }
  if (res.status === 503) {
    let message = "AI не настроен на сервере";
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      /* not json */
    }
    throw new Error(message);
  }
  if (res.status === 502) {
    throw new Error("Не удалось формализовать условие через AI");
  }
  if (!res.ok) {
    let message = "Не удалось загрузить формализацию";
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      /* not json */
    }
    throw new Error(message);
  }
  return (await res.json()) as ProblemStatementExplainView;
}
