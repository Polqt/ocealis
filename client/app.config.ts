import { defineConfig } from "@solidjs/start/config";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  vite: {
    plugins: [tailwindcss()],
    server: {
      // Do NOT proxy "/api" — SolidStart/Vinxi owns that path and returns
      // unrelated 400s (e.g. "Missing X-Agent-ID header").
      proxy: {
        "/backend": {
          target: "http://127.0.0.1:8080",
          changeOrigin: true,
          rewrite: (path: string) => path.replace(/^\/backend/, "")
        },
        "/backend-ws": {
          target: "ws://127.0.0.1:8080",
          ws: true,
          changeOrigin: true,
          rewrite: (path: string) => path.replace(/^\/backend-ws/, "/ws")
        }
      }
    }
  }
});
