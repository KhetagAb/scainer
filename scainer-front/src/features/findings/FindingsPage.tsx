import { useMemo, useState, useEffect } from "react";
import type { UseQueryResult } from "@tanstack/react-query";
import type { FindingView, ReportData, SubmissionView } from "@/client/types.gen";
import { useSensitivity } from "@/features/contests/SensitivityContext";
import { collator } from "@/features/contests/contestHelpers";
import {
  countSuspiciousSubmissionsForProblem,
} from "@/features/contests/problemSignalStats";
import {
  FindingCard,
  GroupTitle,
  useFilteredVisibility,
} from "@/features/findings/FindingCard";
import { groupFindings, problemDisplay } from "@/features/findings/reportModel";

type Props = {
  findingKey?: string | null;
  /** Shared query from ContestPage so header chips reuse the same cache. */
  findingsQuery: UseQueryResult<unknown>;
  /** problemId → число посылок в store. */
  problemSubmissionCounts: Record<string, number>;
};

type HiddenMeta = {
  contest: string;
  contestName: string;
  problem: string;
  problemName: string;
};

export default function FindingsPage({
  findingKey,
  findingsQuery,
  problemSubmissionCounts,
}: Props) {
  const { threshold, groupBy } = useSensitivity();
  const [query, setQuery] = useState("");
  const [hiddenGroups, setHiddenGroups] = useState<Record<string, HiddenMeta>>({});

  const data = findingsQuery.data as ReportData | undefined;
  const findings = data?.findings ?? [];
  const submissions = (data?.submissions ?? {}) as Record<string, SubmissionView>;
  const groups = useMemo(() => groupFindings(findings, groupBy), [findings, groupBy]);
  const visibleKeys = useFilteredVisibility(findings, query, threshold);

  const groupProblemStats = useMemo(() => {
    const map = new Map<string, { suspicious: number; submissions: number }>();
    if (groupBy !== "problem") return map;
    for (const g of groups) {
      const submissions = problemSubmissionCounts[g.problem] ?? 0;
      map.set(g.key, {
        suspicious: countSuspiciousSubmissionsForProblem(
          g.findings as FindingView[],
          threshold,
          g.problem,
        ),
        submissions,
      });
    }
    return map;
  }, [groups, groupBy, threshold, problemSubmissionCounts]);

  useEffect(() => {
    setHiddenGroups({});
  }, [groupBy]);

  useEffect(() => {
    if (findingKey) {
      const element = document.querySelector(
        `[data-finding-key="${CSS.escape(findingKey)}"]`,
      ) as HTMLElement | null;
      if (element) {
        element.scrollIntoView({ behavior: "smooth", block: "center" });
        element.focus();
      }
    }
  }, [findingKey, findings]);

  if (findingsQuery.isLoading) {
    return <div className="page-center">Считаем находки…</div>;
  }
  if (findingsQuery.isError) {
    return <div className="page-center login-error">Не удалось загрузить находки</div>;
  }

  let visibleTotal = 0;

  return (
    <>
      <section className="controls-bar" aria-label="Поиск">
        <div className="controls-bar__search">
          <input
            type="search"
            placeholder="Поиск по участнику / задаче / посылке…"
            autoComplete="off"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>
      </section>

      {Object.keys(hiddenGroups).length > 0 ? (
        <div className="hidden-groups" aria-label="Скрытые задачи">
          <span className="hidden-groups-label">Скрытые задачи:</span>
          {Object.keys(hiddenGroups)
            .sort((a, b) => {
              const ma = hiddenGroups[a]!;
              const mb = hiddenGroups[b]!;
              const byLabel = collator.compare(
                problemDisplay(ma.problem || a, ma.problemName),
                problemDisplay(mb.problem || b, mb.problemName),
              );
              if (byLabel !== 0) return byLabel;
              return collator.compare(a, b);
            })
            .map((key) => {
              const meta = hiddenGroups[key]!;
              return (
                <button
                  key={key}
                  type="button"
                  className="hidden-group-chip"
                  onClick={() =>
                    setHiddenGroups((prev) => {
                      const next = { ...prev };
                      delete next[key];
                      return next;
                    })
                  }
                  title="Показать снова"
                >
                  {problemDisplay(meta.problem || key, meta.problemName)}
                </button>
              );
            })}
        </div>
      ) : null}

      <div className="findings">
        {groups.map((g) => {
          const userHidden = Boolean(hiddenGroups[g.key]);
          const visibleInGroup = g.findings.filter((f) => visibleKeys.has(f.key)).length;
          if (!userHidden) visibleTotal += visibleInGroup;
          const emptyAfterFilter = !userHidden && visibleInGroup === 0;
          const problemStat = groupProblemStats.get(g.key);

          return (
            <section
              key={g.key}
              className={`group${userHidden ? " group-user-hidden" : ""}${emptyAfterFilter ? " hidden-group" : ""}`}
            >
              <div className="group-head">
                <GroupTitle
                  groupBy={groupBy}
                  group={{
                    key: g.key,
                    contest: g.contest,
                    contestName: g.contestName,
                    problem: g.problem,
                    problemName: g.problemName,
                    visibleCount: visibleInGroup,
                  }}
                  suspiciousCount={problemStat?.suspicious}
                  submissionCount={problemStat?.submissions}
                />
                {groupBy === "problem" ? (
                  <button
                    type="button"
                    className="group-hide"
                    title="Скрыть задачу"
                    onClick={() =>
                      setHiddenGroups((prev) => ({
                        ...prev,
                        [g.key]: {
                          contest: g.contest,
                          contestName: g.contestName,
                          problem: g.problem,
                          problemName: g.problemName,
                        },
                      }))
                    }
                  >
                    скрыть
                  </button>
                ) : null}
              </div>
              <div className="group-cards">
                {g.findings.map((f) => (
                  <FindingCard
                    key={f.key}
                    finding={f}
                    groupBy={groupBy}
                    groupKey={g.key}
                    submissions={submissions}
                    hidden={!visibleKeys.has(f.key)}
                  />
                ))}
              </div>
            </section>
          );
        })}
      </div>

      <p className={`empty${visibleTotal > 0 || findings.length === 0 ? " hidden" : ""}`}>
        Нет сигналов по текущим фильтрам.
      </p>
    </>
  );
}
