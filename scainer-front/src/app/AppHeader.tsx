import type { ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";
import AiGradientDefs from "@/app/AiGradientDefs";
import AppTopbarCrumb from "@/app/AppTopbarCrumb";
import AppTopbarTools from "@/app/AppTopbarTools";
import { Logo } from "@/app/Logo";

type Props = {
  children: ReactNode;
};

function isContestChromeRoute(pathname: string): boolean {
  return /^\/(parallels|contests)\//.test(pathname);
}

export default function AppHeader({ children }: Props) {
  const { pathname } = useLocation();
  const contestChrome = isContestChromeRoute(pathname);
  const shellClass = "app-shell" + (contestChrome ? " app-shell--contest-chrome" : "");
  const topbarClass =
    "app-topbar" + (contestChrome ? " app-topbar--workspace-grid" : "");

  return (
    <div className={shellClass}>
      <AiGradientDefs />
      <header className={topbarClass} data-app-topbar>
        <div className="app-topbar__lead">
          <Link to="/" className="app-logo-link" aria-label="На главную">
            <Logo />
          </Link>
          <AppTopbarCrumb />
        </div>
        <div className="app-topbar__toolbar">
          <AppTopbarTools />
        </div>
      </header>
      <main className="app-main">{children}</main>
    </div>
  );
}
