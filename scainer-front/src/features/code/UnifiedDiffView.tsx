import { useMemo } from "react";
import { diffLines } from "diff";
import { Highlight, themes, type Language } from "prism-react-renderer";
import { toPrismLang } from "@/features/code/codeLang";

type DiffRow = {
  kind: "unchanged" | "added" | "removed";
  text: string;
  oldLine: number | null;
  newLine: number | null;
};

function buildDiffRows(oldLines: string[], newLines: string[]): DiffRow[] {
  const oldText = oldLines.join("\n");
  const newText = newLines.join("\n");
  const parts = diffLines(oldText, newText);
  const rows: DiffRow[] = [];
  let oldLine = 1;
  let newLine = 1;

  for (const part of parts) {
    const chunkLines = part.value.split("\n");
    if (chunkLines.length > 0 && chunkLines[chunkLines.length - 1] === "") {
      chunkLines.pop();
    }

    for (const line of chunkLines) {
      if (part.added) {
        rows.push({ kind: "added", text: line, oldLine: null, newLine: newLine++ });
      } else if (part.removed) {
        rows.push({ kind: "removed", text: line, oldLine: oldLine++, newLine: null });
      } else {
        rows.push({ kind: "unchanged", text: line, oldLine: oldLine++, newLine: newLine++ });
      }
    }
  }

  return rows;
}

function DiffLineBody({ text, language }: { text: string; language: Language }) {
  const code = text.length ? text : " ";

  return (
    <Highlight theme={themes.github} code={code} language={language}>
      {({ tokens, getTokenProps }) => (
        <span className="line-body">
          {(tokens[0] ?? []).map((token, j) => (
            <span key={j} {...getTokenProps({ token })} />
          ))}
        </span>
      )}
    </Highlight>
  );
}

type Props = {
  oldLines: string[];
  newLines: string[];
  lang?: string;
  className?: string;
};

export default function UnifiedDiffView({ oldLines, newLines, lang, className }: Props) {
  const rows = useMemo(() => buildDiffRows(oldLines, newLines), [oldLines, newLines]);
  const language = toPrismLang(lang);
  const rootClass = ["code-lines", "unified-diff", className].filter(Boolean).join(" ");

  return (
    <div className="code-lines-wrap">
      <pre className={rootClass}>
        {rows.map((row, i) => (
          <code
            key={i}
            className={`line unified-diff__line unified-diff__line--${row.kind}`}
          >
            <span className="unified-diff__gutter unified-diff__gutter--old" aria-hidden>
              {row.oldLine ?? ""}
            </span>
            <span className="unified-diff__gutter unified-diff__gutter--new" aria-hidden>
              {row.newLine ?? ""}
            </span>
            <DiffLineBody text={row.text} language={language} />
          </code>
        ))}
      </pre>
    </div>
  );
}
