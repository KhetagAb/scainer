import { useState, type ReactNode } from "react";

type Props = {
  children: ReactNode;
  className?: string;
  participant?: ReactNode;
  participantClassName?: string;
};

export function CodePaneMoreBar({
  children,
  className,
  participant,
  participantClassName,
}: Props) {
  const [open, setOpen] = useState(false);

  return (
    <div className={className ? `code-pane-bar ${className}` : "code-pane-bar"}>
      {participant != null ? (
        <span className={participantClassName ?? "code-pane-participant"}>{participant}</span>
      ) : null}
      <button
        type="button"
        className="code-pane-toggle"
        aria-expanded={open}
        onClick={() => setOpen((value) => !value)}
      >
        {open ? "свернуть" : "подробнее"}
      </button>
      {open ? <div className="code-pane-meta">{children}</div> : null}
    </div>
  );
}
