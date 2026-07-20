import { Outlet, useLocation } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { getContestsQueryKey } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import ImportContestsPanel from "@/features/contests/ImportContestsPanel";
import SensitivitySlider from "@/features/contests/SensitivitySlider";
import {
  SensitivityProvider,
  useSensitivity,
} from "@/features/contests/SensitivityContext";
import FindingsGroupTabs from "@/features/findings/FindingsGroupTabs";
import AppHeader from "@/app/AppHeader";

function HeaderActions() {
  const location = useLocation();
  const queryClient = useQueryClient();
  const {
    scopeParallelId,
    onContestPage,
    threshold,
    setThreshold,
    groupBy,
    setGroupBy,
  } = useSensitivity();
  const showImport = location.pathname === "/";

  if (showImport) {
    return (
      <ImportContestsPanel
        onSuccess={() => {
          void queryClient.invalidateQueries({
            queryKey: getContestsQueryKey({ headers: authHeaders() }),
          });
        }}
      />
    );
  }

  if (scopeParallelId == null) return null;

  return (
    <div className="header-findings-controls">
      {onContestPage ? (
        <FindingsGroupTabs groupBy={groupBy} onChange={setGroupBy} />
      ) : null}
      <SensitivitySlider
        className="header-sensitivity"
        threshold={threshold}
        onChange={setThreshold}
      />
    </div>
  );
}

export default function RootLayout() {
  return (
    <SensitivityProvider>
      <AppHeader actions={<HeaderActions />}>
        <Outlet />
      </AppHeader>
    </SensitivityProvider>
  );
}
