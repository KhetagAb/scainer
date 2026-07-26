import { problemDisplay } from "@/features/findings/reportModel";

export const REVIEW_LEGEND_DEMO_PROBLEMS = [
  { id: "A", name: "Shortest Paths", pr: 3, active: true },
  { id: "B", name: "Minimum Spanning Tree", pr: 0, active: false },
] as const;

export function reviewLegendProblemLabel(id: string, name: string): string {
  return problemDisplay(id, name);
}
