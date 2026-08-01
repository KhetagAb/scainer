/** Color-scheme preferences (localStorage). */

export type ThemePreference = "light" | "dark" | "system";
export type ResolvedTheme = "light" | "dark";

export const THEME_STORAGE_KEY = "scainer-theme";
export const DEFAULT_THEME_PREFERENCE: ThemePreference = "system";

export function resolveSystemDark(): boolean {
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

export function resolveTheme(pref: ThemePreference): ResolvedTheme {
  if (pref === "dark") return "dark";
  if (pref === "light") return "light";
  return resolveSystemDark() ? "dark" : "light";
}

export function readThemePreference(): ThemePreference {
  const v = localStorage.getItem(THEME_STORAGE_KEY);
  if (v === "light" || v === "dark" || v === "system") return v;
  return DEFAULT_THEME_PREFERENCE;
}

export function applyThemeToDocument(resolved: ResolvedTheme): void {
  const root = document.documentElement;
  root.dataset.theme = resolved;
  const meta = document.querySelector<HTMLMetaElement>('meta[name="theme-color"]');
  if (meta) {
    meta.content = getComputedStyle(root).getPropertyValue("--bg-canvas").trim() || "#f7f8fa";
  }
}

export function applyStoredTheme(): { preference: ThemePreference; resolved: ResolvedTheme } {
  const preference = readThemePreference();
  const resolved = resolveTheme(preference);
  applyThemeToDocument(resolved);
  return { preference, resolved };
}

export function persistThemePreference(preference: ThemePreference): void {
  localStorage.setItem(THEME_STORAGE_KEY, preference);
}
