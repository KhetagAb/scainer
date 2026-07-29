/** Palette + color-scheme preferences (localStorage). */

export type PaletteId = "arctic" | "creme";
export type ThemePreference = "light" | "dark" | "system";
export type ResolvedTheme = "light" | "dark";

export const PALETTE_STORAGE_KEY = "scainer-palette";
export const THEME_STORAGE_KEY = "scainer-theme";
export const DEFAULT_PALETTE: PaletteId = "arctic";
export const DEFAULT_THEME_PREFERENCE: ThemePreference = "system";

export function resolveSystemDark(): boolean {
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

export function resolveTheme(pref: ThemePreference): ResolvedTheme {
  if (pref === "dark") return "dark";
  if (pref === "light") return "light";
  return resolveSystemDark() ? "dark" : "light";
}

export function readPalette(): PaletteId {
  const v = localStorage.getItem(PALETTE_STORAGE_KEY);
  return v === "creme" ? "creme" : DEFAULT_PALETTE;
}

export function readThemePreference(): ThemePreference {
  const v = localStorage.getItem(THEME_STORAGE_KEY);
  if (v === "light" || v === "dark" || v === "system") return v;
  return DEFAULT_THEME_PREFERENCE;
}

export function applyThemeToDocument(palette: PaletteId, resolved: ResolvedTheme): void {
  const root = document.documentElement;
  root.dataset.palette = palette;
  root.dataset.theme = resolved;
  const meta = document.querySelector<HTMLMetaElement>('meta[name="theme-color"]');
  if (meta) {
    meta.content = getComputedStyle(root).getPropertyValue("--bg-canvas").trim() || "#f7f8fa";
  }
}

export function applyStoredTheme(): { palette: PaletteId; preference: ThemePreference; resolved: ResolvedTheme } {
  const palette = readPalette();
  const preference = readThemePreference();
  const resolved = resolveTheme(preference);
  applyThemeToDocument(palette, resolved);
  return { palette, preference, resolved };
}

export function persistPalette(palette: PaletteId): void {
  localStorage.setItem(PALETTE_STORAGE_KEY, palette);
}

export function persistThemePreference(preference: ThemePreference): void {
  localStorage.setItem(THEME_STORAGE_KEY, preference);
}
