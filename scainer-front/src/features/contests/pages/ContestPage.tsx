import { useCallback, useEffect, useState } from "react";
import { Link, Outlet, useNavigate, useParams } from "react-router-dom";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
import ContestStats from "@/features/contests/cards/ContestStats";
import { useContestPageData } from "@/features/contests/pages/useContestPageData";
import ContestSyncAction from "@/features/contests/sync/ContestSyncAction";
import { useContestSync } from "@/features/contests/sync/useContestSync";
import {
  UNGROUPED_PARALLEL,
  compactContestName,
} from "@/features/contests/shared/contestHelpers";
import { parallelLabel } from "@/features/contests/shared/parallels";
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

  const { syncAction, progress, isBusy, error, startSync } = useContestSync({
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

  const parallelId = contest.parallelId || UNGROUPED_PARALLEL;
  const parallelTo = `/parallels/${encodeURIComponent(parallelId)}`;
  const shortName = compactContestName(contest.name || "") || contest.id;

  return (
    <>
      <div className="contest-head">
        <h1>
          <Link to={parallelTo} className="contest-head__parallel">
            {parallelLabel(parallelId, UNGROUPED_PARALLEL)}
          </Link>
          <span className="contest-head__sep-title" aria-hidden>
            /
          </span>
          <span className="contest-head__name">{shortName}</span>
          {onFindingsPage ? (
            <FindingsGroupTabs groupBy={groupBy} onChange={setGroupBy} />
          ) : null}
        </h1>
        <div className="contest-head__actions">
          {headerStats && <ContestStats stats={headerStats} />}
          <span className="contest-head__sep" aria-hidden>
            |
          </span>
          <ContestSyncAction
            model={syncAction}
            progress={progress}
            disabled={isBusy}
            onClick={startSync}
          />
        </div>

        {error && (
          <p className="contest-head__error">
            Не удалось обновить: {error}
          </p>
        )}
      </div>

      {onReviewPage ? (
        <ReviewProblemPicker
          contestId={contestId}
          problems={(problemsQuery.data ?? []) as ProblemInfo[]}
          submissions={(submissionsQuery.data ?? []) as SubmissionListItem[]}
          activeProblemId={activeProblemId}
          prOnly={reviewPrOnly}
          onPrOnlyChange={setReviewPrOnly}
        />
      ) : null}

      <Outlet
        context={{
          findingsQuery,
          submissionsQuery,
          problems: (problemsQuery.data ?? []) as ProblemInfo[],
          problemSubmissionCounts,
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
