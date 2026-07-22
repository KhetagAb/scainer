import { useCallback, useEffect, useRef, useState } from "react";
import { Highlight, themes, type Language } from "prism-react-renderer";

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

function toPrismLang(lang?: string): Language {
  if (!lang) return "clike";
  const key = lang.trim().toLowerCase();
  return PRISM_LANG[key] ?? "clike";
}

function CiteIcon() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" width="16" height="16" aria-hidden>
      <path
        fill="currentColor"
        fillRule="evenodd"
        d="M22 10a4 4 0 0 0-4-4h-8a4 4 0 0 0-4 4v8a4 4 0 0 0 4 4 1 1 0 1 0 0-2 2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2 1 1 0 1 0 2 0"
        clipRule="evenodd"
      />
      <path
        fill="currentColor"
        fillRule="evenodd"
        d="M3 16a1 1 0 0 1-1-1V9a7 7 0 0 1 7-7h6a1 1 0 1 1 0 2H9a5 5 0 0 0-5 5v6a1 1 0 0 1-1 1"
        clipRule="evenodd"
      />
      <path
        fill="currentColor"
        d="M13.929 13.929a1 1 0 0 1 1.414 0l3.536 3.535c.481.482.963 1.128 1.389 1.77h.203a1 1 0 0 0-.029-.182c-.227-.93-.439-2.033-.442-2.93v-1.113A1 1 0 0 1 22 15v6a1 1 0 0 1-1 1h-5.996A1 1 0 0 1 15 20h1.1c.902 0 2.014.213 2.952.442q.091.023.182.029v-.203c-.642-.426-1.288-.908-1.77-1.39l-3.535-3.535a1 1 0 0 1 0-1.414"
      />
    </svg>
  );
}

function lineRangeFromSelection(
  root: HTMLElement,
  wrap: HTMLElement,
): { from: number; to: number; top: number } | null {
  const sel = window.getSelection();
  if (!sel || sel.isCollapsed || sel.rangeCount === 0) return null;

  const range = sel.getRangeAt(0);
  if (!root.contains(range.commonAncestorContainer)) return null;

  const lineEls = root.querySelectorAll<HTMLElement>(".line");
  let from = -1;
  let to = -1;
  lineEls.forEach((el, i) => {
    if (range.intersectsNode(el)) {
      const n = i + 1;
      if (from < 0) from = n;
      to = n;
    }
  });
  if (from < 0 || to < 0) return null;

  const rect = range.getBoundingClientRect();
  if (rect.width === 0 && rect.height === 0) return null;
  const wrapRect = wrap.getBoundingClientRect();

  return {
    from,
    to,
    top: rect.bottom - wrapRect.top + 4,
  };
}

type Props = {
  lines: string[];
  lang?: string;
  className?: string;
  /** 1-based line numbers that should get .line-match */
  matchLine?: (lineNo: number) => boolean;
  onLineNumberClick?: (lineNo: number) => void;
  onCiteLines?: (from: number, to: number) => void;
};

export default function SourceCode({
  lines,
  lang,
  className,
  matchLine,
  onLineNumberClick,
  onCiteLines,
}: Props) {
  const code = lines.length ? lines.join("\n") : "(нет исходника)";
  const language = toPrismLang(lang);
  const rootClass = ["code-lines", className].filter(Boolean).join(" ");
  const wrapRef = useRef<HTMLDivElement>(null);
  const [cite, setCite] = useState<{ from: number; to: number; top: number } | null>(null);

  const refreshCite = useCallback(() => {
    if (!onCiteLines) return;
    const wrap = wrapRef.current;
    const root = wrap?.querySelector<HTMLElement>(".code-lines");
    if (!wrap || !root) {
      setCite(null);
      return;
    }
    setCite(lineRangeFromSelection(root, wrap));
  }, [onCiteLines]);

  useEffect(() => {
    if (!onCiteLines) return;
    const onSel = () => refreshCite();
    document.addEventListener("selectionchange", onSel);
    return () => document.removeEventListener("selectionchange", onSel);
  }, [onCiteLines, refreshCite]);

  useEffect(() => {
    setCite(null);
  }, [code]);

  return (
    <div ref={wrapRef} className="code-lines-wrap">
      <Highlight theme={themes.github} code={code} language={language}>
        {({ tokens, getLineProps, getTokenProps }) => (
          <pre className={rootClass} onMouseUp={onCiteLines ? refreshCite : undefined}>
            {tokens.map((line, i) => {
              const lineNo = i + 1;
              const lineProps = getLineProps({ line });
              const match = matchLine?.(lineNo) ? " line-match" : "";
              return (
                <code
                  key={i}
                  {...lineProps}
                  className={`line${match}${lineProps.className ? ` ${lineProps.className}` : ""}`}
                >
                  {onLineNumberClick ? (
                    <button
                      type="button"
                      className="line-no line-no--clickable"
                      onClick={() => onLineNumberClick(lineNo)}
                      aria-label={`Строка ${lineNo}`}
                    >
                      {lineNo}
                    </button>
                  ) : (
                    <span className="line-no" aria-hidden>
                      {lineNo}
                    </span>
                  )}
                  {line.map((token, j) => (
                    <span key={j} {...getTokenProps({ token })} />
                  ))}
                  {line.length === 0 ? "\n" : null}
                </code>
              );
            })}
          </pre>
        )}
      </Highlight>
      {onCiteLines && cite ? (
        <button
          type="button"
          className="code-cite-btn"
          style={{ top: cite.top }}
          title="Вставить номера строк в комментарий"
          aria-label={`Вставить строки ${cite.from}${cite.from === cite.to ? "" : `–${cite.to}`} в комментарий`}
          onMouseDown={(e) => e.preventDefault()}
          onClick={() => {
            onCiteLines(cite.from, cite.to);
            setCite(null);
            window.getSelection()?.removeAllRanges();
          }}
        >
          <CiteIcon />
        </button>
      ) : null}
    </div>
  );
}
