export const IMPORT_STALE_MS = 3 * 60 * 1000;

export type SyncActionState = "idle" | "warn" | "running" | "disabled";

export type SyncWarnReason = "no_import" | "import_stale" | "analyze_stale";

export type ContestFreshness = {
  lastImportedAt?: string | null;
  computedAt?: string | null;
};

export type SyncActionModel = {
  state: SyncActionState;
  label: string;
  hint: string;
  warnReason?: SyncWarnReason;
  /** Подробности для tooltip (например, «12 сек. назад»). */
  warnDetail?: string;
};

const SYNC_HINT =
  "Догрузить посылки из ejudge и пересчитать детекторы.";
const SYNC_IDLE_LABEL = "Обновить";
export const SYNC_WARN_LOAD_LABEL = "Синхронизировать посылки";
export const SYNC_WARN_ANALYZE_LABEL = "Актуализировать анализ";

export function syncWarnReasonLabel(reason: SyncWarnReason): string {
  switch (reason) {
    case "no_import":
      return "Посылки не загружались";
    case "import_stale":
      return "Посылки давно не обновлялись";
    case "analyze_stale":
      return "Анализ отстаёт";
  }
}

export function formatTimeAgo(iso: string | null | undefined, now = Date.now()): string | null {
  if (!iso) return null;
  const at = new Date(iso).getTime();
  if (Number.isNaN(at)) return null;
  const sec = Math.max(0, Math.floor((now - at) / 1000));
  if (sec < 60) return `${sec} сек. назад`;
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min} мин. назад`;
  const hours = Math.floor(min / 60);
  if (hours < 48) return `${hours} ч. назад`;
  const days = Math.floor(hours / 24);
  return `${days} д. назад`;
}

function syncWarnReasonDetail(
  reason: SyncWarnReason,
  contest: ContestFreshness,
  now = Date.now(),
): string {
  switch (reason) {
    case "no_import":
      return "Посылки ещё не загружались из ejudge.";
    case "import_stale": {
      const ago = formatTimeAgo(contest.lastImportedAt, now);
      return ago ? `Последнее обновление: ${ago}.` : "Последнее обновление неизвестно.";
    }
    case "analyze_stale": {
      const importAgo = formatTimeAgo(contest.lastImportedAt, now);
      const analyzeAgo = contest.computedAt
        ? formatTimeAgo(contest.computedAt, now)
        : null;
      if (importAgo && analyzeAgo) {
        return `Посылки: ${importAgo}. Анализ: ${analyzeAgo}.`;
      }
      if (analyzeAgo) return `Анализ: ${analyzeAgo}.`;
      return "Анализ ещё не запускался после последней загрузки.";
    }
  }
}

function pickDetailContest(
  contests: ContestFreshness[],
  reason: SyncWarnReason,
): ContestFreshness {
  if (contests.length === 1) return contests[0]!;

  if (reason === "no_import") {
    return contests.find((c) => !c.lastImportedAt) ?? contests[0]!;
  }

  if (reason === "import_stale") {
    let picked = contests[0]!;
    let oldest = Infinity;
    for (const c of contests) {
      if (!c.lastImportedAt) continue;
      const t = new Date(c.lastImportedAt).getTime();
      if (!Number.isNaN(t) && t < oldest) {
        oldest = t;
        picked = c;
      }
    }
    return picked;
  }

  for (const c of contests) {
    if (isAnalysisStale(c.lastImportedAt, c.computedAt)) return c;
  }
  return contests[0]!;
}

export const RESYNC_HINT =
  "Удалить все посылки и сигналы, затем заново загрузить из ejudge и пересчитать детекторы.";

export const RESYNC_CONFIRM =
  "Удалить все посылки и сигналы контеста и заново загрузить из ejudge с полным анализом?";

/** Анализ отстаёт от import (не «нужен AI»). */
export function isAnalysisStale(
  lastImportedAt?: string | null,
  computedAt?: string | null,
): boolean {
  if (!lastImportedAt) return false;
  if (!computedAt) return true;
  return new Date(lastImportedAt).getTime() > new Date(computedAt).getTime();
}

export function isImportStale(
  lastImportedAt?: string | null,
  now = Date.now(),
): boolean {
  if (!lastImportedAt) return true;
  const importedAt = new Date(lastImportedAt).getTime();
  return Number.isNaN(importedAt) || now - importedAt > IMPORT_STALE_MS;
}

type WarnPick = {
  warnReason: SyncWarnReason;
  label: string;
};

function pickWarnReason(
  contest: ContestFreshness,
  now = Date.now(),
): WarnPick | null {
  if (!contest.lastImportedAt) {
    return { warnReason: "no_import", label: SYNC_WARN_LOAD_LABEL };
  }
  if (isImportStale(contest.lastImportedAt, now)) {
    return { warnReason: "import_stale", label: SYNC_WARN_LOAD_LABEL };
  }
  if (isAnalysisStale(contest.lastImportedAt, contest.computedAt)) {
    return { warnReason: "analyze_stale", label: SYNC_WARN_ANALYZE_LABEL };
  }
  return null;
}

function pickParallelWarnReason(
  contests: ContestFreshness[],
  now = Date.now(),
): WarnPick | null {
  if (contests.length === 0) return null;
  if (contests.some((c) => !c.lastImportedAt)) {
    return { warnReason: "no_import", label: SYNC_WARN_LOAD_LABEL };
  }
  if (contests.some((c) => isImportStale(c.lastImportedAt, now))) {
    return { warnReason: "import_stale", label: SYNC_WARN_LOAD_LABEL };
  }
  if (contests.some((c) => isAnalysisStale(c.lastImportedAt, c.computedAt))) {
    return { warnReason: "analyze_stale", label: SYNC_WARN_ANALYZE_LABEL };
  }
  return null;
}

export function buildSyncAction(
  contest: ContestFreshness,
  opts?: { busy?: boolean; now?: number },
): SyncActionModel {
  if (opts?.busy) {
    return {
      state: "running",
      label: SYNC_IDLE_LABEL,
      hint: SYNC_HINT,
    };
  }

  const warn = pickWarnReason(contest, opts?.now);
  if (warn) {
    return {
      state: "warn",
      label: warn.label,
      hint: SYNC_HINT,
      warnReason: warn.warnReason,
      warnDetail: syncWarnReasonDetail(warn.warnReason, contest, opts?.now),
    };
  }

  return {
    state: "idle",
    label: SYNC_IDLE_LABEL,
    hint: SYNC_HINT,
  };
}

export function buildParallelSyncAction(
  contests: ContestFreshness[],
  opts?: { busy?: boolean; now?: number },
): SyncActionModel {
  if (contests.length === 0) {
    return { state: "disabled", label: SYNC_IDLE_LABEL, hint: SYNC_HINT };
  }

  if (opts?.busy) {
    return {
      state: "running",
      label: SYNC_IDLE_LABEL,
      hint: SYNC_HINT,
    };
  }

  const warn = pickParallelWarnReason(contests, opts?.now);
  if (warn) {
    const detailContest = pickDetailContest(contests, warn.warnReason);
    return {
      state: "warn",
      label: warn.label,
      hint: SYNC_HINT,
      warnReason: warn.warnReason,
      warnDetail: syncWarnReasonDetail(warn.warnReason, detailContest, opts?.now),
    };
  }

  return {
    state: "idle",
    label: SYNC_IDLE_LABEL,
    hint: SYNC_HINT,
  };
}

/** Dot на карточке контеста — любая проблема sync. */
export function contestNeedsSyncDot(
  contest: ContestFreshness,
  now = Date.now(),
): boolean {
  return pickWarnReason(contest, now) !== null;
}
