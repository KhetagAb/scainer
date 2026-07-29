import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider, QueryCache, MutationCache } from "@tanstack/react-query";
import App from "@/app/App";
import { client } from "@/client/client.gen";
import { ApiError, normalizeApiError } from "@/lib/apiError";
import { SessionProvider } from "@/features/auth/SessionProvider";
import { ThemeProvider } from "@/app/theme/ThemeProvider";
import { clearToken } from "@/features/auth/authStorage";
import "@/app/styles.css";

// react-query mutations используют throwOnError: true — нормализуем в ApiError (_Error + status).
client.interceptors.error.use((error, response) => normalizeApiError(error, response.status));

const handleUnauthorized = (error: unknown) => {
  if (error instanceof ApiError && error.status === 401) {
    clearToken();
    // Перенаправление на /login будет обработано через auth guard в App
  }
};

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false, refetchOnWindowFocus: false },
  },
  queryCache: new QueryCache({
    onError: handleUnauthorized,
  }),
  mutationCache: new MutationCache({
    onError: handleUnauthorized,
  }),
});

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <SessionProvider>
        <ThemeProvider>
          <App />
        </ThemeProvider>
      </SessionProvider>
    </QueryClientProvider>
  </StrictMode>
);
