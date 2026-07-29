import { Moon, Palette, Sun, SunMoon } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useThemeSettings } from "@/app/theme/ThemeProvider";
import type { PaletteId, ThemePreference } from "@/app/theme/themePreferences";

const PALETTE_OPTIONS: { value: PaletteId; label: string; swatch: "arctic" | "creme" }[] = [
  { value: "arctic", label: "Arctic Slate", swatch: "arctic" },
  { value: "creme", label: "Creme", swatch: "creme" },
];

const THEME_OPTIONS: { value: ThemePreference; label: string; Icon: LucideIcon }[] = [
  { value: "system", label: "Системная", Icon: SunMoon },
  { value: "light", label: "Светлая", Icon: Sun },
  { value: "dark", label: "Тёмная", Icon: Moon },
];

export default function ThemeSettings() {
  const { palette, preference, setPalette, setPreference } = useThemeSettings();

  return (
    <div className="theme-settings">
      <button type="button" className="theme-settings__trigger" aria-label="Оформление">
        <Palette className="btn--icon__glyph" aria-hidden />
      </button>

      <div className="theme-settings__panel" aria-label="Настройки оформления">
        <div className="theme-settings__section">
          <p className="theme-settings__label">Палитра</p>
          <div
            className="theme-settings__group theme-settings__icon-group"
            role="group"
            aria-label="Палитра"
          >
            {PALETTE_OPTIONS.map(({ value, label, swatch }) => (
              <button
                key={value}
                type="button"
                className={`btn btn--icon theme-settings__icon-btn${palette === value ? " theme-settings__icon-btn--active" : ""}`}
                aria-label={label}
                aria-pressed={palette === value}
                onClick={() => setPalette(value)}
              >
                <span
                  className={`theme-settings__palette-swatch theme-settings__palette-swatch--${swatch}`}
                  aria-hidden
                />
              </button>
            ))}
          </div>
        </div>

        <div className="theme-settings__section">
          <p className="theme-settings__label">Яркость</p>
          <div
            className="theme-settings__group theme-settings__icon-group"
            role="group"
            aria-label="Яркость"
          >
            {THEME_OPTIONS.map(({ value, label, Icon }) => (
              <button
                key={value}
                type="button"
                className={`btn btn--icon theme-settings__icon-btn${preference === value ? " theme-settings__icon-btn--active" : ""}`}
                aria-label={label}
                aria-pressed={preference === value}
                onClick={() => setPreference(value)}
              >
                <Icon className="btn--icon__glyph" aria-hidden />
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
