import { Moon, Sun, SunMoon } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useThemeSettings } from "@/app/theme/ThemeProvider";
import type { ThemePreference } from "@/app/theme/themePreferences";

const THEME_LEVELS: ReadonlyArray<{
  value: ThemePreference;
  label: string;
  hint: string;
  Icon: LucideIcon;
}> = [
  { value: "system", label: "Системная", hint: "Как в настройках ОС", Icon: SunMoon },
  { value: "light", label: "Светлая", hint: "Светлая тема", Icon: Sun },
  { value: "dark", label: "Тёмная", hint: "Тёмная тема", Icon: Moon },
];

export default function ThemeSegmentControl() {
  const { preference, setPreference } = useThemeSettings();

  return (
    <nav
      className="chrome-segment-group chrome-segment-group--theme"
      role="radiogroup"
      aria-label="Тема"
    >
      {THEME_LEVELS.map((level, tipIndex) => {
        const active = preference === level.value;
        const { Icon } = level;
        return (
          <button
            key={level.value}
            type="button"
            role="radio"
            aria-checked={active}
            className={`chrome-segment-group-btn chrome-segment-btn--tip theme-segment-btn${
              active ? " is-active" : ""
            }`}
            aria-label={level.label}
            onClick={() => setPreference(level.value)}
          >
            <span className={`chrome-segment-tip theme-segment-tip--${tipIndex}`} role="tooltip">
              {level.hint}
            </span>
            <Icon className="theme-segment-btn__glyph" aria-hidden />
          </button>
        );
      })}
    </nav>
  );
}
