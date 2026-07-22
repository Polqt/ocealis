import { defineConfig } from "@solidjs/start/config";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  vite: {
    plugins: [tailwindcss()],
    server: {
      // Same-origin in dev — avoids browser CORS against :8080 entirely.
      proxy: {
        "/api": {
          target: "http://127.0.0.1:8080",
          changeOrigin: true
        },
        "/ws": {
          target: "ws://127.0.0.1:8080",
          ws: true,
          changeOrigin: true
        }
      }
    }
  }
});
