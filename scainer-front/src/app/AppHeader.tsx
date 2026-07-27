import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import AppTopbarCrumb from "@/app/AppTopbarCrumb";
import { Logo } from "@/features/findings/FindingCard";

type Props = {
  actions?: ReactNode;
  children: ReactNode;
};

export default function AppHeader({ actions, children }: Props) {
  return (
    <div className="app-shell">
      <header className="app-topbar">
        <Link to="/" className="logo-link" aria-label="На главную">
          <Logo />
        </Link>
        <AppTopbarCrumb />
        <div className="app-topbar__toolbar">
          {actions ? <div className="header-actions">{actions}</div> : null}
        </div>
      </header>
      <main className="app-main">{children}</main>
    </div>
  );
}
