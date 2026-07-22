/** База ejudge UI (view-source / master). */
const DEFAULT_EJUDGE_BASE = "https://ejudge.lksh.ru";

export function parseEjudgeSubmissionId(
  id: string,
): { contestId: number; runId: number } | null {
  const m = /^ejudge:(\d+):(\d+)$/.exec(id.trim());
  if (!m) return null;
  return { contestId: Number(m[1]), runId: Number(m[2]) };
}

/**
 * Ссылка на посылку в ejudge master (action=36 = view-source).
 * Паттерн: /cgi-bin/new-master?SID=…&action=36&run_id=…
 * SID опционален (VITE_EJUDGE_SID); без него добавляем contest_id.
 */
export function ejudgeRunUrl(submissionId: string): string | null {
  const key = parseEjudgeSubmissionId(submissionId);
  if (!key) return null;
  const raw = import.meta.env.VITE_EJUDGE_BASE_URL as string | undefined;
  const base = (raw?.replace(/\/$/, "") || DEFAULT_EJUDGE_BASE).replace(/\/$/, "");
  const sid = (import.meta.env.VITE_EJUDGE_SID as string | undefined)?.trim();
  const q = new URLSearchParams();
  if (sid) q.set("SID", sid);
  else q.set("contest_id", String(key.contestId));
  q.set("action", "36");
  q.set("run_id", String(key.runId));
  return `${base}/cgi-bin/new-master?${q.toString()}`;
}
