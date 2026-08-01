import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

type AppChromeContextValue = {
  sync: ReactNode | null;
  setSync: (node: ReactNode | null) => void;
};

const AppChromeContext = createContext<AppChromeContextValue | null>(null);

export function AppChromeProvider({ children }: { children: ReactNode }) {
  const [sync, setSync] = useState<ReactNode | null>(null);
  const value = useMemo(() => ({ sync, setSync }), [sync]);
  return <AppChromeContext.Provider value={value}>{children}</AppChromeContext.Provider>;
}

function useAppChromeContext(): AppChromeContextValue {
  const ctx = useContext(AppChromeContext);
  if (!ctx) throw new Error("AppChrome hooks must be used within AppChromeProvider");
  return ctx;
}

export function useAppChromeSync(): ReactNode | null {
  return useAppChromeContext().sync;
}

/** Register topbar sync slot; cleared on unmount or when `sync` changes to null. */
export function useRegisterAppChrome(sync: ReactNode | null) {
  const { setSync } = useAppChromeContext();
  useEffect(() => {
    setSync(sync);
    return () => setSync(null);
  }, [sync, setSync]);
}
