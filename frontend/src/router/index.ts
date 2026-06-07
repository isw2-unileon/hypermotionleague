import { createRouter, createWebHistory } from "vue-router";
import AuthPage from "@/views/AuthPage.vue";
import { getToken, clearToken } from "@/lib/tokenStore";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/auth",
      name: "auth",
      component: AuthPage,
      meta: { public: true },
    },
    {
      path: "/leagues",
      name: "leagues",
      component: () => import("@/views/LeaguesPage.vue"),
    },
    {
      path: "/leagues/new",
      name: "create-league",
      component: () => import("@/views/CreateLeaguePage.vue"),
    },
    {
      path: "/leagues/:id",
      name: "league-detail",
      component: () => import("@/views/LeagueDetailPage.vue"),
    },
    { path: "/standings", 
      name: "standings",
      component: () => import("@/views/StandingsPage.vue") 
    },
   
    {
      path: "/squad/:leagueId/:userId",
      name: "user-squad",
      component: () => import("@/views/UserSquadPage.vue"),
    },
    {
      path: "/team",
      name: "team",
      component: () => import("@/views/TeamPage.vue"),
    },
    {
      path: "/lineup/:leagueId/:matchdayNumber",
      name: "lineup",
      component: () => import("@/views/LineupPage.vue"),
    },
    {
      path: "/market",
      name: "market",
      component: () => import("@/views/MarketPage.vue"),
    },
    {
      path: "/:pathMatch(.*)*",
      redirect: "/leagues",
    },
   
  ],
});

// Best-effort client-side check of the JWT `exp` claim. Not a security
// verification — the backend still verifies the signature on every request.
// The goal is to avoid sending obviously-expired tokens and to redirect the
// user to /auth without waiting for a 401.
function isTokenValid(token: string | null): boolean {
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

router.beforeEach((to) => {
  const token = getToken();
  const valid = isTokenValid(token);

  // Expired or malformed token: clear it so subsequent requests don't carry
  // a stale credential, then treat the user as unauthenticated.
  if (token && !valid) {
    clearToken();
  }

  if (!to.meta.public && !valid) {
    return "/auth";
  }

  if (to.path === "/auth" && valid) {
    return "/leagues";
  }
});

export default router;
