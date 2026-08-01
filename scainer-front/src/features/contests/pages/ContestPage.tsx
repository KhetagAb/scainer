import { useCallback, useMemo } from "react";
import { Outlet, useNavigate, useParams } from "react-router-dom";
import type { SubmissionListItem } from "@/client/types.gen";
import { useRegisterAppChrome } from "@/features/contests/shared/AppChromeContext";
import { useContestPageData } from "@/features/contests/pages/useContestPageData";
import ContestSyncAction from "@/features/contests/sync/ContestSyncAction";
import { useContestSync } from "@/features/contests/sync/useContestSync";
import { useSensitivity } from "@/features/contests/shared/SensitivityContext";
import FindingsGroupTabs from "@/features/findings/FindingsGroupTabs";
import ReviewProblemPicker from "@/features/review/ReviewProblemPicker";
import { useReviewFilters } from "@/features/review/useReviewFilters";
import type { ReviewFiltersInput } from "@/features/review/reviewFilterUtils";

type Props = {
  onUnauthorized: () => void;
};

export default function ContestPage({ onUnauthorized }: Props) {
  const { id: contestId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { onFindingsPage, groupBy, setGroupBy } = useSensitivity();

  const {
    verdictFilter,
    setVerdictFilter,
    participantFilter,
    setParticipantFilter,
  } = useReviewFilters(contestId);

  const reviewFiltersForData = useMemo<ReviewFiltersInput>(
    () => ({ verdictFilter, participantFilter }),
    [verdictFilter, participantFilter],
  );

  const pageData = useContestPageData({
    contestId,
    onUnauthorized,
    reviewFilters: reviewFiltersForData,
  });
  const {
    contestsQuery,
    contest,
    findingsQuery,
    problems,
    submissionsQuery,
    problemSubmissionCounts,
    activeProblemId,
    onReviewPage,
    findingKey,
    handleQueryError,
  } = pageData;

  const {
    syncAction,
    progress,
    isBusy: jobBusy,
    error: jobError,
    startSync,
    startResync,
  } = useContestSync({
    contestId,
    onUnauthorized,
  });

  const reviewFiltersForUi = useMemo<ReviewFiltersInput>(
    () => ({ verdictFilter, participantFilter }),
    [verdictFilter, participantFilter],
  );

  const handleStartSync = useCallback(() => {
    setParticipantFilter({ ...participantFilter, query: "" });
    startSync();
  }, [setParticipantFilter, participantFilter, startSync]);

  const handleStartResync = useCallback(() => {
    setParticipantFilter({ ...participantFilter, query: "" });
    startResync();
  }, [setParticipantFilter, participantFilter, startResync]);

  const syncNode = useMemo(
    () => (
      <ContestSyncAction
        model={syncAction}
        progress={progress ?? undefined}
        disabled={jobBusy}
        inTopbar
        onClick={handleStartSync}
        onResync={handleStartResync}
      />
    ),
    [
      syncAction,
      progress,
      jobBusy,
      handleStartSync,
      handleStartResync,
    ],
  );

  useRegisterAppChrome(syncNode);

  if (!contestId) {
    navigate("/", { replace: true });
    return null;
  }

  if (contestsQuery.isLoading) {
    return <div className="page-center">Загрузка контеста…</div>;
  }

  if (contestsQuery.isError || !contest) {
    handleQueryError();
    return (
      <div className="page-center login-error">
        {!contest ? "Контест не найден" : "Не удалось загрузить контест"}
      </div>
    );
  }

  return (
    <>
      {onReviewPage ? (
        <div
          className="contest-head contest-head--chrome contest-head--review"
          data-contest-head
        >
          <div className="contest-head__review-nav">
            <ReviewProblemPicker
              contestId={contestId}
              contestName={contest.name}
              problems={problems}
              submissions={(submissionsQuery.data ?? []) as SubmissionListItem[]}
              activeProblemId={activeProblemId}
              statementAvailable={Boolean(contest.parallelId)}
              reviewFilters={reviewFiltersForUi}
              verdictFilter={verdictFilter}
              onVerdictFilterChange={setVerdictFilter}
              participantFilter={participantFilter}
              onParticipantFilterChange={setParticipantFilter}
            />
          </div>
        </div>
      ) : onFindingsPage ? (
        <div className="contest-head contest-head--chrome">
          <div className="contest-head__findings-tabs">
            <FindingsGroupTabs groupBy={groupBy} onChange={setGroupBy} />
          </div>
        </div>
      ) : null}

      {jobError && (
        <p className="contest-head-error">
          Не удалось обновить: {jobError}
        </p>
      )}

      <Outlet
        context={{
          findingsQuery,
          submissionsQuery,
          problems,
          problemSubmissionCounts,
          contestName: contest.name,
          statementAvailable: Boolean(contest.parallelId),
          onUnauthorized,
          findingKey,
          reviewFilters: reviewFiltersForUi,
          setVerdictFilter,
        }}
      />
    </>
  );
}
