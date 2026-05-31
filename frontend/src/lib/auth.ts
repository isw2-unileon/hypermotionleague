// Best-effort, client-side decode of the signed-in user's ID from the JWT in
// localStorage. This is NOT a security check — the backend verifies the token
// signature on every request (see middleware/auth.go). It only lets the UI know
// "who am I" (e.g. to find my league membership / budget) without an extra
// round-trip. Mirrors the decode already used by the router and LeagueDetailPage.
// The backend signs `user_id` into the claims (see backend/internal/auth/jwt.go).
export function currentUserId(): number {
  const token = localStorage.getItem("token");
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
