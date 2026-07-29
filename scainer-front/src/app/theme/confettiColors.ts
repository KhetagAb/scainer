/** Confetti palette from CSS custom properties. */
export function readConfettiColors(): string[] {
  if (typeof document === "undefined") return [];
  const root = getComputedStyle(document.documentElement);
  return ["--confetti-1", "--confetti-2", "--confetti-3", "--confetti-4", "--confetti-5"]
    .map((name) => root.getPropertyValue(name).trim())
    .filter(Boolean);
}
