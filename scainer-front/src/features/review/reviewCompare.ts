import type { SubmissionListItem } from "@/client/types.gen";

export function parseRunId(submissionId: string): string | null {
  const parts = submissionId.split(":");
  if (parts.length !== 3 || parts[0] !== "ejudge") return null;
  return parts[2] || null;
}

export function submissionIdFromRunId(contestId: string, runId: string): string | null {
  const trimmed = runId.trim();
  if (!trimmed || !/^\d+$/.test(trimmed)) return null;
  return `ejudge:${contestId}:${trimmed}`;
}

export function participantSubmissionsOnProblem(
  items: SubmissionListItem[],
  problemId: string,
  participant: string,
): SubmissionListItem[] {
  return items
    .filter((s) => s.problem === problemId && s.participant === participant)
    .slice()
    .sort((a, b) => {
      const t = a.submitted_at.localeCompare(b.submitted_at);
      if (t !== 0) return t;
      return a.id.localeCompare(b.id);
    });
}

/** Предыдущая посылка того же участника по задаче (хронология). */
export function previousParticipantSubmission(
  items: SubmissionListItem[],
  current: SubmissionListItem,
): SubmissionListItem | null {
  const history = participantSubmissionsOnProblem(
    items,
    current.problem,
    current.participant,
  );
  const idx = history.findIndex((s) => s.id === current.id);
  if (idx <= 0) return null;
  return history[idx - 1] ?? null;
}

export function defaultCompareRunId(
  items: SubmissionListItem[],
  current: SubmissionListItem,
): string | null {
  const prev = previousParticipantSubmission(items, current);
  if (!prev) return null;
  return parseRunId(prev.id);
}

export function sourceLinesFromComments(
  source: string[] | undefined,
): string[] {
  if (!source?.length) return [];
  if (source.length === 1 && source[0] === "(нет исходника)") return [];
  return source;
}
