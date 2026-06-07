import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// vi.hoisted so these spies exist when the (hoisted) vi.mock factories below run.
const { signOut, push } = vi.hoisted(() => ({
  signOut: vi.fn(),
  push: vi.fn(),
}));

// logout() drops the Supabase session (A5 fix) and redirects via the router
// singleton, so both are mocked here.
vi.mock("@/lib/supabase", () => ({
  supabase: { auth: { signOut } },
}));
vi.mock("@/router", () => ({
  default: { push },
}));

import { logout } from "./auth";

// happy-dom in this project doesn't expose a functional Storage, so we install a
// minimal in-memory localStorage for these tests (the app uses the real one in
// the browser).
function installLocalStorage(): void {
  const store = new Map<string, string>();
  vi.stubGlobal("localStorage", {
    getItem: (k: string) => store.get(k) ?? null,
    setItem: (k: string, v: string) => void store.set(k, v),
    removeItem: (k: string) => void store.delete(k),
    clear: () => store.clear(),
  });
}

beforeEach(() => {
  signOut.mockReset();
  signOut.mockResolvedValue({ error: null });
  push.mockReset();
  installLocalStorage();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("logout", () => {
  it("clears the token, signs out of Supabase and redirects to /auth", async () => {
    localStorage.setItem("token", "fake-jwt");

    await logout();

    expect(localStorage.getItem("token")).toBeNull();
    expect(signOut).toHaveBeenCalledOnce();
    expect(push).toHaveBeenCalledWith("/auth");
  });

  it("still clears the token and redirects even if Supabase signOut rejects", async () => {
    signOut.mockRejectedValueOnce(new Error("network"));
    localStorage.setItem("token", "fake-jwt");

    await logout();

    expect(localStorage.getItem("token")).toBeNull();
    expect(push).toHaveBeenCalledWith("/auth");
  });
});
