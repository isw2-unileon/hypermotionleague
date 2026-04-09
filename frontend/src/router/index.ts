import { createRouter, createWebHistory } from "vue-router";
import AuthPage from "@/views/AuthPage.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      redirect: "/auth",
    },
    {
      path: "/auth",
      name: "auth",
      component: AuthPage,
    },
  ],
});

export default router;
