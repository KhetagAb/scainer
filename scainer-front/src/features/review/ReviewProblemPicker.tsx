import { NavLink } from "react-router-dom";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
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
    const la = a.name || a.id;
    const lb = b.name || b.id;
    return la < lb ? -1 : la > lb ? 1 : 0;
  });

  if (!sorted.length) {
    return <p className="review-picker__empty">Задач пока нет</p>;
  }

  const base = `/contests/${encodeURIComponent(contestId)}`;

  return (
    <ul className="review-picker" aria-label="Задачи для ревью">
      {sorted.map((p) => {
        const pr = prCounts.get(p.id) ?? 0;
        const active = activeProblemId === p.id;
        return (
          <li key={p.id} className="review-picker__item">
            <NavLink
              to={`${base}/review?problem=${encodeURIComponent(p.id)}`}
              className={`review-picker__chip${active ? " is-active" : ""}${pr === 0 ? " is-empty" : ""}`}
              title={p.name || p.id}
            >
              <span className="review-picker__label">{p.name || p.id}</span>
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
