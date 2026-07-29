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
  persistPalette,
  persistThemePreference,
  readPalette,
  readThemePreference,
  resolveTheme,
  type PaletteId,
  type ResolvedTheme,
  type ThemePreference,
} from "@/app/theme/themePreferences";

type ThemeContextValue = {
  palette: PaletteId;
  preference: ThemePreference;
  resolved: ResolvedTheme;
  revision: number;
  setPalette: (palette: PaletteId) => void;
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
  const [palette, setPaletteState] = useState<PaletteId>(() => readPalette());
  const [preference, setPreferenceState] = useState<ThemePreference>(() => readThemePreference());
  const [resolved, setResolved] = useState<ResolvedTheme>(() => resolveTheme(readThemePreference()));
  const [revision, setRevision] = useState(0);

  const apply = useCallback((p: PaletteId, pref: ThemePreference, animate: boolean) => {
    const next = resolveTheme(pref);
    const run = () => {
      applyThemeToDocument(p, next);
      setPaletteState(p);
      setPreferenceState(pref);
      setResolved(next);
      setRevision((r) => r + 1);
    };
    if (animate) withColorTransition(run);
    else run();
  }, []);

  useEffect(() => {
    const p = readPalette();
    const pref = readThemePreference();
    const resolved = resolveTheme(pref);
    applyThemeToDocument(p, resolved);
    setPaletteState(p);
    setPreferenceState(pref);
    setResolved(resolved);
  }, []);

  useEffect(() => {
    if (preference !== "system") return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = () => {
      const next = resolveTheme("system");
      applyThemeToDocument(palette, next);
      setResolved(next);
      setRevision((r) => r + 1);
    };
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, [palette, preference]);

  const setPalette = useCallback(
    (p: PaletteId) => {
      persistPalette(p);
      apply(p, preference, true);
    },
    [apply, preference],
  );

  const setPreference = useCallback(
    (pref: ThemePreference) => {
      persistThemePreference(pref);
      apply(palette, pref, true);
    },
    [apply, palette],
  );

  const value = useMemo(
    () => ({ palette, preference, resolved, revision, setPalette, setPreference }),
    [palette, preference, resolved, revision, setPalette, setPreference],
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useThemeSettings(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error("useThemeSettings must be used within ThemeProvider");
  return ctx;
}
