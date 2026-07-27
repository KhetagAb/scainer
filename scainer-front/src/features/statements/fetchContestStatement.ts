import { authHeaders } from "@/features/auth/authStorage";

export type StatementLoadState =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "ready"; data: ArrayBuffer }
  | { status: "error"; message: string };

export async function fetchContestStatement(contestId: string): Promise<ArrayBuffer> {
  const res = await fetch(`/api/contests/${encodeURIComponent(contestId)}/statement`, {
    headers: authHeaders(),
  });
  if (res.status === 404) {
    throw new Error("Условие недоступно для этого контеста");
  }
  if (!res.ok) {
    let message = "Не удалось загрузить условие";
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      /* not json */
    }
    throw new Error(message);
  }
  return res.arrayBuffer();
}
