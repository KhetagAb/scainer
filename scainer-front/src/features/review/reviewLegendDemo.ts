import { problemDisplay } from "@/features/findings/reportModel";

export const REVIEW_LEGEND_DEMO_PROBLEMS = [
  { id: "shortest-paths", name: "A", pr: 3, active: true },
  { id: "minimum-spanning-tree", name: "B", pr: 0, active: false },
] as const;

export function reviewLegendProblemLabel(id: string, name: string): string {
  return problemDisplay(id, name);
}
