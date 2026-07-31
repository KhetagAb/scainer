import { Outlet, useLocation } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { getContestsQueryKey } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import ImportContestsPanel from "@/features/contests/ui/ImportContestsPanel";
import { SensitivityProvider } from "@/features/contests/shared/SensitivityContext";
import AppHeader from "@/app/AppHeader";
import AiGradientDefs from "@/app/AiGradientDefs";

export default function RootLayout() {
  const location = useLocation();
  const queryClient = useQueryClient();
  const showImport = location.pathname === "/";

  return (
    <SensitivityProvider>
      <AiGradientDefs />
      <AppHeader>
        <Outlet />
      </AppHeader>
      {showImport ? (
        <div className="import-dock">
          <ImportContestsPanel
            onSuccess={() => {
              void queryClient.invalidateQueries({
                queryKey: getContestsQueryKey({ headers: authHeaders() }),
              });
            }}
          />
        </div>
      ) : null}
    </SensitivityProvider>
  );
}
