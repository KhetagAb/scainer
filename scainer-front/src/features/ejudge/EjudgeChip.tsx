import type { MouseEvent, ReactNode } from "react";

type Props = {
  children: ReactNode;
  className?: string;
  title?: string;
  href?: string;
  onClick?: (event: MouseEvent<HTMLAnchorElement>) => void;
};

export function EjudgeChip({ children, className, title, href, onClick }: Props) {
  const classes = className ? `chip ${className}` : "chip";
  if (href || onClick) {
    return (
      <a
        className={`${classes} chip--link`}
        href={href ?? "#"}
        title={title}
        target={onClick ? undefined : "_blank"}
        rel={onClick ? undefined : "noopener noreferrer"}
        onClick={
          onClick
            ? (event) => {
                event.preventDefault();
                onClick(event);
              }
            : undefined
        }
      >
        {children}
      </a>
    );
  }
  return (
    <span className={classes} title={title}>
      {children}
    </span>
  );
}
