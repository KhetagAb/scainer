import { useCallback, useEffect, useRef, useState, type MouseEvent } from "react";
import { Highlight, themes } from "prism-react-renderer";
import { toPrismLang } from "@/features/code/codeLang";

const LINE_FLASH_MS = 1250;

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

function onLineContentMouseDown(
  e: MouseEvent,
  onLineNumberClick?: (lineNo: number) => void,
  skipNextCiteMouseUpRef?: { current: boolean },
): void {
  if (!onLineNumberClick) return;
  if ((e.target as HTMLElement).closest(".line-no")) return;
  if (e.detail === 2) {
    e.preventDefault();
    if (skipNextCiteMouseUpRef) skipNextCiteMouseUpRef.current = true;
  }
}

function onLineContentDoubleClick(
  e: MouseEvent,
  lineNo: number,
  lineIndex: number,
  onLineNumberClick: (lineNo: number) => void,
  clearCite: () => void,
  triggerFlash: (lineIndex: number) => void,
): void {
  if ((e.target as HTMLElement).closest(".line-no")) return;
  e.preventDefault();
  onLineNumberClick(lineNo);
  triggerFlash(lineIndex);
  clearCite();
  window.getSelection()?.removeAllRanges();
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
  const citeModifierSelectRef = useRef(false);
  const modifierHeldRef = useRef(false);
  const skipNextCiteMouseUpRef = useRef(false);
  const [cite, setCite] = useState<{ from: number; to: number; top: number } | null>(null);
  const [lineFlashes, setLineFlashes] = useState<Record<number, number>>({});

  const triggerLineFlashMany = useCallback((lineIndices: number[]) => {
    if (lineIndices.length === 0) return;
    const token = Date.now();
    setLineFlashes((prev) => {
      const next = { ...prev };
      for (const idx of lineIndices) delete next[idx];
      return next;
    });
    requestAnimationFrame(() => {
      setLineFlashes((prev) => {
        const next = { ...prev };
        for (const idx of lineIndices) next[idx] = token;
        return next;
      });
      window.setTimeout(() => {
        setLineFlashes((prev) => {
          const next = { ...prev };
          for (const idx of lineIndices) {
            if (prev[idx] === token) delete next[idx];
          }
          return next;
        });
      }, LINE_FLASH_MS);
    });
  }, []);

  const triggerLineFlash = useCallback(
    (lineIndex: number) => triggerLineFlashMany([lineIndex]),
    [triggerLineFlashMany],
  );

  const triggerLineFlashRange = useCallback(
    (fromLine: number, toLine: number) => {
      const from = Math.min(fromLine, toLine);
      const to = Math.max(fromLine, toLine);
      const indices: number[] = [];
      for (let n = from; n <= to; n++) indices.push(n - 1);
      triggerLineFlashMany(indices);
    },
    [triggerLineFlashMany],
  );

  const handleLineNumberClick = useCallback(
    (lineNo: number, lineIndex: number) => {
      if (!onLineNumberClick) return;
      onLineNumberClick(lineNo);
      triggerLineFlash(lineIndex);
    },
    [onLineNumberClick, triggerLineFlash],
  );

  const applyCiteRange = useCallback(
    (from: number, to: number) => {
      if (!onCiteLines) return;
      onCiteLines(from, to);
      triggerLineFlashRange(from, to);
      setCite(null);
      window.getSelection()?.removeAllRanges();
    },
    [onCiteLines, triggerLineFlashRange],
  );

  const refreshCite = useCallback(() => {
    if (!onCiteLines) return;
    if (modifierHeldRef.current || citeModifierSelectRef.current) {
      setCite(null);
      return;
    }
    const wrap = wrapRef.current;
    const root = wrap?.querySelector<HTMLElement>(".code-lines");
    if (!wrap || !root) {
      setCite(null);
      return;
    }
    setCite(lineRangeFromSelection(root, wrap));
  }, [onCiteLines]);

  const onPreMouseDown = useCallback(
    (e: MouseEvent) => {
      if (!onCiteLines) return;
      citeModifierSelectRef.current = e.metaKey || e.ctrlKey;
    },
    [onCiteLines],
  );

  const onPreMouseUp = useCallback(
    (e: MouseEvent) => {
      if (!onCiteLines) return;

      if (skipNextCiteMouseUpRef.current) {
        skipNextCiteMouseUpRef.current = false;
        refreshCite();
        return;
      }

      const modifier = e.metaKey || e.ctrlKey || citeModifierSelectRef.current;
      citeModifierSelectRef.current = false;

      const wrap = wrapRef.current;
      const root = wrap?.querySelector<HTMLElement>(".code-lines");
      if (!wrap || !root) return;

      const range = lineRangeFromSelection(root, wrap);
      if (modifier && range) {
        applyCiteRange(range.from, range.to);
        return;
      }

      refreshCite();
    },
    [onCiteLines, applyCiteRange, refreshCite],
  );

  useEffect(() => {
    if (!onCiteLines) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Meta" || e.key === "Control") modifierHeldRef.current = true;
    };
    const onKeyUp = (e: KeyboardEvent) => {
      if (e.key === "Meta" || e.key === "Control") modifierHeldRef.current = false;
    };
    const onSel = () => refreshCite();
    document.addEventListener("keydown", onKeyDown);
    document.addEventListener("keyup", onKeyUp);
    document.addEventListener("selectionchange", onSel);
    return () => {
      document.removeEventListener("keydown", onKeyDown);
      document.removeEventListener("keyup", onKeyUp);
      document.removeEventListener("selectionchange", onSel);
    };
  }, [onCiteLines, refreshCite]);

  useEffect(() => {
    setCite(null);
    setLineFlashes({});
  }, [code]);

  return (
    <div ref={wrapRef} className="code-lines-wrap">
      <Highlight theme={themes.github} code={code} language={language}>
        {({ tokens, getLineProps, getTokenProps }) => (
          <pre
            className={rootClass}
            onMouseDown={onCiteLines ? onPreMouseDown : undefined}
            onMouseUp={onCiteLines ? onPreMouseUp : undefined}
          >
            {tokens.map((line, i) => {
              const lineNo = i + 1;
              const lineProps = getLineProps({ line });
              const match = matchLine?.(lineNo) ? " line-match" : "";
              const flash = lineFlashes[i] ? " line-cite-flash" : "";
              return (
                <code
                  key={i}
                  {...lineProps}
                  className={`line${match}${flash}${lineProps.className ? ` ${lineProps.className}` : ""}`}
                  onMouseDown={
                    onLineNumberClick
                      ? (e) => onLineContentMouseDown(e, onLineNumberClick, skipNextCiteMouseUpRef)
                      : undefined
                  }
                  onDoubleClick={
                    onLineNumberClick
                      ? (e) =>
                          onLineContentDoubleClick(
                            e,
                            lineNo,
                            i,
                            onLineNumberClick,
                            () => setCite(null),
                            triggerLineFlash,
                          )
                      : undefined
                  }
                >
                  {onLineNumberClick ? (
                    <button
                      type="button"
                      className="line-no line-no--clickable"
                      onClick={() => handleLineNumberClick(lineNo, i)}
                      aria-label={`Строка ${lineNo}`}
                    >
                      {lineNo}
                    </button>
                  ) : (
                    <span className="line-no" aria-hidden>
                      {lineNo}
                    </span>
                  )}
                  <span className="line-body">
                    {line.map((token, j) => (
                      <span key={j} {...getTokenProps({ token })} />
                    ))}
                    {line.length === 0 ? "\n" : null}
                  </span>
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
          onClick={() => applyCiteRange(cite.from, cite.to)}
        >
          <CiteIcon />
        </button>
      ) : null}
    </div>
  );
}
