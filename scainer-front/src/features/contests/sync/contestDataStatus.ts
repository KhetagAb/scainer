export const IMPORT_STALE_MS = 10 * 60 * 1000;

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
};

const SYNC_HINT =
  "Догрузить посылки из ejudge и пересчитать детекторы.";
const SYNC_IDLE_LABEL = "Обновить";
export const SYNC_WARN_LOAD_LABEL = "Синхронизировать посылки";
export const SYNC_WARN_ANALYZE_LABEL = "Актуализировать анализ";

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
    return {
      state: "warn",
      label: warn.label,
      hint: SYNC_HINT,
      warnReason: warn.warnReason,
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
