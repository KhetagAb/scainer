import { useOutletContext } from "react-router-dom";
import type { UseQueryResult } from "@tanstack/react-query";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
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
  reviewPrOnly: boolean;
  setReviewPrOnly: (value: boolean) => void;
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
    reviewPrOnly,
    setReviewPrOnly,
  } = useOutletContext<ContestOutletContext>();
  return (
    <ReviewPage
      submissionsQuery={submissionsQuery}
      findingsQuery={findingsQuery}
      problems={problems}
      onUnauthorized={onUnauthorized}
      prOnly={reviewPrOnly}
      setReviewPrOnly={setReviewPrOnly}
    />
  );
}

export function ContestReviewSubmissionRoute() {
  const { submissionsQuery, findingsQuery, onUnauthorized } =
    useOutletContext<ContestOutletContext>();
  return (
    <ReviewSubmissionPage
      submissionsQuery={submissionsQuery}
      findingsQuery={findingsQuery}
      onUnauthorized={onUnauthorized}
    />
  );
}
