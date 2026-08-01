import { authHeaders } from "@/features/auth/authStorage";

export async function deleteProblemStatementExplain(
  contestId: string,
  problemId: string,
): Promise<void> {
  const res = await fetch(
    `/api/contests/${encodeURIComponent(contestId)}/problems/${encodeURIComponent(problemId)}/statement/explain`,
    {
      method: "DELETE",
      headers: authHeaders(),
    },
  );
  if (res.status === 404) {
    throw new Error("Не удалось удалить формализацию для этой задачи");
  }
  if (!res.ok) {
    let message = "Не удалось удалить формализацию";
    try {
      const payload = (await res.json()) as { error?: string };
      if (payload.error) message = payload.error;
    } catch {
      /* not json */
    }
    throw new Error(message);
  }
}
