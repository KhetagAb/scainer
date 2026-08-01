import type { ProblemStatementView } from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";

export async function fetchProblemStatement(
  contestId: string,
  problemId: string,
): Promise<ProblemStatementView> {
  const res = await fetch(
    `/api/contests/${encodeURIComponent(contestId)}/problems/${encodeURIComponent(problemId)}/statement`,
    { headers: authHeaders() },
  );
  if (res.status === 404) {
    throw new Error("Условие задачи не найдено");
  }
  if (!res.ok) {
    let message = "Не удалось загрузить условие задачи";
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      /* not json */
    }
    throw new Error(message);
  }
  return (await res.json()) as ProblemStatementView;
}
