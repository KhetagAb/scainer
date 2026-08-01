import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Link, useLocation } from "react-router-dom";
import type { ContestInfo } from "@/client/types.gen";
import {
  UNGROUPED_PARALLEL,
  compareContestId,
  compactContestName,
} from "@/features/contests/shared/contestHelpers";

type Props = {
  currentContestId: string;
  parallelId: string;
  contests: ContestInfo[];
  displayName: string;
};

type MenuPosition = {
  top: number;
  left: number;
  minWidth: number;
};

export function contestSwitchSuffix(
  pathname: string,
  contestId: string,
): "review" | "findings" {
  const base = `/contests/${encodeURIComponent(contestId)}`;
  if (pathname.startsWith(`${base}/findings`)) return "findings";
  return "review";
}

function canHover(): boolean {
  return window.matchMedia("(hover: hover)").matches;
}

export default function ContestCrumbSwitcher({
  currentContestId,
  parallelId,
  contests,
  displayName,
}: Props) {
  const { pathname } = useLocation();
  const [touchOpen, setTouchOpen] = useState(false);
  const [hoverOpen, setHoverOpen] = useState(false);
  const [menuPosition, setMenuPosition] = useState<MenuPosition | null>(null);
  const rootRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLUListElement>(null);
  const closeTimerRef = useRef<number | undefined>(undefined);

  const parallelContests = useMemo(
    () =>
      contests
        .filter((c) =>
          parallelId === UNGROUPED_PARALLEL ? !c.parallelId : c.parallelId === parallelId,
        )
        .sort((a, b) => compareContestId(a.id, b.id)),
    [contests, parallelId],
  );

  const section = contestSwitchSuffix(pathname, currentContestId);
  const menuOpen = touchOpen || hoverOpen;

  const updateMenuPosition = useCallback(() => {
    const el = rootRef.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    setMenuPosition({
      top: rect.bottom + 4,
      left: rect.left,
      minWidth: Math.max(rect.width, 160),
    });
  }, []);

  const openHoverMenu = useCallback(() => {
    if (!canHover()) return;
    window.clearTimeout(closeTimerRef.current);
    updateMenuPosition();
    setHoverOpen(true);
  }, [updateMenuPosition]);

  const scheduleHoverClose = useCallback(() => {
    if (!canHover()) return;
    window.clearTimeout(closeTimerRef.current);
    closeTimerRef.current = window.setTimeout(() => {
      setHoverOpen(false);
    }, 120);
  }, []);

  useLayoutEffect(() => {
    if (!menuOpen) return;
    updateMenuPosition();
    const onLayoutChange = () => updateMenuPosition();
    window.addEventListener("resize", onLayoutChange);
    window.addEventListener("scroll", onLayoutChange, true);
    return () => {
      window.removeEventListener("resize", onLayoutChange);
      window.removeEventListener("scroll", onLayoutChange, true);
    };
  }, [menuOpen, updateMenuPosition]);

  useEffect(() => {
    if (!touchOpen) return;
    const onDocClick = (e: MouseEvent) => {
      const target = e.target as Node;
      if (rootRef.current?.contains(target)) return;
      if (menuRef.current?.contains(target)) return;
      setTouchOpen(false);
    };
    const id = window.setTimeout(() => {
      document.addEventListener("mousedown", onDocClick);
    }, 0);
    return () => {
      window.clearTimeout(id);
      document.removeEventListener("mousedown", onDocClick);
    };
  }, [touchOpen]);

  useEffect(
    () => () => {
      window.clearTimeout(closeTimerRef.current);
    },
    [],
  );

  if (parallelContests.length <= 1) {
    return <span className="app-topbar__crumb-name">{displayName}</span>;
  }

  const onTriggerClick = () => {
    if (canHover()) return;
    updateMenuPosition();
    setTouchOpen((open) => !open);
  };

  const menu = menuOpen && menuPosition ? (
    <ul
      ref={menuRef}
      className="app-topbar__crumb-menu app-topbar__crumb-menu--portal"
      role="listbox"
      aria-label="Контесты параллели"
      style={{
        top: `${menuPosition.top}px`,
        left: `${menuPosition.left}px`,
        minWidth: `${menuPosition.minWidth}px`,
      }}
      onMouseEnter={openHoverMenu}
      onMouseLeave={scheduleHoverClose}
    >
      {parallelContests.map((c) => {
        const label = compactContestName(c.name || "") || c.id;
        const selected = c.id === currentContestId;
        const to = `/contests/${encodeURIComponent(c.id)}/${section}`;

        return (
          <li key={c.id} role="presentation">
            {selected ? (
              <span
                className="app-topbar__crumb-menu-option app-topbar__crumb-menu-option--selected"
                role="option"
                aria-selected
                aria-current="page"
              >
                {label}
              </span>
            ) : (
              <Link
                to={to}
                className="app-topbar__crumb-menu-option"
                role="option"
                aria-selected={false}
                onClick={() => {
                  setTouchOpen(false);
                  setHoverOpen(false);
                }}
              >
                {label}
              </Link>
            )}
          </li>
        );
      })}
    </ul>
  ) : null;

  return (
    <>
      <div
        ref={rootRef}
        className={[
          "app-topbar__crumb-menu-anchor",
          menuOpen ? "app-topbar__crumb-menu-anchor--open" : "",
        ]
          .filter(Boolean)
          .join(" ")}
        onMouseEnter={openHoverMenu}
        onMouseLeave={scheduleHoverClose}
      >
        <span
          className="app-topbar__crumb-name app-topbar__crumb-name--switcher"
          role="button"
          tabIndex={0}
          aria-haspopup="listbox"
          aria-expanded={menuOpen}
          aria-label="Переключить контест"
          onClick={onTriggerClick}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              onTriggerClick();
            }
          }}
        >
          <span className="app-topbar__crumb-name-label">{displayName}</span>
          <span className="app-topbar__crumb-chevron" aria-hidden />
        </span>
      </div>
      {menu ? createPortal(menu, document.body) : null}
    </>
  );
}
