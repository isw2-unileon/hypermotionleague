import { test, expect, type Page, type Browser } from "@playwright/test";

// Happy-path lifecycle: two users register, one creates a league, the other
// joins, both see each other in the league. Uses two browser contexts so
// cookies/localStorage don't collide between users.
//
// Idempotency: emails and usernames embed Date.now() + a random suffix so the
// test can re-run against a non-cleaned DB. Proper DB cleanup fixtures will
// land later.

interface TestUser {
  username: string;
  displayName: string;
  email: string;
  password: string;
}

function newUser(label: string): TestUser {
  const stamp = `${Date.now()}-${Math.random().toString(36).slice(2, 6)}`;
  return {
    username: `u-${label}-${stamp}`.slice(0, 50),
    displayName: `User ${label} ${stamp}`,
    email: `user-${label}-${stamp}@test.local`,
    password: "password123",
  };
}

async function registerUser(page: Page, user: TestUser): Promise<void> {
  await page.goto("/auth");
  await page.getByRole("button", { name: "Registrarse" }).click();
  await page.getByPlaceholder("tu_nombre").fill(user.username);
  await page.getByPlaceholder("Nombre que verán los demás").fill(user.displayName);
  await page.getByPlaceholder("tu@email.com").fill(user.email);
  await page.getByPlaceholder("Mínimo 8 caracteres").fill(user.password);
  await page.getByRole("button", { name: "Fichar como mánager" }).click();
  await expect(page).toHaveURL(/\/leagues$/);
}

async function openContext(browser: Browser): Promise<Page> {
  const context = await browser.newContext();
  return context.newPage();
}

test("league lifecycle: A creates, B joins, both see each other", async ({ browser }) => {
  const userA = newUser("a");
  const userB = newUser("b");

  const pageA = await openContext(browser);
  const pageB = await openContext(browser);

  // Step 1: register user A.
  await registerUser(pageA, userA);

  // Step 2: user A creates a league and we capture the invite code shown on
  // the resulting LeagueDetailPage.
  await pageA.goto("/leagues/new");
  const leagueName = `E2E League ${Date.now()}`;
  await pageA.getByPlaceholder("Liga de amigos").fill(leagueName);
  await pageA.getByRole("button", { name: "Crear liga" }).click();
  await expect(pageA).toHaveURL(/\/leagues\/\d+$/);
  // The invite code is rendered inside the only <code> element on the page.
  const inviteCode = (await pageA.locator("code").first().innerText()).trim();
  expect(inviteCode).not.toEqual("");

  // Step 3: register user B in a fresh context.
  await registerUser(pageB, userB);

  // Step 4: user B joins via the invite code.
  await pageB.goto("/leagues");
  await pageB.getByPlaceholder("Código de invitación").fill(inviteCode);
  await pageB.getByRole("button", { name: "Unirse" }).click();
  await expect(pageB).toHaveURL(/\/leagues\/\d+$/);

  // Step 5: user A saves a 4-4-2 lineup.
  // TODO: enable once two prerequisites are in place.
  //   (a) Migration 003 must be applied — until then uq_lineup_position
  //       rejects any formation with more than one starter per position.
  //   (b) User A must own at least 11 players matching a 4-4-2 (1 GK, 4 DEF,
  //       4 MID, 2 FWD). This test does not seed market purchases, so the
  //       lineup page would show "No tienes jugadores en esta liga".
  // Add roster-seeding fixtures (e.g. direct DB inserts via a test-only
  // endpoint or a fixture script) before un-skipping.
  test.info().annotations.push({
    type: "skip-step",
    description: "lineup save deferred — needs migration 003 + roster seeding",
  });

  // Step 6: both users see two members listed.
  // The spec said "standings", but the standings rankings table is sourced
  // from matchday data this test does not seed (no matchdays exist for a
  // freshly-created league), so it would render empty. The functional intent
  // — "both users see each other as members of the league" — is verified via
  // the LeagueDetailPage member list, which is also what Task 1.3 enriched
  // with display names from the users JOIN.
  for (const page of [pageA, pageB]) {
    // Navigate via /leagues so each user picks their own league id.
    await page.goto("/leagues");
    await page.getByRole("link", { name: new RegExp(leagueName) }).click();
    await expect(page).toHaveURL(/\/leagues\/\d+$/);

    const memberList = page.locator("h2", { hasText: "Miembros" }).locator("..");
    await expect(memberList).toContainText(userA.displayName);
    await expect(memberList).toContainText(userB.displayName);
  }
});
