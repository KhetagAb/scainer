import type { CiteRangeState } from "@/features/code/codeCiteInteractions";

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

type Props = {
  cite: CiteRangeState;
  className?: string;
  onApply: (from: number, to: number) => void;
};

export default function CodeCiteButton({ cite, className, onApply }: Props) {
  return (
    <button
      type="button"
      className={["code-cite-btn", className].filter(Boolean).join(" ")}
      style={{ top: cite.top }}
      title="Вставить номера строк в комментарий"
      aria-label={`Вставить строки ${cite.from}${cite.from === cite.to ? "" : `–${cite.to}`} в комментарий`}
      onMouseDown={(e) => e.preventDefault()}
      onClick={() => onApply(cite.from, cite.to)}
    >
      <CiteIcon />
    </button>
  );
}
