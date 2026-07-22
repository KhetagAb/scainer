import { NavLink } from "react-router-dom";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
import { problemDisplay } from "@/features/findings/reportModel";
import { prCountByProblem } from "@/features/review/reviewFindings";

type Props = {
  contestId: string;
  problems: ProblemInfo[];
  submissions: SubmissionListItem[];
  activeProblemId: string | null;
};

export default function ReviewProblemPicker({
  contestId,
  problems,
  submissions,
  activeProblemId,
}: Props) {
  const prCounts = prCountByProblem(submissions);
  const sorted = problems.slice().sort((a, b) => {
    const la = problemDisplay(a.id, a.name);
    const lb = problemDisplay(b.id, b.name);
    return la < lb ? -1 : la > lb ? 1 : 0;
  });

  if (!sorted.length) {
    return <p className="review-picker__empty">Задач пока нет</p>;
  }

  const base = `/contests/${encodeURIComponent(contestId)}`;

  return (
    <ul className="review-picker" aria-label="Задачи для ревью">
      {sorted.map((p) => {
        // submissions актуальнее после смены verdict; pendingCount — fallback с /problems
        const pr = prCounts.get(p.id) ?? p.pendingCount ?? 0;
        const active = activeProblemId === p.id;
        const label = problemDisplay(p.id, p.name);
        return (
          <li key={p.id} className="review-picker__item">
            <NavLink
              to={`${base}/review?problem=${encodeURIComponent(p.id)}`}
              className={`chip problem-chip review-picker__chip${active ? " is-active" : ""}${pr === 0 ? " is-empty" : ""}`}
              title={label}
            >
              {label}
              <span className="review-picker__pr" aria-label={`${pr} pending review`}>
                {pr}
              </span>
            </NavLink>
          </li>
        );
      })}
    </ul>
  );
}
