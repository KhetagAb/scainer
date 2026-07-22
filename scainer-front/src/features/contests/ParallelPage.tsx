import { useCallback, useMemo, useState } from "react";
import { useQueries, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  getContestFindingsOptions,
  getContestFindingsQueryKey,
  getContestProblemsOptions,
  getContestProblemsQueryKey,
  getContestsOptions,
} from "@/client/@tanstack/react-query.gen";
import type { FindingView, ProblemInfo, ReportData } from "@/client/types.gen";
import helpIconUrl from "@/assets/help-icon.png";
import importIconUrl from "@/assets/import-icon.png";
import { authHeaders } from "@/features/auth/authStorage";
import AddContestForm from "@/features/contests/AddContestForm";
import ContestLegendModal, {
  markContestLegendSeen,
  wasContestLegendSeen,
} from "@/features/contests/ContestLegendModal";
import ContestStats from "@/features/contests/ContestStats";
import ContestIdCopy from "@/features/contests/ContestIdCopy";
import {
  ContestFindingsIcon,
  ContestReviewIcon,
} from "@/features/contests/ContestSectionNav";
import ImportProgressBar from "@/features/contests/ImportProgressBar";
import SensitivitySlider from "@/features/contests/SensitivitySlider";
import { useSensitivity } from "@/features/contests/SensitivityContext";
import { parallelLabel } from "@/features/contests/parallels";
import {
  UNGROUPED_PARALLEL,
  compareContestId,
  compactContestName,
} from "@/features/contests/contestHelpers";
import {
  buildProblemSignalStats,
  contestHasStrongSignals,
  contestPendingCount,
} from "@/features/contests/problemSignalStats";
import { useParallelImportJobs } from "@/features/contests/useParallelImportJobs";

type Props = {
  onUnauthorized: () => void;
};

export default function ParallelPage({ onUnauthorized }: Props) {
  const { id: parallelIdParam } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const parallelId = parallelIdParam ? decodeURIComponent(parallelIdParam) : UNGROUPED_PARALLEL;
  const { threshold, setThreshold } = useSensitivity();
  const [legendOpen, setLegendOpen] = useState(() => !wasContestLegendSeen());

  const closeLegend = useCallback(() => {
    markContestLegendSeen();
    setLegendOpen(false);
  }, []);

  const contestsQuery = useQuery({
    ...getContestsOptions({ headers: authHeaders() }),
  });

  const contests = contestsQuery.data ?? [];
  const parallelContests = useMemo(
    () =>
      contests.filter((c) =>
        parallelId === UNGROUPED_PARALLEL ? !c.parallelId : c.parallelId === parallelId,
      ),
    [contests, parallelId],
  );

  const contestIds = useMemo(
    () => parallelContests.map((c) => c.id),
    [parallelContests],
  );

  const headers = authHeaders();
  const findingsQueries = useQueries({
    queries: parallelContests.map((c) => ({
      ...getContestFindingsOptions({ path: { id: c.id }, headers }),
    })),
  });
  const problemsQueries = useQueries({
    queries: parallelContests.map((c) => ({
      ...getContestProblemsOptions({ path: { id: c.id }, headers }),
    })),
  });

  const refreshContests = useCallback(() => {
    void contestsQuery.refetch();
  }, [contestsQuery]);

  // Схлопываем пачку onImportDone в один refetch contests.
  const refreshContestsDebounced = useMemo(() => {
    let timer: ReturnType<typeof setTimeout> | null = null;
    return () => {
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => {
        timer = null;
        refreshContests();
      }, 400);
    };
  }, [refreshContests]);

  const refreshContestCard = useCallback(
    async (id: string, ok: boolean) => {
      if (!ok) return;
      await Promise.all([
        contestsQuery.refetch(),
        queryClient.invalidateQueries({
          queryKey: getContestFindingsQueryKey({ path: { id }, headers: authHeaders() }),
        }),
        queryClient.invalidateQueries({
          queryKey: getContestProblemsQueryKey({ path: { id }, headers: authHeaders() }),
        }),
      ]);
    },
    [contestsQuery, queryClient],
  );

  const importJobs = useParallelImportJobs({
    contestIds,
    onUnauthorized,
    onImportDone: refreshContestsDebounced,
    onContestSettled: refreshContestCard,
    onSettled: refreshContests,
  });

  if (!parallelIdParam) {
    navigate("/", { replace: true });
    return null;
  }

  if (contestsQuery.isLoading) {
    return <div className="page-center">Загрузка контестов…</div>;
  }
  if (contestsQuery.isError) {
    if ((contestsQuery.error as { status?: number })?.status === 401) onUnauthorized();
    return <div className="page-center login-error">Не удалось загрузить контесты</div>;
  }

  for (const q of [...findingsQueries, ...problemsQueries]) {
    if (q.isError && (q.error as { status?: number })?.status === 401) {
      onUnauthorized();
      break;
    }
  }

  const canAdd = parallelId !== UNGROUPED_PARALLEL;
  const title = parallelLabel(parallelId, UNGROUPED_PARALLEL);

  const rows = parallelContests
    .map((c, i) => {
      const report = findingsQueries[i]?.data as ReportData | undefined;
      const findings = (report?.findings ?? []) as FindingView[];
      const problems = (problemsQueries[i]?.data ?? []) as ProblemInfo[];
      const ready = Boolean(findingsQueries[i]?.isSuccess && problemsQueries[i]?.isSuccess);
      const submissionCount = c.submissionCount ?? 0;
      const problemStats = ready
        ? buildProblemSignalStats(problems, findings, threshold)
        : undefined;

      return {
        id: c.id,
        name: c.name,
        displayName: compactContestName(c.name),
        hasStrongSignals: contestHasStrongSignals(problemStats),
        pendingCount: contestPendingCount(problemStats),
        stats: {
          id: c.id,
          submissionCount,
          problemCount: c.problemCount,
        },
        problemStats,
      };
    })
    .sort((a, b) => compareContestId(a.id, b.id));

  const importLabel = importJobs.isRunning
    ? importJobs.batchProgress
      ? `${importJobs.batchProgress.done}/${importJobs.batchProgress.total}`
      : "Загрузка…"
    : "Догрузить все контесты";

  return (
    <>
      <div className="contest-head">
        <h1>
          <span className="contest-head__name">{title}</span>
        </h1>
        <div className="contest-head__actions">
          {importJobs.isRunning && importJobs.batchProgress ? (
            <ImportProgressBar
              progress={{
                phase: "importing",
                done: importJobs.batchProgress.done,
                total: importJobs.batchProgress.total,
              }}
              label={`${importJobs.batchProgress.done}/${importJobs.batchProgress.total}`}
            />
          ) : null}
          <button
            type="button"
            className={`btn btn--primary btn--icon${importJobs.isRunning ? " is-busy" : ""}`}
            disabled={importJobs.isRunning || contestIds.length === 0}
            onClick={() => importJobs.startAll()}
            aria-label={importLabel}
            title={importLabel}
          >
            <span
              className="btn--icon__glyph"
              style={{ maskImage: `url(${importIconUrl})`, WebkitMaskImage: `url(${importIconUrl})` }}
              aria-hidden
            />
          </button>
        </div>

        {importJobs.error && (
          <p className="contest-head__error">
            Не удалось догрузить посылки: {importJobs.error}
          </p>
        )}
      </div>

      <ul className="contest-grid">
        {rows.map((c) => {
          const reviewTo = `/contests/${encodeURIComponent(c.id)}/review`;
          const findingsTo = `/contests/${encodeURIComponent(c.id)}/findings`;
          const title = c.name || c.id;
          return (
            <li key={c.id}>
              <div className="contest-card">
                <Link
                  to={reviewTo}
                  className="contest-card__hit"
                  title={title}
                  aria-label={c.displayName || c.id}
                />
                <div className="contest-card__top">
                  <div className="contest-card__title">
                    <span className="contest-card__name">{c.displayName || c.id}</span>
                    {c.id !== c.name ? <ContestIdCopy id={c.id} /> : null}
                  </div>
                  <div className="contest-card__nav">
                    <div className="contest-card__nav-icons">
                      <Link
                        to={reviewTo}
                        className={
                          "contest-card__nav-link contest-card__nav-link--review" +
                          (c.pendingCount > 0 ? " contest-card__nav-link--review-pending" : "")
                        }
                        aria-label={
                          c.pendingCount > 0
                            ? `Ревью, ${c.pendingCount} PR`
                            : "Ревью"
                        }
                      >
                        <ContestReviewIcon size={16} />
                      </Link>
                      <Link
                        to={findingsTo}
                        className={
                          "contest-card__nav-link contest-card__nav-link--findings" +
                          (c.hasStrongSignals ? " contest-card__nav-link--findings-green" : "")
                        }
                        aria-label="Детект"
                      >
                        <ContestFindingsIcon size={16} />
                      </Link>
                    </div>
                    <span className="contest-card__nav-hint contest-card__nav-hint--review" aria-hidden>
                      Ревью
                      {c.pendingCount > 0 ? (
                        <span className="contest-card__nav-key contest-card__nav-key--violet">
                          <span className="contest-card__nav-key__dot" />
                          {c.pendingCount} PR
                        </span>
                      ) : null}
                    </span>
                    <span className="contest-card__nav-hint contest-card__nav-hint--findings" aria-hidden>
                      Детект
                    </span>
                  </div>
                </div>
                <ContestStats
                  stats={c.stats}
                  problemStats={c.problemStats}
                  importProgress={
                    c.id in importJobs.progressById
                      ? importJobs.progressById[c.id]
                      : undefined
                  }
                />
              </div>
            </li>
          );
        })}
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
