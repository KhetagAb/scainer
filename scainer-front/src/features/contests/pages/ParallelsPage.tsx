import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { getContestsOptions } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import {
  UNGROUPED_PARALLEL,
  collator,
  compactContestName,
  contestsCountLabel,
} from "@/features/contests/shared/contestHelpers";
import { DEFAULT_PARALLEL_IDS, parallelLabel } from "@/features/contests/shared/parallels";
import HomeStatsBar from "@/features/contests/ui/HomeStatsBar";

type Props = {
  onUnauthorized: () => void;
};

type ContestChip = {
  id: string;
  label: string;
};

type ParallelCard = {
  id: string;
  contests: ContestChip[];
  submissionCount: number;
};

export default function ParallelsPage({ onUnauthorized }: Props) {
  const contestsQuery = useQuery({
    ...getContestsOptions({ headers: authHeaders() }),
  });

  if (contestsQuery.isLoading) {
    return <div className="page-center">Загрузка параллелей…</div>;
  }
  if (contestsQuery.isError) {
    if ((contestsQuery.error as { status?: number })?.status === 401) onUnauthorized();
    return <div className="page-center login-error">Не удалось загрузить параллели</div>;
  }

  const contests = contestsQuery.data ?? [];
  const byParallel = new Map<string, ContestChip[]>();
  const submissionsByParallel = new Map<string, number>();

  for (const c of contests) {
    const pid = c.parallelId || UNGROUPED_PARALLEL;
    const list = byParallel.get(pid) ?? [];
    list.push({
      id: c.id,
      label: compactContestName(c.name) || c.id,
    });
    byParallel.set(pid, list);
    submissionsByParallel.set(
      pid,
      (submissionsByParallel.get(pid) ?? 0) + (c.submissionCount ?? 0),
    );
  }

  for (const list of byParallel.values()) {
    list.sort((a, b) => collator.compare(a.id, b.id));
  }

  const extras = [...byParallel.keys()]
    .filter(
      (id) =>
        id !== UNGROUPED_PARALLEL &&
        !(DEFAULT_PARALLEL_IDS as readonly string[]).includes(id),
    )
    .map((id) => ({
      id,
      contests: byParallel.get(id) ?? [],
      submissionCount: submissionsByParallel.get(id) ?? 0,
    }))
    .sort((a, b) => collator.compare(a.id, b.id));

  const cards: ParallelCard[] = [
    ...DEFAULT_PARALLEL_IDS.map((id) => ({
      id,
      contests: byParallel.get(id) ?? [],
      submissionCount: submissionsByParallel.get(id) ?? 0,
    })),
    ...extras,
  ];

  const ungrouped = byParallel.get(UNGROUPED_PARALLEL);
  if (ungrouped?.length) {
    cards.push({
      id: UNGROUPED_PARALLEL,
      contests: ungrouped,
      submissionCount: submissionsByParallel.get(UNGROUPED_PARALLEL) ?? 0,
    });
  }

  const maxContests = Math.max(0, ...cards.map((c) => c.contests.length));
  const minRows = Math.max(1, Math.ceil(maxContests / 3));
  const totalContests = contests.length;
  const totalSubmissions = contests.reduce((s, c) => s + (c.submissionCount ?? 0), 0);

  return (
    <>
      <HomeStatsBar contestCount={totalContests} submissionCount={totalSubmissions} />
      <ul
      className="parallels-grid"
      style={{ ["--parallels-min-rows" as string]: String(minRows) }}
    >
      {cards.map((p) => {
        const title = parallelLabel(p.id, UNGROUPED_PARALLEL);
        return (
          <li key={p.id}>
            <div className="parallel-card">
              <Link
                to={`/parallels/${encodeURIComponent(p.id)}`}
                className="parallel-card__nav"
                aria-label={title}
              />
              <div className="parallel-card__head">
                <span className="parallel-card__title">{title}</span>
                <span className="parallel-card__count">
                  {contestsCountLabel(p.contests.length)}
                </span>
              </div>
              <div className="parallel-card__contests">
                {p.contests.map((c) => (
                  <Link
                    key={c.id}
                    to={`/contests/${encodeURIComponent(c.id)}/review`}
                    className="chip parallel-card__chip"
                    title={c.id}
                  >
                    {c.label}
                  </Link>
                ))}
              </div>
              {p.submissionCount > 0 ? (
                <div className="parallel-card__subs">{p.submissionCount} посылок</div>
              ) : null}
            </div>
          </li>
        );
      })}
    </ul>
    </>
  );
}
