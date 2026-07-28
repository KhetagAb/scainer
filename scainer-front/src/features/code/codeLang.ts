import type { Language } from "prism-react-renderer";

const PRISM_LANG: Record<string, Language> = {
  cpp: "cpp",
  c: "c",
  python: "python",
  java: "java",
  go: "go",
  javascript: "javascript",
  js: "javascript",
  typescript: "typescript",
  ts: "typescript",
};

export function toPrismLang(lang?: string): Language {
  if (!lang) return "clike";
  const key = lang.trim().toLowerCase();
  return PRISM_LANG[key] ?? "clike";
}
