import path from "node:path";
import type { ProxyOptions } from "vite";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import type { ClientRequest, IncomingMessage, ServerResponse } from "node:http";

const apiTarget = process.env.VITE_API_PROXY_TARGET ?? "http://127.0.0.1:8080";

/** Час — SSE job progress может идти долго, сервер шлёт heartbeat раз в 15с. */
const SSE_PROXY_TIMEOUT_MS = 60 * 60 * 1000;

function configureSSEProxy(proxy: {
  on(event: "proxyReq", listener: (proxyReq: ClientRequest) => void): void;
  on(
    event: "proxyRes",
    listener: (proxyRes: IncomingMessage, req: IncomingMessage, res: ServerResponse) => void,
  ): void;
}): void {
  proxy.on("proxyReq", (proxyReq) => {
    proxyReq.setHeader("Accept-Encoding", "identity");
  });
  proxy.on("proxyRes", (proxyRes) => {
    const ct = proxyRes.headers["content-type"];
    if (typeof ct === "string" && ct.includes("text/event-stream")) {
      proxyRes.headers["cache-control"] = "no-cache, no-transform";
      proxyRes.headers["connection"] = "keep-alive";
      proxyRes.headers["x-accel-buffering"] = "no";
      delete proxyRes.headers["content-length"];
      delete proxyRes.headers["content-encoding"];
    }
  });
}

const sseProxy: ProxyOptions = {
  target: apiTarget,
  changeOrigin: true,
  timeout: SSE_PROXY_TIMEOUT_MS,
  proxyTimeout: SSE_PROXY_TIMEOUT_MS,
  configure: configureSSEProxy,
};

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
      "/api/jobs": sseProxy,
      "/api": { target: apiTarget, changeOrigin: true },
    },
  },
});
