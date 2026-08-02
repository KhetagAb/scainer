import { useState, type FormEvent } from "react";
import { clearToken, setToken } from "@/features/auth/authStorage";
import { useSession } from "@/features/auth/SessionProvider";
import { Logo } from "@/app/Logo";

type LoginResponse = { access_token: string; expires_in: number };
type ErrorBody = { error?: string };

export default function LoginPage() {
  const { setUsername, refreshSession } = useSession();
  const [username, setLoginUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      });
      const body = (await res.json()) as LoginResponse & ErrorBody;
      if (!res.ok) {
        setError(
          body.error === "invalid credentials"
            ? "Неверный логин или пароль"
            : body.error === "teacher not found"
              ? "Обратитесь к Хету"
              : body.error || "Ошибка входа"
        );
        return;
      }
      setToken(body.access_token);
      await refreshSession();
      setUsername(username);
      window.location.href = "/";
    } catch {
      clearToken();
      setError("Сервер недоступен");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-wrap">
      <div className="login-card">
        <Logo />
        <p className="meta">Вход для просмотра находок</p>
        <form onSubmit={onSubmit}>
          <label>
            Логин
            <input
              autoComplete="username"
              value={username}
              onChange={(e) => setLoginUsername(e.target.value)}
              required
            />
          </label>
          <label>
            Пароль
            <input
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </label>
          {error ? <div className="login-error">{error}</div> : null}
          <button className="btn btn--primary" type="submit" disabled={loading}>
            {loading ? "Вход…" : "Войти"}
          </button>
        </form>
      </div>
    </div>
  );
}
