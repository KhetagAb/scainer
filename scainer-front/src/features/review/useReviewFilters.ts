import { useCallback, useLayoutEffect, useState } from "react";
import { DEFAULT_VERDICT_FILTER, type ReviewVerdictFilter } from "@/features/review/reviewTypes";

const PARTICIPANT_STORAGE_KEY = "scainer.review.participantFilter";

function readParticipantQuery(): string {
  try {
    return localStorage.getItem(PARTICIPANT_STORAGE_KEY) ?? "";
  } catch {
    return "";
  }
}

/** Фильтры ревью: вердикт — только в рамках контеста; участник — в localStorage. */
export function useReviewFilters(contestId: string | undefined) {
  const [verdictFilter, setVerdictFilterState] = useState(DEFAULT_VERDICT_FILTER);
  const [participantQuery, setParticipantQueryState] = useState(readParticipantQuery);

  useLayoutEffect(() => {
    setVerdictFilterState(DEFAULT_VERDICT_FILTER);
  }, [contestId]);

  const setVerdictFilter = useCallback((value: ReviewVerdictFilter) => {
    setVerdictFilterState(value);
  }, []);

  const setParticipantQuery = useCallback((value: string) => {
    setParticipantQueryState(value);
    try {
      localStorage.setItem(PARTICIPANT_STORAGE_KEY, value);
    } catch {
      /* ignore quota / private mode */
    }
  }, []);

  return {
    verdictFilter,
    setVerdictFilter,
    participantQuery,
    setParticipantQuery,
  } as const;
}
