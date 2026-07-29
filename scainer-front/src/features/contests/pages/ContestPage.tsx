import { Outlet, useNavigate, useParams } from "react-router-dom";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
import { useContestPageData } from "@/features/contests/pages/useContestPageData";
import ContestSyncAction from "@/features/contests/sync/ContestSyncAction";
import { useContestSync } from "@/features/contests/sync/useContestSync";
import { useSensitivity } from "@/features/contests/shared/SensitivityContext";
import ContestHeadToolbar from "@/features/contests/ui/ContestHeadToolbar";
import FindingsGroupTabs from "@/features/findings/FindingsGroupTabs";
import ReviewProblemPicker from "@/features/review/ReviewProblemPicker";
import { useReviewPrOnlyFilter } from "@/features/review/useReviewPrOnlyFilter";

type Props = {
  onUnauthorized: () => void;
};

export default function ContestPage({ onUnauthorized }: Props) {
  const { id: contestId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { onFindingsPage, groupBy, setGroupBy } = useSensitivity();

  const pageData = useContestPageData({ contestId, onUnauthorized });
  const {
    contestsQuery,
    contest,
    findingsQuery,
    problemsQuery,
    submissionsQuery,
    problemSubmissionCounts,
    headerStats,
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

  const { prOnly: reviewPrOnly, setPrOnly: setReviewPrOnly } = useReviewPrOnlyFilter();

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
      <div
        className={
          "contest-head" +
          (onFindingsPage || onReviewPage ? " contest-head--chrome" : "") +
          (onReviewPage ? " contest-head--review" : "")
        }
        {...(onReviewPage ? { "data-contest-head": true } : {})}
      >
        {onFindingsPage ? (
          <div className="contest-head__findings-tabs">
            <FindingsGroupTabs groupBy={groupBy} onChange={setGroupBy} />
          </div>
        ) : null}

        {onReviewPage ? (
          <div className="contest-head__review-nav">
            <ReviewProblemPicker
              contestId={contestId}
              contestName={contest.name}
              problems={(problemsQuery.data ?? []) as ProblemInfo[]}
              submissions={(submissionsQuery.data ?? []) as SubmissionListItem[]}
              activeProblemId={activeProblemId}
              statementAvailable={Boolean(contest.parallelId)}
            />
          </div>
        ) : null}

        <ContestHeadToolbar
          submissionCount={headerStats?.submissionCount}
          showSensitivity={onFindingsPage || onReviewPage}
          prOnly={onReviewPage ? reviewPrOnly : undefined}
          onPrOnlyChange={onReviewPage ? setReviewPrOnly : undefined}
          sync={
            <ContestSyncAction
              model={syncAction}
              progress={progress ?? undefined}
              disabled={jobBusy}
              onClick={startSync}
              onResync={startResync}
            />
          }
        />

        {jobError && (
          <p className="contest-head__error">
            Не удалось обновить: {jobError}
          </p>
        )}
      </div>

      <Outlet
        context={{
          findingsQuery,
          submissionsQuery,
          problems: (problemsQuery.data ?? []) as ProblemInfo[],
          problemSubmissionCounts,
          contestName: contest.name,
          statementAvailable: Boolean(contest.parallelId),
          onUnauthorized,
          findingKey,
          reviewPrOnly,
          setReviewPrOnly,
        }}
      />
    </>
  );
}
