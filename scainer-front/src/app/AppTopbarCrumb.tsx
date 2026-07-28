import { FileText } from "lucide-react";
import { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { getContestsOptions } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import {
  UNGROUPED_PARALLEL,
  compactContestName,
} from "@/features/contests/shared/contestHelpers";
import { parallelLabel } from "@/features/contests/shared/parallels";
import ProblemStatementModal from "@/features/statements/ProblemStatementModal";

function decodePathSegment(segment: string): string {
  try {
    return decodeURIComponent(segment);
  } catch {
    return segment;
  }
}

export default function AppTopbarCrumb() {
  const { pathname } = useLocation();
  const [statementOpen, setStatementOpen] = useState(false);
  const contestsQuery = useQuery({
    ...getContestsOptions({ headers: authHeaders() }),
  });

  const parallelMatch = pathname.match(/^\/parallels\/([^/]+)\/?$/);
  if (parallelMatch) {
    const parallelId = decodePathSegment(parallelMatch[1]);
    return (
      <nav className="app-topbar__crumb" aria-label="Текущий раздел">
        <span className="app-topbar__crumb-name">
          {parallelLabel(parallelId, UNGROUPED_PARALLEL)}
        </span>
      </nav>
    );
  }

  const contestMatch = pathname.match(/^\/contests\/([^/]+)/);
  if (!contestMatch) return null;

  const contestId = decodePathSegment(contestMatch[1]);
  const contest = contestsQuery.data?.find((c) => c.id === contestId);
  if (!contest) return null;

  const parallelId = contest.parallelId || UNGROUPED_PARALLEL;
  const shortName = compactContestName(contest.name || "") || contest.id;
  const statementAvailable = Boolean(contest.parallelId);

  return (
    <>
      <nav className="app-topbar__crumb" aria-label="Текущий раздел">
        <Link
          to={`/parallels/${encodeURIComponent(parallelId)}`}
          className="app-topbar__crumb-parallel"
        >
          {parallelLabel(parallelId, UNGROUPED_PARALLEL)}
        </Link>
        <span className="app-topbar__crumb-sep" aria-hidden>
          /
        </span>
        <span className="app-topbar__crumb-name">{shortName}</span>
        {statementAvailable ? (
          <>
            <span className="app-topbar__crumb-sep" aria-hidden>
              /
            </span>
            <button
              type="button"
              className="app-topbar__crumb-statement"
              aria-label={`Условия задач — ${shortName}`}
              onClick={() => setStatementOpen(true)}
            >
              <FileText size={15} strokeWidth={2} aria-hidden />
              <span className="app-topbar__crumb-statement-label">Условия задач</span>
            </button>
          </>
        ) : null}
      </nav>
      {statementAvailable ? (
        <ProblemStatementModal
          open={statementOpen}
          contestId={contestId}
          title={shortName}
          onClose={() => setStatementOpen(false)}
        />
      ) : null}
    </>
  );
}
