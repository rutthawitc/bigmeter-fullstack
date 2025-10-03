import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react-swc";
import tailwind from "@tailwindcss/vite";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const api = env.VITE_API_BASE_URL || "http://localhost:8089";
  const loginEndpoint =
    env.VITE_LOGIN_API ||
    "https://intranet.pwa.co.th/login/webservice_login6.php";

  let loginUrl: URL | null = null;
  try {
    loginUrl = new URL(loginEndpoint);
  } catch (error) {
    // Only warn during development, this is a build-time config
    if (mode === "development") {
      console.warn(
        "[vite] invalid VITE_LOGIN_API value, falling back to default proxy path",
        error,
      );
    }
  }

  return {
    plugins: [react(), tailwind()],
    server: {
      port: 5173,
      open: true,
      proxy: {
        "/api": {
          target: api,
          changeOrigin: true,
        },
        ...(loginUrl
          ? {
              "/auth/login": {
                target: `${loginUrl.protocol}//${loginUrl.host}`,
                changeOrigin: true,
                secure: false,
                rewrite: () => loginUrl!.pathname,
              },
            }
          : {}),
      },
    },
  };
});
