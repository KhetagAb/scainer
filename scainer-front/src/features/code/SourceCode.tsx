import { useRef } from "react";
import { Highlight } from "prism-react-renderer";
import { toPrismLang } from "@/features/code/codeLang";
import { usePrismTheme } from "@/app/theme/usePrismTheme";
import CodeCiteButton from "@/features/code/CodeCiteButton";
import {
  getSequentialLineNo,
  onLineContentDoubleClick,
  onLineContentMouseDown,
  useCodeCiteInteractions,
} from "@/features/code/codeCiteInteractions";

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
  const prismTheme = usePrismTheme();
  const rootClass = ["code-lines", className].filter(Boolean).join(" ");
  const wrapRef = useRef<HTMLDivElement>(null);

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
    getLineNoFromLineEl: getSequentialLineNo,
    resetKey: code,
  });

  return (
    <div ref={wrapRef} className="code-lines-wrap">
      <Highlight theme={prismTheme} code={code} language={language}>
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
                      ? (e) =>
                          onLineContentMouseDown(
                            e,
                            ".line-no",
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
                            ".line-no",
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
        <CodeCiteButton cite={cite} onApply={applyCiteRange} />
      ) : null}
    </div>
  );
}
