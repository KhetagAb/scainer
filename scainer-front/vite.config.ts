import path from "node:path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiTarget = process.env.VITE_API_PROXY_TARGET ?? "http://127.0.0.1:8080";

/** Час — SSE job может идти долго (analyze), сервер шлёт heartbeat раз в 15с. */
const SSE_PROXY_TIMEOUT_MS = 60 * 60 * 1000;

export default defineConfig({
  plugins: [react()],
  assetsInclude: ["**/*.lottie"],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  server: {
    host: true,
    port: 5173,
    strictPort: true,
    proxy: {
      // Более специфичный путь раньше /api — иначе SSE буферизуется общим прокси.
      "/api/jobs": {
        target: apiTarget,
        changeOrigin: true,
        timeout: SSE_PROXY_TIMEOUT_MS,
        proxyTimeout: SSE_PROXY_TIMEOUT_MS,
        configure: (proxy) => {
          proxy.on("proxyRes", (proxyRes) => {
            const ct = proxyRes.headers["content-type"];
            if (typeof ct === "string" && ct.includes("text/event-stream")) {
              proxyRes.headers["cache-control"] = "no-cache";
              proxyRes.headers["x-accel-buffering"] = "no";
              delete proxyRes.headers["content-length"];
            }
          });
        },
      },
      "/api": { target: apiTarget, changeOrigin: true },
    },
  },
});
