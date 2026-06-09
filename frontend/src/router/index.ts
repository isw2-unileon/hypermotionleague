import { createRouter, createWebHistory } from "vue-router";
import AuthPage from "@/views/AuthPage.vue";
import { getToken, clearToken } from "@/lib/tokenStore";
import { isTokenValid } from "@/lib/jwt";

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
