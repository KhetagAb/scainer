import { useCallback, useEffect, useMemo } from "react";
import { RouterProvider, createBrowserRouter, Navigate } from "react-router-dom";
import { client } from "@/client/client.gen";
import { useSession } from "@/features/auth/SessionProvider";
import { authHeaders, clearToken } from "@/features/auth/authStorage";
import LoginPage from "@/features/auth/LoginPage";
import ParallelsPage from "@/features/contests/ParallelsPage";
import ParallelPage from "@/features/contests/ParallelPage";
import ContestPage from "@/features/contests/ContestPage";
import {
  ContestFindingsRoute,
  ContestReviewRoute,
  ContestReviewSubmissionRoute,
} from "@/features/contests/ContestRoutes";
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

  const handleUnauthorized = useCallback(() => {
    clearToken();
    logout();
  }, [logout]);

  const loginRouter = useMemo(
    () =>
      createBrowserRouter([
        { path: "/login", element: <LoginPage /> },
        { path: "*", element: <Navigate to="/login" replace /> },
      ]),
    [],
  );

  const appRouter = useMemo(
    () =>
      createBrowserRouter([
        {
          element: <RootLayout />,
          children: [
            { path: "/", element: <ParallelsPage onUnauthorized={handleUnauthorized} /> },
            {
              path: "/parallels/:id",
              element: <ParallelPage onUnauthorized={handleUnauthorized} />,
            },
            {
              path: "/contests/:id",
              element: <ContestPage onUnauthorized={handleUnauthorized} />,
              children: [
                { path: "findings", element: <ContestFindingsRoute /> },
                { path: "review", element: <ContestReviewRoute /> },
                {
                  path: "review/:submissionId",
                  element: <ContestReviewSubmissionRoute />,
                },
              ],
            },
            { path: "*", element: <NotFoundPage /> },
          ],
        },
      ]),
    [handleUnauthorized],
  );

  if (isLoading) {
    return <div className="page-center">Загрузка…</div>;
  }

  if (!username) {
    return <RouterProvider router={loginRouter} />;
  }

  return <RouterProvider router={appRouter} />;
}

export default App;
