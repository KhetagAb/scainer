const KEY_PREFIX = "scainer.explainFormalizationSeen:";
const LEGACY_KEY_PREFIX = "scainer.explainTypingSeen:";

function storageKey(contestId: string, problemId: string): string {
  return `${KEY_PREFIX}${contestId}/${problemId}`;
}

function legacyStorageKey(contestId: string, problemId: string): string {
  return `${LEGACY_KEY_PREFIX}${contestId}/${problemId}`;
}

export function wasExplainFormalizationSeen(contestId: string, problemId: string): boolean {
  try {
    const key = storageKey(contestId, problemId);
    if (localStorage.getItem(key) === "1") return true;

    const legacyKey = legacyStorageKey(contestId, problemId);
    if (localStorage.getItem(legacyKey) === "1") {
      localStorage.setItem(key, "1");
      localStorage.removeItem(legacyKey);
      return true;
    }

    return false;
  } catch {
    return true;
  }
}

export function markExplainFormalizationSeen(contestId: string, problemId: string): void {
  try {
    localStorage.setItem(storageKey(contestId, problemId), "1");
    localStorage.removeItem(legacyStorageKey(contestId, problemId));
  } catch {
    /* ignore quota / private mode */
  }
}

export function clearExplainFormalizationSeen(contestId: string, problemId: string): void {
  try {
    localStorage.removeItem(storageKey(contestId, problemId));
    localStorage.removeItem(legacyStorageKey(contestId, problemId));
  } catch {
    /* ignore */
  }
}
