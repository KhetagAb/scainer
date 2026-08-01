import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import {
  applyThemeToDocument,
  persistThemePreference,
  readThemePreference,
  resolveTheme,
  type ResolvedTheme,
  type ThemePreference,
} from "@/app/theme/themePreferences";

type ThemeContextValue = {
  preference: ThemePreference;
  resolved: ResolvedTheme;
  revision: number;
  setPreference: (preference: ThemePreference) => void;
};

const ThemeContext = createContext<ThemeContextValue | null>(null);

const TRANSITION_MS = 220;

function withColorTransition(run: () => void): void {
  const root = document.documentElement;
  const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  if (reduced) {
    run();
    return;
  }
  root.classList.add("theme-transition");
  run();
  window.setTimeout(() => root.classList.remove("theme-transition"), TRANSITION_MS);
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [preference, setPreferenceState] = useState<ThemePreference>(() => readThemePreference());
  const [resolved, setResolved] = useState<ResolvedTheme>(() => resolveTheme(readThemePreference()));
  const [revision, setRevision] = useState(0);

  const apply = useCallback((pref: ThemePreference, animate: boolean) => {
    const next = resolveTheme(pref);
    const run = () => {
      applyThemeToDocument(next);
      setPreferenceState(pref);
      setResolved(next);
      setRevision((r) => r + 1);
    };
    if (animate) withColorTransition(run);
    else run();
  }, []);

  useEffect(() => {
    const pref = readThemePreference();
    const next = resolveTheme(pref);
    applyThemeToDocument(next);
    setPreferenceState(pref);
    setResolved(next);
  }, []);

  useEffect(() => {
    if (preference !== "system") return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = () => {
      const next = resolveTheme("system");
      applyThemeToDocument(next);
      setResolved(next);
      setRevision((r) => r + 1);
    };
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, [preference]);

  const setPreference = useCallback(
    (pref: ThemePreference) => {
      persistThemePreference(pref);
      apply(pref, true);
    },
    [apply],
  );

  const value = useMemo(
    () => ({ preference, resolved, revision, setPreference }),
    [preference, resolved, revision, setPreference],
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useThemeSettings(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error("useThemeSettings must be used within ThemeProvider");
  return ctx;
}
