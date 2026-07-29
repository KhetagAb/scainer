import type { SubmissionListItem } from "@/client/types.gen";
import type { ReviewFiltersInput, ReviewVerdictFilter } from "./reviewTypes";

function problemSortLabel(id: string, shortLabel?: string | null): string {
  if (shortLabel && id && shortLabel !== id) return `${shortLabel} - ${id}`;
  return shortLabel || id || "";
}

function isPendingReviewVerdict(verdict: string): boolean {
  const v = verdict.trim().toUpperCase();
  return v === "PR" || v === "PD";
}

export function matchesParticipantQuery(participant: string, query: string): boolean {
  const q = query.trim();
  if (!q) return true;
  if (q.length >= 2 && q.startsWith("/") && q.endsWith("/")) {
    try {
      return new RegExp(q.slice(1, -1), "i").test(participant);
    } catch {
      return participant.toLowerCase().includes(q.toLowerCase());
    }
  }
  return participant.toLowerCase().includes(q.toLowerCase());
}

export function matchesVerdictFilter(verdict: string, filter: ReviewVerdictFilter): boolean {
  if (!filter.active || !filter.verdicts.length) return true;
  const v = verdict.trim().toUpperCase();
  return filter.verdicts.some((code) => code.trim().toUpperCase() === v);
}

export function matchesSubmission(
  item: SubmissionListItem,
  filters: ReviewFiltersInput,
): boolean {
  return (
    matchesVerdictFilter(item.verdict, filters.verdictFilter) &&
    matchesParticipantQuery(item.participant, filters.participantQuery)
  );
}

export function indexSubmissionsByProblem(
  items: SubmissionListItem[],
): Map<string, SubmissionListItem[]> {
  const map = new Map<string, SubmissionListItem[]>();
  for (const item of items) {
    const list = map.get(item.problem);
    if (list) list.push(item);
    else map.set(item.problem, [item]);
  }
  return map;
}

export function countByProblem(
  items: SubmissionListItem[],
  filters: ReviewFiltersInput,
): Map<string, number> {
  const map = new Map<string, number>();
  for (const item of items) {
    if (!matchesSubmission(item, filters)) continue;
    map.set(item.problem, (map.get(item.problem) ?? 0) + 1);
  }
  return map;
}

export function queueForProblem(
  problemItems: SubmissionListItem[],
  filters: ReviewFiltersInput,
): SubmissionListItem[] {
  return problemItems
    .filter((s) => matchesSubmission(s, filters))
    .slice()
    .sort((a, b) => {
      const ap = isPendingReviewVerdict(a.verdict);
      const bp = isPendingReviewVerdict(b.verdict);
      if (ap !== bp) return ap ? -1 : 1;
      return a.submitted_at.localeCompare(b.submitted_at);
    });
}

export function hasHiddenSubmissions(
  problemItems: SubmissionListItem[],
  filters: ReviewFiltersInput,
): boolean {
  if (!filters.verdictFilter.active) return false;
  return problemItems.some(
    (s) =>
      !matchesVerdictFilter(s.verdict, filters.verdictFilter) &&
      matchesParticipantQuery(s.participant, filters.participantQuery),
  );
}

/** Оставшиеся PR по задаче — только для celebrate overlay, без UI-фильтров. */
export function prCountForCelebrate(problemItems: SubmissionListItem[]): number {
  let count = 0;
  for (const item of problemItems) {
    if (item.verdict === "PR") count++;
  }
  return count;
}

export function sortProblems<T extends { id: string; name?: string | null }>(
  problems: T[],
): T[] {
  return problems.slice().sort((a, b) => {
    const la = problemSortLabel(a.id, a.name);
    const lb = problemSortLabel(b.id, b.name);
    return la < lb ? -1 : la > lb ? 1 : 0;
  });
}

export function firstProblem(
  problems: { id: string; name?: string | null }[],
): string | null {
  const sorted = sortProblems(problems);
  return sorted[0]?.id ?? null;
}

export function firstProblemWithMatches(
  problems: { id: string; name?: string | null }[],
  items: SubmissionListItem[],
  filters: ReviewFiltersInput,
): string | null {
  const counts = countByProblem(items, filters);
  for (const p of sortProblems(problems)) {
    if ((counts.get(p.id) ?? 0) > 0) return p.id;
  }
  for (const [id, n] of counts) {
    if (n > 0) return id;
  }
  return null;
}

export function nextProblemWithMatches(
  problems: { id: string; name?: string | null }[],
  items: SubmissionListItem[],
  currentProblemId: string,
  filters: ReviewFiltersInput,
): string | null {
  const counts = countByProblem(items, filters);
  const sorted = sortProblems(problems);
  const idx = sorted.findIndex((p) => p.id === currentProblemId);
  const start = idx >= 0 ? idx + 1 : 0;
  for (let i = start; i < sorted.length; i++) {
    if ((counts.get(sorted[i].id) ?? 0) > 0) return sorted[i].id;
  }
  return null;
}
