import type { PrismTheme } from "prism-react-renderer";

function cssVar(name: string, fallback: string): string {
  if (typeof document === "undefined") return fallback;
  const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  return v || fallback;
}

/** Token-driven Prism theme; call when theme revision changes. */
export function buildPrismTheme(): PrismTheme {
  return {
    plain: {
      color: cssVar("--prism-fg", "#3a3d45"),
      backgroundColor: cssVar("--prism-bg", "#f7f8fa"),
    },
    styles: [
      {
        types: ["comment", "prolog", "doctype", "cdata"],
        style: { color: cssVar("--prism-comment", "#7a7f88") },
      },
      {
        types: ["namespace"],
        style: { opacity: 0.7 },
      },
      {
        types: ["string", "char", "attr-value", "regex", "inserted"],
        style: { color: cssVar("--prism-string", "#2f6b3a") },
      },
      {
        types: ["number", "boolean"],
        style: { color: cssVar("--prism-number", "#8a5a20") },
      },
      {
        types: ["keyword", "atrule", "selector"],
        style: { color: cssVar("--prism-keyword", "#4f5bd5") },
      },
      {
        types: ["function", "class-name", "maybe-class-name"],
        style: { color: cssVar("--prism-function", "#4a5cc0") },
      },
      {
        types: ["operator", "entity", "url", "variable"],
        style: { color: cssVar("--prism-operator", "#5c6068") },
      },
      {
        types: ["punctuation", "tag", "builtin", "important"],
        style: { color: cssVar("--prism-punctuation", "#7a7f88") },
      },
      {
        types: ["property", "constant", "symbol", "deleted"],
        style: { color: cssVar("--prism-fg", "#3a3d45") },
      },
    ],
  };
}
