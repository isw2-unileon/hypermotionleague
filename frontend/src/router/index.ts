import { createRouter, createWebHistory } from "vue-router";
import AuthPage from "@/views/AuthPage.vue";

const PlaceholderPage = {
  template: `
    <div class="flex items-center justify-center h-full p-8">
      <div class="text-center">
        <p class="text-4xl mb-4">🚧</p>
        <h2 class="text-xl font-bold text-white mb-2">Próximamente</h2>
        <p class="text-green-300/60 text-sm">Esta sección está en desarrollo</p>
      </div>
    </div>
  `,
};

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
    {
      path: "/standings",
      name: "standings",
      component: PlaceholderPage,
    },
    {
      path: "/team",
      name: "team",
      component: PlaceholderPage,
    },
    {
      path: "/market",
      name: "market",
      component: PlaceholderPage,
    },
    {
      path: "/:pathMatch(.*)*",
      redirect: "/leagues",
    },
    { path: "/standings", 
      name: "standings",
      component: () => import("@/views/StandingsPage.vue") },

  ],
});

router.beforeEach((to) => {
  const token = localStorage.getItem("token");

  if (!to.meta.public && !token) {
    return "/auth";
  }

  if (to.path === "/auth" && token) {
    return "/leagues";
  }
});

export default router;
