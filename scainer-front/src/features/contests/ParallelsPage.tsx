import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { getContestsOptions } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { DEFAULT_PARALLELS } from "@/features/contests/parallels";
import {
  UNGROUPED_PARALLEL,
  collator,
  compactContestName,
  contestsCountLabel,
} from "@/features/contests/contestHelpers";

type Props = {
  onUnauthorized: () => void;
};

type ContestChip = {
  id: string;
  label: string;
};

type ParallelCard = {
  id: string;
  name: string;
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
  const names = new Map<string, string>();

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
    if (c.parallelId && c.parallelName && !names.has(c.parallelId)) {
      names.set(c.parallelId, c.parallelName);
    }
  }

  for (const list of byParallel.values()) {
    list.sort((a, b) => collator.compare(a.id, b.id));
  }

  const extras = [...byParallel.keys()]
    .filter(
      (id) =>
        id !== UNGROUPED_PARALLEL && !DEFAULT_PARALLELS.some((p) => p.id === id),
    )
    .map((id) => ({
      id,
      name: names.get(id) || id,
      contests: byParallel.get(id) ?? [],
      submissionCount: submissionsByParallel.get(id) ?? 0,
    }))
    .sort((a, b) => collator.compare(a.name, b.name));

  const cards: ParallelCard[] = [
    ...DEFAULT_PARALLELS.map((p) => ({
      id: p.id,
      name: p.name,
      contests: byParallel.get(p.id) ?? [],
      submissionCount: submissionsByParallel.get(p.id) ?? 0,
    })),
    ...extras,
  ];

  const ungrouped = byParallel.get(UNGROUPED_PARALLEL);
  if (ungrouped?.length) {
    cards.push({
      id: UNGROUPED_PARALLEL,
      name: "Без параллели",
      contests: ungrouped,
      submissionCount: submissionsByParallel.get(UNGROUPED_PARALLEL) ?? 0,
    });
  }

  const maxContests = Math.max(0, ...cards.map((c) => c.contests.length));
  // Ориентир ~3 чипа в ряд — одинаковая min-height у всех блоков.
  const minRows = Math.max(1, Math.ceil(maxContests / 3));

  return (
    <ul
      className="parallels-grid"
      style={{ ["--parallels-min-rows" as string]: String(minRows) }}
    >
      {cards.map((p) => (
        <li key={p.id}>
          <Link
            to={`/parallels/${encodeURIComponent(p.id)}`}
            className="parallel-card"
          >
            <div className="parallel-card__head">
              <span className="parallel-card__title">{p.name}</span>
              <span className="parallel-card__count">
                {contestsCountLabel(p.contests.length)}
              </span>
            </div>
            <ul className="parallel-card__contests">
              {p.contests.map((c) => (
                <li key={c.id} className="chip parallel-card__chip" title={c.id}>
                  {c.label}
                </li>
              ))}
            </ul>
            <span className="parallel-card__subs">{p.submissionCount} посылок</span>
          </Link>
        </li>
      ))}
    </ul>
  );
}
