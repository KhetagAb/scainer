import { useCallback, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import helpIconUrl from "@/assets/help-icon.png";
import ContestCard from "@/features/contests/cards/ContestCard";
import ContestLegendModal, {
  markContestLegendSeen,
  wasContestLegendSeen,
} from "@/features/contests/legend/ContestLegendModal";
import { useParallelPageData } from "@/features/contests/pages/useParallelPageData";
import { UNGROUPED_PARALLEL } from "@/features/contests/shared/contestHelpers";
import { useSensitivity } from "@/features/contests/shared/SensitivityContext";
import AddContestForm from "@/features/contests/ui/AddContestForm";
import SensitivitySlider from "@/features/contests/ui/SensitivitySlider";
import { buildParallelSyncAction } from "@/features/contests/sync/contestDataStatus";
import ContestSyncAction from "@/features/contests/sync/ContestSyncAction";
import { useParallelSync } from "@/features/contests/sync/useParallelSync";

type Props = {
  onUnauthorized: () => void;
};

export default function ParallelPage({ onUnauthorized }: Props) {
  const { id: parallelIdParam } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const parallelId = parallelIdParam ? decodeURIComponent(parallelIdParam) : UNGROUPED_PARALLEL;
  const { threshold, setThreshold } = useSensitivity();
  const [legendOpen, setLegendOpen] = useState(() => !wasContestLegendSeen());

  const closeLegend = useCallback(() => {
    markContestLegendSeen();
    setLegendOpen(false);
  }, []);

  const { contestsQuery, parallelContests, contestIds, rows, checkUnauthorized } =
    useParallelPageData({ parallelId, threshold, onUnauthorized });

  const { syncJobs } = useParallelSync({ contestIds, onUnauthorized });

  if (!parallelIdParam) {
    navigate("/", { replace: true });
    return null;
  }

  if (contestsQuery.isLoading) {
    return <div className="page-center">Загрузка контестов…</div>;
  }
  if (contestsQuery.isError) {
    checkUnauthorized();
    return <div className="page-center login-error">Не удалось загрузить контесты</div>;
  }

  checkUnauthorized();

  const canAdd = parallelId !== UNGROUPED_PARALLEL;
  const syncAction = buildParallelSyncAction(parallelContests, {
    busy: syncJobs.isBusy,
  });

  return (
    <>
      <div className="contest-head">
        <div className="contest-head__actions">
          <ContestSyncAction
            model={syncAction}
            batchProgress={syncJobs.batchProgress}
            disabled={syncJobs.isBusy}
            onClick={() => syncJobs.startAll()}
          />
        </div>

        {syncJobs.error && (
          <p className="contest-head__error">
            Не удалось обновить: {syncJobs.error}
          </p>
        )}
      </div>

      <ul className="contest-grid">
        {rows.map((c) => (
          <li key={c.id}>
            <ContestCard
              id={c.id}
              name={c.name}
              displayName={c.displayName || c.id}
              pendingCount={c.pendingCount}
              hasStrongSignals={c.hasStrongSignals}
              stats={c.stats}
              problemStats={c.problemStats}
              freshness={c.freshness}
              statsLoading={c.statsLoading}
              syncProgress={
                c.id in syncJobs.progressById
                  ? syncJobs.progressById[c.id]
                  : undefined
              }
            />
          </li>
        ))}
        {canAdd && (
          <li className="contest-grid__add">
            <AddContestForm
              parallelId={parallelId}
              onDone={() => void contestsQuery.refetch()}
            />
          </li>
        )}
      </ul>

      {rows.length === 0 && !canAdd && <p className="empty-state">Нет контестов</p>}

      <div className="page-corner-actions">
        <SensitivitySlider threshold={threshold} onChange={setThreshold} />
        <button
          type="button"
          className="page-help-btn"
          aria-label="Что означают % и цвета на карточке контеста"
          title="Справка по карточке контеста"
          onClick={() => setLegendOpen(true)}
        >
          <span
            className="page-help-btn__glyph"
            style={{ maskImage: `url(${helpIconUrl})`, WebkitMaskImage: `url(${helpIconUrl})` }}
            aria-hidden
          />
        </button>
      </div>
      <ContestLegendModal open={legendOpen} onClose={closeLegend} />
    </>
  );
}
