import { useCallback, useEffect, useState } from "react";
import { Outlet, useNavigate, useParams } from "react-router-dom";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
import ContestStats from "@/features/contests/cards/ContestStats";
import { useContestPageData } from "@/features/contests/pages/useContestPageData";
import ContestSyncAction from "@/features/contests/sync/ContestSyncAction";
import { useContestSync } from "@/features/contests/sync/useContestSync";
import { useSensitivity } from "@/features/contests/shared/SensitivityContext";
import FindingsGroupTabs from "@/features/findings/FindingsGroupTabs";
import ReviewProblemPicker from "@/features/review/ReviewProblemPicker";
import ReviewLegendModal, {
  markReviewLegendSeen,
  wasReviewLegendSeen,
} from "@/features/review/ReviewLegendModal";
import { useReviewPrOnlyFilter } from "@/features/review/useReviewPrOnlyFilter";
import helpIconUrl from "@/assets/help-icon.png";

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
  const [reviewLegendOpen, setReviewLegendOpen] = useState(false);

  useEffect(() => {
    if (onReviewPage && !wasReviewLegendSeen()) {
      setReviewLegendOpen(true);
    }
  }, [onReviewPage]);

  const closeReviewLegend = useCallback(() => {
    markReviewLegendSeen();
    setReviewLegendOpen(false);
  }, []);

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
          "contest-head" + (onReviewPage ? " contest-head--review" : "")
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
              problems={(problemsQuery.data ?? []) as ProblemInfo[]}
              submissions={(submissionsQuery.data ?? []) as SubmissionListItem[]}
              activeProblemId={activeProblemId}
              prOnly={reviewPrOnly}
              onPrOnlyChange={setReviewPrOnly}
            />
          </div>
        ) : null}

        <div className="contest-head__actions">
          {headerStats && <ContestStats stats={headerStats} />}
          <span className="contest-head__sep" aria-hidden>
            |
          </span>
          <div className="contest-head__job-actions">
            <ContestSyncAction
              model={syncAction}
              progress={progress ?? undefined}
              disabled={jobBusy}
              onClick={startSync}
              onResync={startResync}
            />
          </div>
        </div>

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

      {onReviewPage ? (
        <>
          <div className="page-corner-actions">
            <button
              type="button"
              className="page-help-btn"
              aria-label="Справка по ревью посылок"
              title="Справка по ревью посылок"
              onClick={() => setReviewLegendOpen(true)}
            >
              <span
                className="page-help-btn__glyph"
                style={{
                  maskImage: `url(${helpIconUrl})`,
                  WebkitMaskImage: `url(${helpIconUrl})`,
                }}
                aria-hidden
              />
            </button>
          </div>
          <ReviewLegendModal open={reviewLegendOpen} onClose={closeReviewLegend} />
        </>
      ) : null}
    </>
  );
}
