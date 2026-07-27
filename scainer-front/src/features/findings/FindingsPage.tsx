import { useMemo, useEffect } from "react";
import type { UseQueryResult } from "@tanstack/react-query";
import type { FindingView, ReportData, SubmissionView } from "@/client/types.gen";
import { useSensitivity } from "@/features/contests/shared/SensitivityContext";
import { countSignalsForProblem } from "@/features/contests/shared/problemSignalStats";
import {
  FindingRow,
  GroupTitle,
  useFilteredVisibility,
} from "@/features/findings/FindingCard";
import { groupFindings } from "@/features/findings/reportModel";

function chunk<T>(items: T[], size: number): T[][] {
  const out: T[][] = [];
  for (let i = 0; i < items.length; i += size) {
    out.push(items.slice(i, i + size));
  }
  return out;
}

type Props = {
  findingKey?: string | null;
  /** Shared query from ContestPage so header chips reuse the same cache. */
  findingsQuery: UseQueryResult<unknown>;
  /** problemId → число посылок в store. */
  problemSubmissionCounts: Record<string, number>;
};

export default function FindingsPage({
  findingKey,
  findingsQuery,
  problemSubmissionCounts,
}: Props) {
  const { threshold, groupBy } = useSensitivity();

  const data = findingsQuery.data as ReportData | undefined;
  const findings = data?.findings ?? [];
  const submissions = (data?.submissions ?? {}) as Record<string, SubmissionView>;
  const groups = useMemo(() => groupFindings(findings, groupBy), [findings, groupBy]);
  const visibleKeys = useFilteredVisibility(findings, threshold);

  const groupProblemStats = useMemo(() => {
    const map = new Map<string, { signals: number; submissions: number }>();
    if (groupBy !== "problem") return map;
    for (const g of groups) {
      const submissions = problemSubmissionCounts[g.problem] ?? 0;
      map.set(g.key, {
        signals: countSignalsForProblem(
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
    if (!findingKey) return;
    const scrollToFinding = () => {
      const element = document.querySelector(
        `[data-finding-key="${CSS.escape(findingKey)}"]`,
      ) as HTMLElement | null;
      if (element) {
        element.scrollIntoView({ behavior: "smooth", block: "center" });
        element.focus();
      }
    };
    const frame = requestAnimationFrame(scrollToFinding);
    return () => cancelAnimationFrame(frame);
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
      <div className="findings">
        {groups.map((g) => {
          const visibleInGroup = g.findings.filter((f) => visibleKeys.has(f.key)).length;
          visibleTotal += visibleInGroup;
          const emptyAfterFilter = visibleInGroup === 0;
          const problemStat = groupProblemStats.get(g.key);

          return (
            <section
              key={g.key}
              className={`group${emptyAfterFilter ? " hidden-group" : ""}`}
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
                  signalCount={problemStat?.signals}
                  submissionCount={problemStat?.submissions}
                />
              </div>
              <div className="group-cards">
                {chunk(
                  g.findings.filter((f) => visibleKeys.has(f.key)),
                  2,
                ).map((pair, rowIdx) => (
                  <FindingRow
                    key={pair.map((f) => f.key).join(":") || `row-${rowIdx}`}
                    pair={pair}
                    groupBy={groupBy}
                    groupKey={g.key}
                    submissions={submissions}
                    deepLinkKey={findingKey}
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
