import { useMemo, useRef } from "react";
import { diffLines } from "diff";
import { Highlight, type Language } from "prism-react-renderer";
import { toPrismLang } from "@/features/code/codeLang";
import { usePrismTheme } from "@/app/theme/usePrismTheme";
import CodeCiteButton from "@/features/code/CodeCiteButton";
import {
  getCiteLineNoFromData,
  onLineContentDoubleClick,
  onLineContentMouseDown,
  useCodeCiteInteractions,
} from "@/features/code/codeCiteInteractions";

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
  const prismTheme = usePrismTheme();

  return (
    <Highlight theme={prismTheme} code={code} language={language}>
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

const DIFF_GUTTER_SKIP = ".line-no, .unified-diff__gutter";

type Props = {
  oldLines: string[];
  newLines: string[];
  lang?: string;
  className?: string;
  onLineNumberClick?: (lineNo: number) => void;
  onCiteLines?: (from: number, to: number) => void;
};

export default function UnifiedDiffView({
  oldLines,
  newLines,
  lang,
  className,
  onLineNumberClick,
  onCiteLines,
}: Props) {
  const rows = useMemo(() => buildDiffRows(oldLines, newLines), [oldLines, newLines]);
  const language = toPrismLang(lang);
  const rootClass = ["code-lines", "unified-diff", className].filter(Boolean).join(" ");
  const wrapRef = useRef<HTMLDivElement>(null);
  const resetKey = useMemo(() => rows.map((r) => `${r.kind}:${r.newLine}:${r.text}`).join("\n"), [rows]);

  const {
    cite,
    setCite,
    lineFlashes,
    skipNextCiteMouseUpRef,
    handleLineNumberClick,
    triggerLineFlash,
    applyCiteRange,
    onPreMouseDown,
    onPreMouseUp,
  } = useCodeCiteInteractions({
    wrapRef,
    onCiteLines,
    onLineNumberClick,
    getLineNoFromLineEl: getCiteLineNoFromData,
    resetKey,
  });

  return (
    <div ref={wrapRef} className="code-lines-wrap">
      <pre
        className={rootClass}
        onMouseDown={onCiteLines ? onPreMouseDown : undefined}
        onMouseUp={onCiteLines ? onPreMouseUp : undefined}
      >
        {rows.map((row, i) => {
          const flash = lineFlashes[i] ? " line-cite-flash" : "";
          return (
            <code
              key={i}
              className={`line unified-diff__line unified-diff__line--${row.kind}${flash}`}
              {...(row.newLine != null ? { "data-cite-line": row.newLine } : {})}
              onMouseDown={
                onLineNumberClick
                  ? (e) =>
                      onLineContentMouseDown(
                        e,
                        DIFF_GUTTER_SKIP,
                        onLineNumberClick,
                        skipNextCiteMouseUpRef,
                      )
                  : undefined
              }
              onDoubleClick={
                onLineNumberClick
                  ? (e) =>
                      onLineContentDoubleClick(
                        e,
                        DIFF_GUTTER_SKIP,
                        row.newLine,
                        i,
                        onLineNumberClick,
                        () => setCite(null),
                        triggerLineFlash,
                      )
                  : undefined
              }
            >
              <span className="unified-diff__gutter unified-diff__gutter--old" aria-hidden>
                {row.oldLine ?? ""}
              </span>
              {onLineNumberClick && row.newLine != null ? (
                <button
                  type="button"
                  className="unified-diff__gutter unified-diff__gutter--new line-no--clickable"
                  onClick={() => handleLineNumberClick(row.newLine!, i)}
                  aria-label={`Строка ${row.newLine}`}
                >
                  {row.newLine}
                </button>
              ) : (
                <span className="unified-diff__gutter unified-diff__gutter--new" aria-hidden>
                  {row.newLine ?? ""}
                </span>
              )}
              <span
                className={`unified-diff__sign unified-diff__sign--${row.kind}`}
                aria-hidden
              >
                {row.kind === "added" ? "+" : row.kind === "removed" ? "−" : ""}
              </span>
              <DiffLineBody text={row.text} language={language} />
            </code>
          );
        })}
      </pre>
      {onCiteLines && cite ? (
        <CodeCiteButton
          cite={cite}
          className="code-cite-btn--unified-diff"
          onApply={applyCiteRange}
        />
      ) : null}
    </div>
  );
}
