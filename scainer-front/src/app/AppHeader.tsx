import type { ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";
import AiGradientDefs from "@/app/AiGradientDefs";
import AppTopbarCrumb from "@/app/AppTopbarCrumb";
import AppTopbarTools from "@/app/AppTopbarTools";
import { Logo } from "@/features/findings/FindingCard";

type Props = {
  actions?: ReactNode;
  children: ReactNode;
};

function isContestChromeRoute(pathname: string): boolean {
  return /^\/(parallels|contests)\//.test(pathname);
}

export default function AppHeader({ actions, children }: Props) {
  const { pathname } = useLocation();
  const shellClass =
    "app-shell" + (isContestChromeRoute(pathname) ? " app-shell--contest-chrome" : "");

  return (
    <div className={shellClass}>
      <AiGradientDefs />
      <header className="app-topbar app-topbar--workspace-grid" data-app-topbar>
        <div className="app-topbar__lead">
          <Link to="/" className="logo-link" aria-label="На главную">
            <Logo />
          </Link>
          <AppTopbarCrumb />
        </div>
        <div className="app-topbar__toolbar">
          {actions ? <div className="header-actions">{actions}</div> : null}
          <AppTopbarTools />
        </div>
      </header>
      <main className="app-main">{children}</main>
    </div>
  );
}
