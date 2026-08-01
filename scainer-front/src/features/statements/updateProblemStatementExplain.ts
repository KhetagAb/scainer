import type {
  ProblemStatementExplainView,
  UpdateProblemStatementExplainRequest,
} from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";

export async function updateProblemStatementExplain(
  contestId: string,
  problemId: string,
  body: UpdateProblemStatementExplainRequest,
): Promise<ProblemStatementExplainView> {
  const res = await fetch(
    `/api/contests/${encodeURIComponent(contestId)}/problems/${encodeURIComponent(problemId)}/statement/explain`,
    {
      method: "PATCH",
      headers: {
        ...authHeaders(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    },
  );
  if (res.status === 404) {
    throw new Error("Не удалось сохранить формализацию для этой задачи");
  }
  if (!res.ok) {
    let message = "Не удалось сохранить формализацию";
    try {
      const payload = (await res.json()) as { error?: string };
      if (payload.error) message = payload.error;
    } catch {
      /* not json */
    }
    throw new Error(message);
  }
  return (await res.json()) as ProblemStatementExplainView;
}
