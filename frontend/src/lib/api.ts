  import router from "@/router";
import { getToken, clearToken } from "@/lib/tokenStore";

// Full backend base URL in production (e.g. https://xxx.onrender.com), read
// from the VITE_API_URL env var at build time. In dev it's typically unset, so
// it falls back to "" and requests use relative paths that hit the Vite proxy
// (see vite.config.ts). Exported so non-`api` call sites (e.g. the auth flow in
// AuthPage.vue) resolve the backend the same way.
export const BASE_URL = import.meta.env.VITE_API_URL ?? "";

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const token = getToken();

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    clearToken();
    router.push("/auth");
    throw new Error("No autorizado");
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const data = await res.json();

  if (!res.ok) {
    throw new Error(data.error || `Error ${res.status}`);
  }

  return data as T;
}

const api = {
  get: <T = unknown>(path: string) => request<T>("GET", path),
  post: <T = unknown>(path: string, body?: unknown) => request<T>("POST", path, body),
  put: <T = unknown>(path: string, body?: unknown) => request<T>("PUT", path, body),
  delete: <T = unknown>(path: string) => request<T>("DELETE", path),
};

export default api;
