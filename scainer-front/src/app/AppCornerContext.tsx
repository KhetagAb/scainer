import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import ThemeFloatingControl from "@/app/theme/ThemeFloatingControl";

type AppCornerContextValue = {
  setExtra: (node: ReactNode | null) => void;
};

const AppCornerContext = createContext<AppCornerContextValue | null>(null);

export function AppCornerProvider({ children }: { children: ReactNode }) {
  const [extra, setExtra] = useState<ReactNode | null>(null);
  const value = useMemo(() => ({ setExtra }), []);

  return (
    <AppCornerContext.Provider value={value}>
      {children}
      <div className="app-corner-dock">
        {extra}
        <ThemeFloatingControl />
      </div>
    </AppCornerContext.Provider>
  );
}

function useAppCornerContext(): AppCornerContextValue {
  const ctx = useContext(AppCornerContext);
  if (!ctx) throw new Error("AppCorner hooks must be used within AppCornerProvider");
  return ctx;
}

/** Register extra bottom-right corner actions; cleared on unmount. */
export function useRegisterAppCorner(node: ReactNode | null) {
  const { setExtra } = useAppCornerContext();
  useEffect(() => {
    setExtra(node);
    return () => setExtra(null);
  }, [node, setExtra]);
}
