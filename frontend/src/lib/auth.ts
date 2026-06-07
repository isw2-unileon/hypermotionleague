import router from "@/router";
import { supabase } from "@/lib/supabase";
import { getToken, clearToken } from "@/lib/tokenStore";

// Best-effort, client-side decode of the signed-in user's ID from the JWT in
// memory. This is NOT a security check — the backend verifies the token
// signature on every request (see middleware/auth.go). It only lets the UI know
// "who am I" (e.g. to find my league membership / budget) without an extra
// round-trip. Mirrors the decode already used by the router and LeagueDetailPage.
// The backend signs `user_id` into the claims (see backend/internal/auth/jwt.go).
export function currentUserId(): number {
  const token = getToken();
  if (!token) return 0;

  const segment = token.split(".")[1];
  if (!segment) return 0;

  try {
    const payload: unknown = JSON.parse(atob(segment));
    if (
      typeof payload === "object" &&
      payload !== null &&
      "user_id" in payload &&
      typeof (payload as { user_id: unknown }).user_id === "number"
    ) {
      return (payload as { user_id: number }).user_id;
    }
    return 0;
  } catch {
    return 0;
  }
}

// Ends the session and returns the user to the auth screen. Mirrors the
// existing 401/expiry handling — clear the localStorage "token" and
// router.push("/auth") (the router.beforeEach guard does the rest) — and also
// drops Supabase's own session so an OAuth login doesn't leave a dangling
// provider session behind.
export async function logout(): Promise<void> {
  clearToken();
  try {
    await supabase.auth.signOut();
  } catch {
    // Best-effort: the local token is already cleared, so never block the
    // redirect on a Supabase/network hiccup.
  }
  router.push("/auth");
}
