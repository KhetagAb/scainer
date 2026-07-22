import { Outlet, useLocation } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { getContestsQueryKey } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import ImportContestsPanel from "@/features/contests/ImportContestsPanel";
import { SensitivityProvider } from "@/features/contests/SensitivityContext";
import ContestSectionNav from "@/features/contests/ContestSectionNav";
import AppHeader from "@/app/AppHeader";

function HeaderActions() {
  const location = useLocation();
  const queryClient = useQueryClient();
  const showImport = location.pathname === "/";

  if (!showImport) return null;

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

export default function RootLayout() {
  return (
    <SensitivityProvider>
      <AppHeader actions={<HeaderActions />} sidebarNav={<ContestSectionNav />}>
        <Outlet />
      </AppHeader>
    </SensitivityProvider>
  );
}
