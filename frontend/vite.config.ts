import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const apiUrl = env.VITE_API_URL || "http://localhost:8080";

  const isDev = mode === "development";

  const csp = [
    "default-src 'self'",
    isDev ? "script-src 'self' 'unsafe-inline'" : "script-src 'self'",
    "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
    "font-src 'self' https://fonts.gstatic.com data:",
    "img-src 'self' data: blob:",
    isDev
      ? `connect-src 'self' ${apiUrl} ws://localhost:* http://localhost:*`
      : `connect-src 'self' ${apiUrl}`,
    "base-uri 'self'",
    "form-action 'self'",
    "frame-ancestors 'none'",
    ...(isDev ? [] : ["upgrade-insecure-requests"]),
  ].join("; ");

  const securityHeaders = {
    "Content-Security-Policy": csp,
    "X-Content-Type-Options": "nosniff",
    "X-Frame-Options": "DENY",
    "X-XSS-Protection": "1; mode=block",
    "Referrer-Policy": "strict-origin-when-cross-origin",
    "Permissions-Policy": "camera=(), microphone=(), geolocation=()",
  };

  return {
    server: {
      port: 5173,
      open: true,
      headers: securityHeaders,
    },
    preview: {
      headers: securityHeaders,
    },
    plugins: [react(), tailwindcss(), tsconfigPaths()],
    build: {
      rollupOptions: {
        output: {
          manualChunks: {
            vendor: ["react", "react-dom", "react-router"],
            state: ["zustand", "@tanstack/react-query"],
            http: ["axios"],
          },
        },
      },
      target: "es2015",
      minify: "terser",
      sourcemap: false,
    },
  };
});
