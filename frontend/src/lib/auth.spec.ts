import { describe, it, expect, vi, beforeEach } from "vitest";

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
import { setToken, getToken } from "./tokenStore";

beforeEach(() => {
  signOut.mockReset();
  signOut.mockResolvedValue({ error: null });
  push.mockReset();
});

describe("logout", () => {
  it("clears the token, signs out of Supabase and redirects to /auth", async () => {
    setToken("fake-jwt");

    await logout();

    expect(getToken()).toBeNull();
    expect(signOut).toHaveBeenCalledOnce();
    expect(push).toHaveBeenCalledWith("/auth");
  });

  it("still clears the token and redirects even if Supabase signOut rejects", async () => {
    signOut.mockRejectedValueOnce(new Error("network"));
    setToken("fake-jwt");

    await logout();

    expect(getToken()).toBeNull();
    expect(push).toHaveBeenCalledWith("/auth");
  });
});
