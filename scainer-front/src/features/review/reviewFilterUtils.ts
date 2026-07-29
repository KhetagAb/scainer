import type { SubmissionListItem } from "@/client/types.gen";
import {
  matchesParticipantQuery,
  matchesSubmission,
  matchesVerdictFilter,
} from "@/features/review/reviewModel";
import {
  REVIEW_VERDICT_OPTIONS,
  type ReviewFiltersInput,
  type ReviewVerdictFilter,
  type ReviewVerdictOption,
} from "@/features/review/reviewTypes";

export type { ReviewFiltersInput, ReviewVerdictFilter, ReviewVerdictOption };
export {
  DEFAULT_VERDICT_FILTER,
  REVIEW_VERDICT_OPTIONS,
} from "@/features/review/reviewTypes";

export { matchesParticipantQuery, matchesVerdictFilter };

export function matchesReviewFilters(
  item: SubmissionListItem,
  filters: ReviewFiltersInput,
): boolean {
  return matchesSubmission(item, filters);
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

export function verdictFiltersEqual(a: ReviewVerdictFilter, b: ReviewVerdictFilter): boolean {
  if (a.active !== b.active) return false;
  if (a.verdicts.length !== b.verdicts.length) return false;
  const sa = sortVerdicts(a.verdicts).join("\0");
  const sb = sortVerdicts(b.verdicts).join("\0");
  return sa === sb;
}

export function filtersEqual(a: ReviewFiltersInput, b: ReviewFiltersInput): boolean {
  return (
    verdictFiltersEqual(a.verdictFilter, b.verdictFilter) &&
    a.participantQuery === b.participantQuery
  );
}
