import { useOutletContext } from "react-router-dom";
import type { UseQueryResult } from "@tanstack/react-query";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
import type { ReviewFiltersInput, ReviewVerdictFilter } from "@/features/review/reviewFilterUtils";
import FindingsPage from "@/features/findings/FindingsPage";
import ReviewPage from "@/features/review/ReviewPage";
import ReviewSubmissionPage from "@/features/review/ReviewSubmissionPage";

export type ContestOutletContext = {
  findingsQuery: UseQueryResult<unknown>;
  submissionsQuery: UseQueryResult<SubmissionListItem[]>;
  problems: ProblemInfo[];
  problemSubmissionCounts: Record<string, number>;
  contestName: string;
  statementAvailable: boolean;
  onUnauthorized: () => void;
  findingKey: string | null;
  reviewFilters: ReviewFiltersInput;
  setVerdictFilter: (value: ReviewVerdictFilter) => void;
};

export function ContestFindingsRoute() {
  const { findingsQuery, problemSubmissionCounts, findingKey } =
    useOutletContext<ContestOutletContext>();
  return (
    <FindingsPage
      findingKey={findingKey}
      findingsQuery={findingsQuery}
      problemSubmissionCounts={problemSubmissionCounts}
    />
  );
}

export function ContestReviewRoute() {
  const {
    submissionsQuery,
    findingsQuery,
    problems,
    onUnauthorized,
    reviewFilters,
    setVerdictFilter,
  } = useOutletContext<ContestOutletContext>();
  return (
    <ReviewPage
      submissionsQuery={submissionsQuery}
      findingsQuery={findingsQuery}
      problems={problems}
      onUnauthorized={onUnauthorized}
      reviewFilters={reviewFilters}
      setVerdictFilter={setVerdictFilter}
    />
  );
}

export function ContestReviewSubmissionRoute() {
  const { submissionsQuery } = useOutletContext<ContestOutletContext>();
  return <ReviewSubmissionPage submissionsQuery={submissionsQuery} />;
}
