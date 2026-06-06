import { test, expect } from "@playwright/test";

// The app has no "/" route. Root resolves through the catch-all to /leagues,
// which is guarded, so an unauthenticated visit ends up at /auth. We assert
// that redirect and that the auth page actually rendered (its tab controls
// are always present in both viewports).
test("homepage redirects unauthenticated users to auth", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveURL(/\/auth$/);

  // The login/register toggle is always rendered. These are role="tab",
  // not buttons (see the auth page markup).
  await expect(page.getByRole("tab", { name: "Iniciar sesión" })).toBeVisible();
  await expect(page.getByRole("tab", { name: "Registrarse" })).toBeVisible();
});

test("health endpoint responds", async ({ request }) => {
  const response = await request.get("http://localhost:8080/health");
  expect(response.ok()).toBeTruthy();
  expect(await response.json()).toEqual({ status: "ok" });
});