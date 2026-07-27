import { useEffect, useRef, useState, type MouseEvent } from "react";

function FilesIcon({ size = 12 }: { size?: number }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M15 2h-4a2 2 0 0 0-2 2v11a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V8" />
      <path d="M16.706 2.706A2.4 2.4 0 0 0 15 2v5a1 1 0 0 0 1 1h5a2.4 2.4 0 0 0-.706-1.706z" />
      <path d="M5 7a2 2 0 0 0-2 2v11a2 2 0 0 0 2 2h8a2 2 0 0 0 1.732-1" />
    </svg>
  );
}

type Props = {
  id: string;
  /** Hover/cursor only; click does not copy. */
  preview?: boolean;
};

export default function ContestIdCopy({ id, preview }: Props) {
  const [copied, setCopied] = useState(false);
  const timerRef = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (timerRef.current != null) window.clearTimeout(timerRef.current);
    };
  }, []);

  const onClick = (e: MouseEvent<HTMLButtonElement>) => {
    e.preventDefault();
    e.stopPropagation();
    if (preview) return;
    void (async () => {
      try {
        await navigator.clipboard.writeText(id);
      } catch {
        return;
      }
      setCopied(true);
      if (timerRef.current != null) window.clearTimeout(timerRef.current);
      timerRef.current = window.setTimeout(() => setCopied(false), 1000);
    })();
  };

  return (
    <button
      type="button"
      className="contest-card__id"
      onClick={onClick}
      tabIndex={preview ? -1 : undefined}
      title={preview ? undefined : "Скопировать ID"}
      aria-label={preview ? undefined : `Скопировать ID ${id}`}
    >
      <span className="contest-card__id-text">{id}</span>
      {!preview && copied ? (
        <span className="contest-card__id-icon" aria-hidden>
          <FilesIcon />
        </span>
      ) : null}
    </button>
  );
}
