import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// 开发期把 /api 代理到 BFF，避免跨域；生产由 nginx 同源反代
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8002",
        changeOrigin: true,
      },
    },
  },
});
