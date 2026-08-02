import { useLocation } from "react-router-dom";
import { useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { getContestsQueryKey } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { useAppChromeSync } from "@/features/contests/shared/AppChromeContext";
import { useSensitivity } from "@/features/contests/shared/SensitivityContext";
import ImportContestsPanel from "@/features/contests/ui/ImportContestsPanel";
import SensitivitySlider from "@/features/contests/ui/SensitivitySlider";

function isContestChromeRoute(pathname: string): boolean {
  return /^\/(parallels|contests)\//.test(pathname);
}

export default function AppTopbarTools() {
  const { pathname } = useLocation();
  const queryClient = useQueryClient();
  const showContestTools = isContestChromeRoute(pathname);
  const showImport = pathname === "/";
  const sync = useAppChromeSync();
  const { threshold, setThreshold } = useSensitivity();
  const onImportSuccess = useCallback(() => {
    void queryClient.invalidateQueries({
      queryKey: getContestsQueryKey({ headers: authHeaders() }),
    });
  }, [queryClient]);

  return (
    <div className="app-topbar__tools app-topbar__tools--contest">
      <div className="app-topbar__right-rail">
        {showContestTools && sync ? (
          <div className="app-topbar__sync-slot">{sync}</div>
        ) : null}
        {showContestTools ? (
          <SensitivitySlider threshold={threshold} onChange={setThreshold} />
        ) : null}
        {showImport ? <ImportContestsPanel onSuccess={onImportSuccess} /> : null}
      </div>
    </div>
  );
}
