import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// Rehydration runs once at module load, so each test arranges storage and then
// re-imports the module (via vi.resetModules + dynamic import) to exercise it.
// The Node test runtime's ambient localStorage/sessionStorage is a partial stub
// (missing removeItem/clear), so we install a working in-memory Storage mock —
// the source reads these globals at call time, so the re-imported module uses it.

const KEY = "hml.auth.token";

function createStorageMock(): Storage {
  const map = new Map<string, string>();
  return {
    get length() {
      return map.size;
    },
    clear: () => map.clear(),
    getItem: (k: string) => (map.has(k) ? (map.get(k) as string) : null),
    key: (i: number) => Array.from(map.keys())[i] ?? null,
    removeItem: (k: string) => void map.delete(k),
    setItem: (k: string, v: string) => void map.set(k, String(v)),
  };
}

// Minimal JWT (header.payload.signature) with a given `exp` (seconds). Encoded
// with btoa so the rehydration's atob(payload) decodes it back.
function makeJwt(expSeconds: number): string {
  const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const payload = btoa(JSON.stringify({ exp: expSeconds, user_id: 1 }));
  return `${header}.${payload}.sig`;
}

const FUTURE = Math.floor(Date.now() / 1000) + 3600;
const PAST = Math.floor(Date.now() / 1000) - 3600;

beforeEach(() => {
  vi.stubGlobal("localStorage", createStorageMock());
  vi.stubGlobal("sessionStorage", createStorageMock());
  vi.resetModules();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("tokenStore rehydration", () => {
  it("rehydrates a valid token from localStorage on load", async () => {
    const token = makeJwt(FUTURE);
    localStorage.setItem(KEY, token);

    const { getToken } = await import("./tokenStore");

    expect(getToken()).toBe(token);
  });

  it("rehydrates a valid token from sessionStorage on load", async () => {
    const token = makeJwt(FUTURE);
    sessionStorage.setItem(KEY, token);

    const { getToken } = await import("./tokenStore");

    expect(getToken()).toBe(token);
  });

  it("discards and cleans up an expired token in storage on load", async () => {
    localStorage.setItem(KEY, makeJwt(PAST));

    const { getToken } = await import("./tokenStore");

    expect(getToken()).toBeNull();
    expect(localStorage.getItem(KEY)).toBeNull();
  });

  it("discards a malformed token in storage on load", async () => {
    sessionStorage.setItem(KEY, "not-a-jwt");

    const { getToken } = await import("./tokenStore");

    expect(getToken()).toBeNull();
    expect(sessionStorage.getItem(KEY)).toBeNull();
  });
});

describe("tokenStore persistence", () => {
  it("setToken(token, true) persists to localStorage only", async () => {
    const { setToken } = await import("./tokenStore");
    const token = makeJwt(FUTURE);

    setToken(token, true);

    expect(localStorage.getItem(KEY)).toBe(token);
    expect(sessionStorage.getItem(KEY)).toBeNull();
  });

  it("setToken(token, false) persists to sessionStorage only", async () => {
    const { setToken } = await import("./tokenStore");
    const token = makeJwt(FUTURE);

    setToken(token, false);

    expect(sessionStorage.getItem(KEY)).toBe(token);
    expect(localStorage.getItem(KEY)).toBeNull();
  });

  it("keeps the token in a single store when rememberMe changes", async () => {
    const { setToken } = await import("./tokenStore");

    setToken(makeJwt(FUTURE), true); // localStorage
    const next = makeJwt(FUTURE);
    setToken(next, false); // should move to sessionStorage only

    expect(localStorage.getItem(KEY)).toBeNull();
    expect(sessionStorage.getItem(KEY)).toBe(next);
  });

  it("clearToken wipes memory and both storages", async () => {
    const { setToken, clearToken, getToken } = await import("./tokenStore");
    // Simulate a residue from a previous session with the other option.
    localStorage.setItem(KEY, "stale-local");
    setToken(makeJwt(FUTURE), false);

    clearToken();

    expect(getToken()).toBeNull();
    expect(localStorage.getItem(KEY)).toBeNull();
    expect(sessionStorage.getItem(KEY)).toBeNull();
  });
});
