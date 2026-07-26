import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import type { ContestStatsFields } from "@/features/contests/contestHelpers";
import type { ProblemSignalStat } from "@/features/contests/problemSignalStats";
import type { ImportProgress } from "@/features/contests/importJobShared";
import ContestIdCopy from "@/features/contests/ContestIdCopy";
import ContestProblemChips from "@/features/contests/ContestProblemChips";
import ImportProgressBar from "@/features/contests/ImportProgressBar";
import {
  ContestFindingsIcon,
  ContestReviewIcon,
} from "@/features/contests/ContestSectionNav";

const ACTION_ICON_SIZE = 40;

type StatsProps = {
  stats: ContestStatsFields;
  problemStats?: ProblemSignalStat[];
  importProgress?: ImportProgress;
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
  importProgress,
  className,
  preview,
  "aria-hidden": ariaHidden,
}: Props) {
  const reviewTo = `/contests/${encodeURIComponent(id)}/review`;
  const findingsTo = `/contests/${encodeURIComponent(id)}/findings`;
  const subs = stats.submissionCount ?? 0;
  const showFindingsMeta =
    importProgress !== undefined ||
    (problemStats !== undefined && problemStats.length > 0);

  return (
    <div
      className={className ? `contest-card ${className}` : "contest-card"}
      aria-hidden={ariaHidden}
    >
      <div className="contest-card__top">
        <div className="contest-card__title">
          <span className="contest-card__name">{displayName || id}</span>
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
              {importProgress !== undefined ? (
                <ImportProgressBar
                  progress={importProgress}
                  className="import-progress-bar--card"
                />
              ) : problemStats ? (
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
