import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { Logo } from "@/features/findings/FindingCard";

type Props = {
  meta?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
};

export default function AppHeader({ meta, actions, children }: Props) {
  return (
    <div className="app-shell">
      <header className="header">
        <Link to="/" className="logo-link" aria-label="На главную">
          <Logo />
          <span className="logo-beta" aria-hidden />
        </Link>
        {typeof meta === "string" ? <p className="meta">{meta}</p> : meta}
        {actions ? <div className="header-actions">{actions}</div> : null}
      </header>
      {children}
    </div>
  );
}
