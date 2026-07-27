import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import ContestIdCopy from "@/features/contests/cards/ContestIdCopy";
import ContestProblemChips from "@/features/contests/cards/ContestProblemChips";
import JobProgressBar from "@/features/contests/jobs/JobProgressBar";
import type { JobProgress } from "@/features/contests/jobs/jobShared";
import type { ContestFreshness } from "@/features/contests/sync/contestDataStatus";
import { contestNeedsSyncDot } from "@/features/contests/sync/contestDataStatus";
import type { ContestStatsFields } from "@/features/contests/shared/contestHelpers";
import type { ProblemSignalStat } from "@/features/contests/shared/problemSignalStats";
import {
  ContestFindingsIcon,
  ContestReviewIcon,
} from "@/features/contests/ui/ContestSectionNav";

const ACTION_ICON_SIZE = 40;

type StatsProps = {
  stats: ContestStatsFields;
  problemStats?: ProblemSignalStat[];
  syncProgress?: JobProgress;
  freshness?: ContestFreshness;
};

type Props = StatsProps & {
  id: string;
  name: string;
  displayName: string;
  pendingCount: number;
  hasStrongSignals: boolean;
  className?: string;
  /** Looks like a live card (links, hover) but clicks do nothing. */
  preview?: boolean;
  "aria-hidden"?: boolean;
};

function ActionSlot({
  preview,
  to,
  className,
  ariaLabel,
  children,
}: {
  preview?: boolean;
  to: string;
  className: string;
  ariaLabel: string;
  children: ReactNode;
}) {
  return (
    <Link
      to={to}
      className={className}
      aria-label={ariaLabel}
      tabIndex={preview ? -1 : undefined}
      onClick={preview ? (e) => e.preventDefault() : undefined}
      onKeyDown={
        preview
          ? (e) => {
              if (e.key === "Enter" || e.key === " ") e.preventDefault();
            }
          : undefined
      }
    >
      {children}
    </Link>
  );
}

export default function ContestCard({
  id,
  name,
  displayName,
  pendingCount,
  hasStrongSignals,
  stats,
  problemStats,
  syncProgress,
  freshness,
  className,
  preview,
  "aria-hidden": ariaHidden,
}: Props) {
  const reviewTo = `/contests/${encodeURIComponent(id)}/review`;
  const findingsTo = `/contests/${encodeURIComponent(id)}/findings`;
  const subs = stats.submissionCount ?? 0;
  const showStaleDot = freshness ? contestNeedsSyncDot(freshness) : false;
  const showFindingsMeta =
    syncProgress !== undefined ||
    (problemStats !== undefined && problemStats.length > 0);

  return (
    <div
      className={className ? `contest-card ${className}` : "contest-card"}
      aria-hidden={ariaHidden}
    >
      <div className="contest-card__top">
        <div className="contest-card__title">
          <span className="contest-card__name">{displayName || id}</span>
          {showStaleDot ? (
            <span className="contest-card__stale-dot" title="Данные устарели" aria-hidden />
          ) : null}
          {id !== name ? <ContestIdCopy id={id} preview={preview} /> : null}
        </div>
        <span className="contest-card__meta">{subs} посылок</span>
      </div>
      <div className="contest-card__actions">
        <ActionSlot
          preview={preview}
          to={reviewTo}
          className={
            "contest-card__action contest-card__action--review" +
            (pendingCount > 0 ? " contest-card__action--review-pending" : "")
          }
          ariaLabel={
            pendingCount > 0 ? `Ревью посылок, ${pendingCount} PR` : "Ревью посылок"
          }
        >
          <ContestReviewIcon size={ACTION_ICON_SIZE} />
          <span className="contest-card__action-label">Ревью посылок</span>
          {pendingCount > 0 ? (
            <span className="contest-card__nav-key contest-card__nav-key--violet">
              <span className="contest-card__nav-key__dot" />
              {pendingCount} PR
            </span>
          ) : null}
        </ActionSlot>
        <ActionSlot
          preview={preview}
          to={findingsTo}
          className={
            "contest-card__findings-side contest-card__action contest-card__action--findings" +
            (hasStrongSignals ? " contest-card__action--findings-green" : "")
          }
          ariaLabel="Детект"
        >
          <span className="contest-card__findings-main">
            <ContestFindingsIcon size={ACTION_ICON_SIZE} />
            <span className="contest-card__action-label">Детект</span>
          </span>
          {showFindingsMeta ? (
            <div className="contest-card__findings-meta">
              {syncProgress !== undefined ? (
                <JobProgressBar
                  progress={syncProgress}
                  className="import-progress-bar--card"
                />
              ) : null}
              {problemStats ? (
                <ContestProblemChips problemStats={problemStats} variant="card" />
              ) : null}
            </div>
          ) : null}
        </ActionSlot>
      </div>
    </div>
  );
}

export type { StatsProps as ContestCardStatsProps };
