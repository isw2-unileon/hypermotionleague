// Best-effort, client-side JWT helpers. NOT a security check — the backend
// verifies the token signature on every request (see middleware/auth.go). The
// only goal here is to avoid trusting an obviously-expired token client-side
// (redirect to /auth without waiting for a 401, and never rehydrate a stale
// session from storage).
//
// Extracted so the router guard (router/index.ts) and the token rehydration
// (tokenStore.ts) share a single `exp` check instead of duplicating it.
export function isTokenValid(token: string | null): boolean {
  if (!token) return false;
  const parts = token.split(".");
  if (parts.length !== 3) return false;
  const payloadStr = parts[1];
  if (!payloadStr) return false;
  try {
    const payload = JSON.parse(atob(payloadStr)) as { exp?: number };
    return typeof payload.exp === "number"
      && payload.exp > Date.now() / 1000;
  } catch {
    return false;
  }
}
