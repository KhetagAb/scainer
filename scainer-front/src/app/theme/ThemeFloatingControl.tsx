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

export default function ThemeFloatingControl() {
  const { preference, setPreference } = useThemeSettings();
  const activeLevel = THEME_LEVELS.find((level) => level.value === preference) ?? THEME_LEVELS[0];
  const ActiveIcon = activeLevel.Icon;

  return (
    <div className="theme-floating">
      <nav className="theme-floating__menu" role="radiogroup" aria-label="Тема">
        {THEME_LEVELS.map((level) => {
          const active = preference === level.value;
          const { Icon } = level;

          return (
            <button
              key={level.value}
              type="button"
              role="radio"
              aria-checked={active}
              className={"theme-floating__option" + (active ? " is-active" : "")}
              aria-label={level.label}
              title={level.hint}
              onClick={(event) => {
                setPreference(level.value);
                event.currentTarget.blur();
              }}
            >
              <Icon className="theme-floating__glyph" aria-hidden />
            </button>
          );
        })}
      </nav>
      <button
        type="button"
        className="theme-floating__trigger"
        aria-label={`Тема: ${activeLevel.label}`}
        title={activeLevel.hint}
      >
        <ActiveIcon className="theme-floating__glyph" aria-hidden />
      </button>
    </div>
  );
}
