import { fileURLToPath } from "node:url";
import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const target = env.BACKEND_URL || "http://127.0.0.1:8080";
  const proxy = {
    "/api": { target, changeOrigin: true },
    "/health": { target, changeOrigin: true },
  };
  return {
    plugins: [react()],
    define: { __BACKEND_TARGET__: JSON.stringify(target) },
    server: {
      host: "127.0.0.1",
      proxy,
      fs: { allow: [fileURLToPath(new URL("..", import.meta.url))] },
    },
    preview: { proxy },
  };
});
