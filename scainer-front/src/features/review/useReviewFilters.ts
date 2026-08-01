import { useCallback, useLayoutEffect, useState } from "react";
import {
  DEFAULT_PARTICIPANT_FILTER,
  DEFAULT_VERDICT_FILTER,
  type ReviewFiltersInput,
  type ReviewParticipantFilter,
  type ReviewVerdictFilter,
} from "@/features/review/reviewTypes";

const FILTERS_STORAGE_KEY = "scainer.review.filtersByContest";
const LEGACY_PARTICIPANT_STORAGE_KEY = "scainer.review.participantFilter";

type StoredContestFilters = Partial<ReviewFiltersInput>;

type StoredFiltersByContest = Record<string, StoredContestFilters>;

function parseVerdictFilter(raw: unknown): ReviewVerdictFilter | null {
  if (!raw || typeof raw !== "object") return null;
  const value = raw as Partial<ReviewVerdictFilter>;
  if (typeof value.active !== "boolean" || !Array.isArray(value.verdicts)) return null;
  const verdicts = value.verdicts.filter((v): v is string => typeof v === "string");
  if (verdicts.length === 0) return null;
  return { active: value.active, verdicts };
}

function parseParticipantFilter(raw: unknown): ReviewParticipantFilter | null {
  if (!raw || typeof raw !== "object") return null;
  const value = raw as Partial<ReviewParticipantFilter>;
  if (typeof value.query !== "string") return null;
  return {
    active: value.active !== false,
    query: value.query,
  };
}

function readLegacyParticipantFilter(): ReviewParticipantFilter | null {
  try {
    const raw = localStorage.getItem(LEGACY_PARTICIPANT_STORAGE_KEY);
    if (!raw) return null;
    try {
      const parsed = JSON.parse(raw) as Partial<ReviewParticipantFilter>;
      if (parsed && typeof parsed.query === "string") {
        return {
          active: parsed.active !== false,
          query: parsed.query,
        };
      }
    } catch {
      return { active: true, query: raw };
    }
  } catch {
    /* ignore quota / private mode */
  }
  return null;
}

function readStoredFiltersByContest(): StoredFiltersByContest {
  try {
    const raw = localStorage.getItem(FILTERS_STORAGE_KEY);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as unknown;
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) return {};
    return parsed as StoredFiltersByContest;
  } catch {
    return {};
  }
}

function readContestFilters(contestId: string | undefined): ReviewFiltersInput {
  if (!contestId) {
    return {
      verdictFilter: DEFAULT_VERDICT_FILTER,
      participantFilter: DEFAULT_PARTICIPANT_FILTER,
    };
  }

  const stored = readStoredFiltersByContest()[contestId];
  const verdictFilter = parseVerdictFilter(stored?.verdictFilter) ?? DEFAULT_VERDICT_FILTER;
  const participantFilter =
    parseParticipantFilter(stored?.participantFilter) ??
    readLegacyParticipantFilter() ??
    DEFAULT_PARTICIPANT_FILTER;

  return { verdictFilter, participantFilter };
}

function writeContestFilters(contestId: string | undefined, filters: ReviewFiltersInput) {
  if (!contestId) return;
  try {
    const all = readStoredFiltersByContest();
    all[contestId] = {
      verdictFilter: filters.verdictFilter,
      participantFilter: filters.participantFilter,
    };
    localStorage.setItem(FILTERS_STORAGE_KEY, JSON.stringify(all));
  } catch {
    /* ignore quota / private mode */
  }
}

/** Фильтры ревью: вердикт и участник запоминаются отдельно для каждого контеста. */
export function useReviewFilters(contestId: string | undefined) {
  const [filters, setFiltersState] = useState(() => readContestFilters(contestId));

  useLayoutEffect(() => {
    setFiltersState(readContestFilters(contestId));
  }, [contestId]);

  const setVerdictFilter = useCallback(
    (value: ReviewVerdictFilter) => {
      setFiltersState((prev) => {
        const next = { ...prev, verdictFilter: value };
        writeContestFilters(contestId, next);
        return next;
      });
    },
    [contestId],
  );

  const setParticipantFilter = useCallback(
    (value: ReviewParticipantFilter) => {
      setFiltersState((prev) => {
        const next = { ...prev, participantFilter: value };
        writeContestFilters(contestId, next);
        return next;
      });
    },
    [contestId],
  );

  return {
    verdictFilter: filters.verdictFilter,
    setVerdictFilter,
    participantFilter: filters.participantFilter,
    setParticipantFilter,
  } as const;
}
