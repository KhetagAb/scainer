import { useEffect } from "react";
import { RouterProvider, createBrowserRouter, Navigate } from "react-router-dom";
import { client } from "@/client/client.gen";
import { useSession } from "@/features/auth/SessionProvider";
import { authHeaders, clearToken } from "@/features/auth/authStorage";
import LoginPage from "@/features/auth/LoginPage";
import ParallelsPage from "@/features/contests/ParallelsPage";
import ParallelPage from "@/features/contests/ParallelPage";
import ContestPage from "@/features/contests/ContestPage";
import RootLayout from "@/app/RootLayout";
import NotFoundPage from "@/app/NotFoundPage";

function App() {
  const { username, isLoading, logout } = useSession();

  useEffect(() => {
    client.setConfig({
      baseUrl: "/",
      headers: authHeaders(),
    });
  }, [username]);

  const handleUnauthorized = () => {
    clearToken();
    logout();
  };

  if (isLoading) {
    return <div className="page-center">Загрузка…</div>;
  }

  // Auth guard: перенаправляем на /login если не авторизован
  if (!username) {
    const loginRouter = createBrowserRouter([
      { path: "/login", element: <LoginPage /> },
      { path: "*", element: <Navigate to="/login" replace /> },
    ]);
    return <RouterProvider router={loginRouter} />;
  }

  // Авторизованный маршрут
  const router = createBrowserRouter([
    {
      element: <RootLayout />,
      children: [
        { path: "/", element: <ParallelsPage onUnauthorized={handleUnauthorized} /> },
        {
          path: "/parallels/:id",
          element: <ParallelPage onUnauthorized={handleUnauthorized} />,
        },
        { path: "/contests/:id", element: <ContestPage onUnauthorized={handleUnauthorized} /> },
        { path: "*", element: <NotFoundPage /> },
      ],
    },
  ]);

  return <RouterProvider router={router} />;
}

export default App;
