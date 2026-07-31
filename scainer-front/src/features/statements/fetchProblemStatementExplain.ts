import type { ProblemStatementExplainView } from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";

export async function fetchProblemStatementExplain(
  contestId: string,
  problemId: string,
): Promise<ProblemStatementExplainView> {
  const res = await fetch(
    `/api/contests/${encodeURIComponent(contestId)}/problems/${encodeURIComponent(problemId)}/statement/explain`,
    { headers: authHeaders() },
  );
  if (res.status === 404) {
    throw new Error("Не удалось разобрать условие для этой задачи");
  }
  if (!res.ok) {
    let message = "Не удалось загрузить разбор условия";
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
