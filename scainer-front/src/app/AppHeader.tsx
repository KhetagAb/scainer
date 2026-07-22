import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { Logo } from "@/features/findings/FindingCard";

type Props = {
  meta?: ReactNode;
  actions?: ReactNode;
  sidebarNav?: ReactNode;
  children: ReactNode;
};

export default function AppHeader({ meta, actions, sidebarNav, children }: Props) {
  const hasToolbar = Boolean(meta || actions);

  return (
    <div className="app-shell">
      <aside className="app-sidebar">
        <Link to="/" className="logo-link" aria-label="На главную">
          <Logo />
        </Link>
        {sidebarNav}
      </aside>
      <div className="app-main">
        {hasToolbar ? (
          <div className="app-toolbar">
            {typeof meta === "string" ? <p className="meta">{meta}</p> : meta}
            {actions ? <div className="header-actions">{actions}</div> : null}
          </div>
        ) : null}
        {children}
      </div>
    </div>
  );
}
