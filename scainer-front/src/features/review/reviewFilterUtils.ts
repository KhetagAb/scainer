import type { SubmissionListItem } from "@/client/types.gen";

export type ReviewVerdictFilter = {
  active: boolean;
  verdicts: string[];
};

export const DEFAULT_VERDICT_FILTER: ReviewVerdictFilter = {
  active: true,
  verdicts: ["PR"],
};

/** Порядок в выпадашке фильтра (OK в API отображается как AC). */
export const REVIEW_VERDICT_OPTIONS = [
  "PR",
  "OK",
  "RJ",
  "WA",
  "TL",
  "ML",
  "CE",
  "CF",
  "DQ",
] as const;

export type ReviewVerdictOption = (typeof REVIEW_VERDICT_OPTIONS)[number];

export type ReviewFiltersInput = {
  verdictFilter: ReviewVerdictFilter;
  participantQuery: string;
};

export function isReviewVerdictOption(value: string): value is ReviewVerdictOption {
  return (REVIEW_VERDICT_OPTIONS as readonly string[]).includes(value);
}

export function sortVerdicts(verdicts: string[]): string[] {
  const order = new Map<string, number>(
    REVIEW_VERDICT_OPTIONS.map((v, i) => [v, i]),
  );
  return verdicts
    .slice()
    .sort(
      (a, b) =>
        (order.get(a.toUpperCase()) ?? 99) - (order.get(b.toUpperCase()) ?? 99),
    );
}

/** Короткий код в чипе: OK → AC (как в ejudge UI). */
export function verdictFilterShortCode(verdict: string): string {
  const v = verdict.trim().toUpperCase();
  if (v === "OK") return "AC";
  return v;
}

export function formatVerdictFilterChipLabel(verdicts: string[]): string {
  return sortVerdicts(verdicts).map(verdictFilterShortCode).join(", ");
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

export function matchesReviewFilters(
  item: SubmissionListItem,
  filters: ReviewFiltersInput,
): boolean {
  return (
    matchesVerdictFilter(item.verdict, filters.verdictFilter) &&
    matchesParticipantQuery(item.participant, filters.participantQuery)
  );
}

export function verdictFiltersEqual(a: ReviewVerdictFilter, b: ReviewVerdictFilter): boolean {
  if (a.active !== b.active) return false;
  if (a.verdicts.length !== b.verdicts.length) return false;
  const sa = sortVerdicts(a.verdicts).join("\0");
  const sb = sortVerdicts(b.verdicts).join("\0");
  return sa === sb;
}

export function isPrOnlyVerdictFilter(filter: ReviewVerdictFilter): boolean {
  return (
    filter.active &&
    filter.verdicts.length === 1 &&
    filter.verdicts[0].trim().toUpperCase() === "PR"
  );
}
