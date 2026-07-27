import { useEffect, useRef, useState } from "react";
import * as pdfjs from "pdfjs-dist";
import pdfWorker from "pdfjs-dist/build/pdf.worker.min.mjs?url";
import { resolveStatementPages } from "@/features/statements/statementPdfSections";

pdfjs.GlobalWorkerOptions.workerSrc = pdfWorker;

type Props = {
  data: ArrayBuffer;
  /** Буква задачи из ejudge (ProblemInfo.name), напр. «A». */
  problemLabel?: string | null;
};

function PdfPage({
  pdf,
  pageNumber,
  width,
}: {
  pdf: pdfjs.PDFDocumentProxy;
  pageNumber: number;
  width: number;
}) {
  const hostRef = useRef<HTMLDivElement>(null);
  const [rendered, setRendered] = useState(false);

  useEffect(() => {
    const host = hostRef.current;
    if (!host || rendered) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (!entries.some((e) => e.isIntersecting)) return;
        observer.disconnect();
        void (async () => {
          const page = await pdf.getPage(pageNumber);
          const baseViewport = page.getViewport({ scale: 1 });
          const cssScale = width / baseViewport.width;
          const pixelRatio = window.devicePixelRatio || 1;
          const viewport = page.getViewport({ scale: cssScale * pixelRatio });
          const canvas = document.createElement("canvas");
          canvas.width = Math.floor(viewport.width);
          canvas.height = Math.floor(viewport.height);
          canvas.style.width = `${Math.floor(viewport.width / pixelRatio)}px`;
          canvas.style.height = `${Math.floor(viewport.height / pixelRatio)}px`;
          canvas.className = "problem-statement-page__canvas";
          const ctx = canvas.getContext("2d");
          if (!ctx) return;
          await page.render({ canvasContext: ctx, viewport }).promise;
          host.replaceChildren(canvas);
          setRendered(true);
        })();
      },
      { rootMargin: "240px 0px" },
    );
    observer.observe(host);
    return () => observer.disconnect();
  }, [pdf, pageNumber, width, rendered]);

  return <div ref={hostRef} className="problem-statement-page" data-page={pageNumber} />;
}

export default function ProblemStatementViewer({ data, problemLabel }: Props) {
  const shellRef = useRef<HTMLDivElement>(null);
  const [pdf, setPdf] = useState<pdfjs.PDFDocumentProxy | null>(null);
  const [pageNumbers, setPageNumbers] = useState<number[] | null>(null);
  const [width, setWidth] = useState(0);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const task = pdfjs.getDocument({ data: data.slice(0) });
        const doc = await task.promise;
        if (cancelled) {
          void doc.destroy();
          return;
        }
        const pages = await resolveStatementPages(doc, problemLabel);
        if (cancelled) {
          void doc.destroy();
          return;
        }
        setPdf(doc);
        setPageNumbers(pages);
      } catch {
        if (!cancelled) setError("Не удалось открыть PDF");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [data, problemLabel]);

  useEffect(() => {
    const shell = shellRef.current;
    if (!shell) return;
    const update = () => setWidth(shell.clientWidth);
    update();
    const ro = new ResizeObserver(update);
    ro.observe(shell);
    return () => ro.disconnect();
  }, []);

  useEffect(() => {
    return () => {
      void pdf?.destroy();
    };
  }, [pdf]);

  if (error) {
    return <p className="problem-statement-viewer__error">{error}</p>;
  }

  const ready = pdf != null && pageNumbers != null && width > 0;

  return (
    <div ref={shellRef} className="problem-statement-viewer">
      {!ready ? <p className="problem-statement-viewer__loading">Открываем PDF…</p> : null}
      {ready
        ? pageNumbers.map((pageNumber) => (
            <PdfPage key={pageNumber} pdf={pdf} pageNumber={pageNumber} width={width} />
          ))
        : null}
    </div>
  );
}
