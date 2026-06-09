import { isTokenValid } from "@/lib/jwt";

// The signed-in user's backend JWT — the single source of truth for "estoy
// logueado" (Supabase is only a one-shot OAuth bridge; its session is dropped
// after we trade it for this token). Persisted so it survives a manual page
// refresh (F5), which re-evaluates every module from scratch and would otherwise
// reset this to null and bounce the user to /auth.
//
// Two backing stores, chosen by the login "Mantener sesión" checkbox (rememberMe):
//   - rememberMe = true  -> localStorage   (survives closing the browser)
//   - rememberMe = false -> sessionStorage (survives F5, gone when the tab closes)
// The token lives in exactly ONE store at a time: setToken() always clears the
// other store, so whichever store holds it on reload implicitly tells us which
// mode was chosen — no separate flag needed. clearToken() wipes both, so a
// previous session with the other rememberMe option never leaves a residue.
export const TOKEN_KEY = "hml.auth.token";

// Lazy, synchronous rehydration at module load — runs before the first
// router.beforeEach, so the guard sees the restored token instead of null.
let _token: string | null = rehydrate();

function rehydrate(): string | null {
  try {
    // sessionStorage first, then localStorage. Order is irrelevant for
    // correctness (the token is only ever in one), but keeps lookup deterministic.
    const stored = sessionStorage.getItem(TOKEN_KEY) ?? localStorage.getItem(TOKEN_KEY);
    if (isTokenValid(stored)) return stored;
    // Expired or malformed: drop it so it can't linger across reloads.
    if (stored) clearStorage();
    return null;
  } catch {
    // Storage unavailable (e.g. Safari private mode): fall back to memory-only.
    return null;
  }
}

function clearStorage(): void {
  try {
    localStorage.removeItem(TOKEN_KEY);
    sessionStorage.removeItem(TOKEN_KEY);
  } catch {
    // Best-effort: nothing to clean if storage is unavailable.
  }
}

export function getToken(): string | null {
  return _token;
}

export function setToken(token: string, rememberMe = true): void {
  _token = token;
  try {
    const store = rememberMe ? localStorage : sessionStorage;
    const other = rememberMe ? sessionStorage : localStorage;
    other.removeItem(TOKEN_KEY); // keep the token in exactly one store
    store.setItem(TOKEN_KEY, token);
  } catch {
    // Storage unavailable: the in-memory token still authenticates this tab;
    // it just won't survive a refresh.
  }
}

export function clearToken(): void {
  _token = null;
  clearStorage();
}
