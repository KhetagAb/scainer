import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { authHeaders, clearToken, getToken } from "@/features/auth/authStorage";

type SessionValue = {
  username: string | null;
  isLoading: boolean;
  setUsername: (username: string | null) => void;
  logout: () => void;
  refreshSession: () => Promise<void>;
};

const SessionContext = createContext<SessionValue | null>(null);

async function fetchMe(): Promise<{ ok: true; username: string } | { ok: false }> {
  const res = await fetch("/api/auth/me", { headers: authHeaders() });
  if (!res.ok) return { ok: false };
  const body = (await res.json()) as { username?: string };
  if (!body.username) return { ok: false };
  return { ok: true, username: body.username };
}

export function SessionProvider({ children }: { children: ReactNode }) {
  const [username, setUsername] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const refreshSession = useCallback(async () => {
    if (!getToken()) {
      setUsername(null);
      setIsLoading(false);
      return;
    }
    setIsLoading(true);
    try {
      const result = await fetchMe();
      if (result.ok) {
        setUsername(result.username);
      } else {
        clearToken();
        setUsername(null);
      }
    } catch {
      clearToken();
      setUsername(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void refreshSession();
  }, [refreshSession]);

  const logout = useCallback(() => {
    clearToken();
    setUsername(null);
    window.location.href = "/login";
  }, []);

  const value = useMemo(
    () => ({ username, isLoading, setUsername, logout, refreshSession }),
    [username, isLoading, logout, refreshSession],
  );

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession() {
  const ctx = useContext(SessionContext);
  if (!ctx) throw new Error("useSession must be used within SessionProvider");
  return ctx;
}
