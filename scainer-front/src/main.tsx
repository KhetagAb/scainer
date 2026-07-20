import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider, QueryCache, MutationCache } from "@tanstack/react-query";
import App from "@/app/App";
import { client } from "@/client/client.gen";
import { SessionProvider } from "@/features/auth/SessionProvider";
import { clearToken } from "@/features/auth/authStorage";
import "@/app/styles.css";

// Сгенерированный клиент (throwOnError: true) бросает только тело ответа ({error: "..."}),
// без HTTP-статуса — прикладываем его сами, иначе все проверки `.status === 401/400` по
// приложению никогда не срабатывают.
client.interceptors.error.use((error, response) => {
  if (error && typeof error === "object") {
    return { ...(error as object), status: response.status };
  }
  return { error, status: response.status };
});

const handleUnauthorized = (error: any) => {
  if (error?.status === 401) {
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
        <App />
      </SessionProvider>
    </QueryClientProvider>
  </StrictMode>
);
