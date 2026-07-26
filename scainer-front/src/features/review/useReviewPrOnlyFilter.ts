import { useCallback, useState } from "react";

const STORAGE_KEY = "scainer.review.prOnly";

function readStored(): boolean {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw === "0" || raw === "false") return false;
    if (raw === "1" || raw === "true") return true;
  } catch {
    /* ignore quota / private mode */
  }
  return true;
}

/** Фильтр «только задачи с PR» на странице ревью; по умолчанию включён. */
export function useReviewPrOnlyFilter() {
  const [prOnly, setPrOnlyState] = useState(readStored);

  const setPrOnly = useCallback((value: boolean) => {
    setPrOnlyState(value);
    try {
      localStorage.setItem(STORAGE_KEY, value ? "1" : "0");
    } catch {
      /* ignore quota / private mode */
    }
  }, []);

  return { prOnly, setPrOnly } as const;
}
