import { Outlet } from "react-router-dom";
import { AppChromeProvider } from "@/features/contests/shared/AppChromeContext";
import { SensitivityProvider } from "@/features/contests/shared/SensitivityContext";
import AppHeader from "@/app/AppHeader";

export default function RootLayout() {
  return (
    <SensitivityProvider>
      <AppChromeProvider>
        <AppHeader>
          <Outlet />
        </AppHeader>
      </AppChromeProvider>
    </SensitivityProvider>
  );
}
