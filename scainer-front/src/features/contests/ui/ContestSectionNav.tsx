import { NavLink, useLocation } from "react-router-dom";

export function ContestFindingsIcon({ size = 20 }: { size?: number }) {
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
      <path d="M19.07 4.93A10 10 0 0 0 6.99 3.34" />
      <path d="M4 6h.01" />
      <path d="M2.29 9.62A10 10 0 1 0 21.31 8.35" />
      <path d="M16.24 7.76A6 6 0 1 0 8.23 16.67" />
      <path d="M12 18h.01" />
      <path d="M17.99 11.66A6 6 0 0 1 15.77 16.67" />
      <circle cx="12" cy="12" r="2" />
      <path d="m13.41 10.59 5.66-5.66" />
    </svg>
  );
}

export function ContestReviewIcon({ size = 20 }: { size?: number }) {
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
      <path d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

/** Contest section icons in app-sidebar; only on findings / review. */
export default function ContestSectionNav() {
  const { pathname } = useLocation();
  const match = pathname.match(/^\/contests\/([^/]+)\/(findings|review)(\/|$)/);
  if (!match) return null;

  let contestId = match[1];
  try {
    contestId = decodeURIComponent(contestId);
  } catch {
    /* keep raw */
  }
  const base = `/contests/${encodeURIComponent(contestId)}`;

  return (
    <nav className="app-sidebar__nav" aria-label="Разделы контеста">
      <NavLink
        to={`${base}/review`}
        className={({ isActive }) =>
          `app-sidebar__link${isActive ? " is-active" : ""}`
        }
        aria-label="Ревью"
      >
        <ContestReviewIcon />
        <span className="app-sidebar__link-label">Ревью</span>
      </NavLink>
      <NavLink
        to={`${base}/findings`}
        end
        className={({ isActive }) =>
          `app-sidebar__link${isActive ? " is-active" : ""}`
        }
        aria-label="Детект"
      >
        <ContestFindingsIcon />
        <span className="app-sidebar__link-label">Детект</span>
      </NavLink>
    </nav>
  );
}
