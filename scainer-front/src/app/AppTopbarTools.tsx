import { useLocation } from "react-router-dom";
import ThemeSegmentControl from "@/app/theme/ThemeSegmentControl";
import { useAppChromeSync } from "@/features/contests/shared/AppChromeContext";
import { useSensitivity } from "@/features/contests/shared/SensitivityContext";
import SensitivitySlider from "@/features/contests/ui/SensitivitySlider";

function isContestChromeRoute(pathname: string): boolean {
  return /^\/(parallels|contests)\//.test(pathname);
}

export default function AppTopbarTools() {
  const { pathname } = useLocation();
  const showContestTools = isContestChromeRoute(pathname);
  const sync = useAppChromeSync();
  const { threshold, setThreshold } = useSensitivity();

  return (
    <div className="app-topbar__tools app-topbar__tools--contest">
      <div className="app-topbar__right-rail">
        {showContestTools && sync ? (
          <div className="app-topbar__sync-slot">{sync}</div>
        ) : null}
        {showContestTools ? (
          <SensitivitySlider threshold={threshold} onChange={setThreshold} />
        ) : null}
        <ThemeSegmentControl />
      </div>
    </div>
  );
}
