import { useMemo } from "react";
import { buildPrismTheme } from "@/app/theme/prismTheme";
import { useThemeSettings } from "@/app/theme/ThemeProvider";

/** Rebuild Prism theme when palette or color-scheme changes. */
export function usePrismTheme() {
  const { revision } = useThemeSettings();
  return useMemo(() => buildPrismTheme(), [revision]);
}
