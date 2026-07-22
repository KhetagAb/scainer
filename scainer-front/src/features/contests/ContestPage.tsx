import { useCallback, useMemo } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Link,
  NavLink,
  Outlet,
  useNavigate,
  useParams,
  useSearchParams,
  useLocation,
} from "react-router-dom";
import {
  getContestFindingsOptions,
  getContestFindingsQueryKey,
  getContestProblemsOptions,
  getContestProblemsQueryKey,
  getContestSubmissionsOptions,
  getContestSubmissionsQueryKey,
  getContestsOptions,
} from "@/client/@tanstack/react-query.gen";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";
import ContestStats from "@/features/contests/ContestStats";
import { UNGROUPED_PARALLEL } from "@/features/contests/contestHelpers";
import { formatImportProgress } from "@/features/contests/importJobShared";
import ImportProgressBar from "@/features/contests/ImportProgressBar";
import { problemSubmissionCountsMap } from "@/features/contests/problemSignalStats";
import { useImportJob } from "@/features/contests/useImportJob";
import ReviewProblemPicker from "@/features/review/ReviewProblemPicker";
import importIconUrl from "@/assets/import-icon.png";

type Props = {
  onUnauthorized: () => void;
};

function formatImportedTime(value?: string | null): string {
  if (!value) return "ещё не импортировался";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

export default function ContestPage({ onUnauthorized }: Props) {
  const { id: contestId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const location = useLocation();
  const queryClient = useQueryClient();

  const contestsQuery = useQuery({
    ...getContestsOptions({ headers: authHeaders() }),
  });

  const contest = contestId
    ? contestsQuery.data?.find((c) => c.id === contestId)
    : undefined;

  const findingsQuery = useQuery({
    ...getContestFindingsOptions({
      path: { id: contestId ?? "" },
      headers: authHeaders(),
    }),
    enabled: Boolean(contestId),
  });

  const problemsQuery = useQuery({
    ...getContestProblemsOptions({
      path: { id: contestId ?? "" },
      headers: authHeaders(),
    }),
    enabled: Boolean(contestId),
  });

  const submissionsQuery = useQuery({
    ...getContestSubmissionsOptions({
      path: { id: contestId ?? "" },
      headers: authHeaders(),
    }),
    enabled: Boolean(contestId),
  });

  const problemSubmissionCounts = useMemo(
    () => problemSubmissionCountsMap((problemsQuery.data ?? []) as ProblemInfo[]),
    [problemsQuery.data],
  );

  const refreshContests = useCallback(() => {
    void contestsQuery.refetch();
  }, [contestsQuery]);

  const refreshAfterImport = useCallback(
    async (ok: boolean) => {
      if (!ok || !contestId) return;
      await Promise.all([
        contestsQuery.refetch(),
        queryClient.invalidateQueries({
          queryKey: getContestFindingsQueryKey({
            path: { id: contestId },
            headers: authHeaders(),
          }),
        }),
        queryClient.invalidateQueries({
          queryKey: getContestProblemsQueryKey({
            path: { id: contestId },
            headers: authHeaders(),
          }),
        }),
        queryClient.invalidateQueries({
          queryKey: getContestSubmissionsQueryKey({
            path: { id: contestId },
            headers: authHeaders(),
          }),
        }),
      ]);
    },
    [contestId, contestsQuery, queryClient],
  );

  const importJob = useImportJob(onUnauthorized, {
    contestId,
    onImportDone: refreshContests,
    onSettled: (ok) => {
      void refreshAfterImport(ok);
    },
  });

  const headerStats = useMemo(() => {
    if (!contest) return null;
    return {
      id: contest.id,
      submissionCount: contest.submissionCount ?? 0,
      problemCount: contest.problemCount,
    };
  }, [contest]);

  const activeProblemId = useMemo(() => {
    const fromQuery = searchParams.get("problem");
    if (fromQuery) return fromQuery;
    const match = location.pathname.match(/\/review\/([^/]+)/);
    if (!match) return null;
    let sid = match[1];
    try {
      sid = decodeURIComponent(sid);
    } catch {
      /* keep */
    }
    const sub = (submissionsQuery.data ?? []).find((s) => s.id === sid);
    return sub?.problem ?? null;
  }, [searchParams, location.pathname, submissionsQuery.data]);

  if (!contestId) {
    navigate("/", { replace: true });
    return null;
  }

  if (contestsQuery.isLoading) {
    return <div className="page-center">Загрузка контеста…</div>;
  }

  if (contestsQuery.isError || !contest) {
    if ((contestsQuery.error as { status?: number })?.status === 401) onUnauthorized();
    return (
      <div className="page-center login-error">
        {!contest ? "Контест не найден" : "Не удалось загрузить контест"}
      </div>
    );
  }

  const importLabel = importJob.isRunning
    ? formatImportProgress(importJob.progress)
    : "Догрузить посылки";
  const backTo = `/parallels/${encodeURIComponent(contest.parallelId || UNGROUPED_PARALLEL)}`;
  const base = `/contests/${encodeURIComponent(contestId)}`;

  return (
    <>
      <Link to={backTo} className="back-link">
        ← К параллели
      </Link>

      <div className="contest-head">
        <h1>{contest.name || contest.id}</h1>
        <div className="contest-head__actions">
          {headerStats && <ContestStats stats={headerStats} />}
          <span className="contest-head__sep" aria-hidden>
            |
          </span>
          {importJob.isRunning ? (
            <ImportProgressBar progress={importJob.progress} />
          ) : (
            <span className="contest-head__meta">
              {formatImportedTime(contest.lastImportedAt)}
            </span>
          )}
          <button
            type="button"
            className={`btn btn--primary btn--icon${importJob.isRunning ? " is-busy" : ""}`}
            disabled={importJob.isRunning}
            onClick={() => void importJob.start(contestId)}
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

        {importJob.error && (
          <p className="contest-head__error">
            Не удалось догрузить посылки: {importJob.error}
          </p>
        )}
      </div>

      <nav className="contest-tabs" aria-label="Разделы контеста">
        <NavLink
          to={`${base}/findings`}
          className={({ isActive }) => `contest-tabs__link${isActive ? " is-active" : ""}`}
          end
        >
          Findings
        </NavLink>
        <NavLink
          to={`${base}/review`}
          className={({ isActive }) => `contest-tabs__link${isActive ? " is-active" : ""}`}
        >
          Review
        </NavLink>
      </nav>

      {location.pathname.includes("/review") ? (
        <ReviewProblemPicker
          contestId={contestId}
          problems={(problemsQuery.data ?? []) as ProblemInfo[]}
          submissions={(submissionsQuery.data ?? []) as SubmissionListItem[]}
          activeProblemId={activeProblemId}
        />
      ) : null}

      <Outlet
        context={{
          findingsQuery,
          submissionsQuery,
          problemSubmissionCounts,
          onUnauthorized,
          findingKey: searchParams.get("finding"),
        }}
      />
    </>
  );
}
