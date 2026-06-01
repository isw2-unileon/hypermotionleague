<template>
  <div class="app-shell">
    <!-- Mobile-only top bar: the TabBar already holds the 4 nav items, so the
         logout action lives here instead. Hidden on desktop. -->
    <header class="mobile-topbar">
      <button type="button" class="logout-mobile" @click="onLogout">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
          <path d="M6 2 H3 V14 H6" />
          <path d="M10 5 L13 8 L10 11" />
          <path d="M13 8 H6" />
        </svg>
        Cerrar sesión
      </button>
    </header>

    <aside class="sidebar">
      <Logo :size="26" />
      <div class="sidebar-gap" />

      <nav class="nav">
        <button
          v-for="item in navItems"
          :key="item.id"
          type="button"
          class="nav-item"
          :class="{ active: item.id === activeTabId }"
          @click="onNavClick(item)"
        >
          <span class="nav-icon" :class="{ active: item.id === activeTabId }" />
          <span class="nav-label">{{ item.label }}</span>
        </button>
      </nav>

      <div class="spacer" />

      <div class="countdown card">
        <div class="countdown-label mono">PRÓXIMO CIERRE</div>
        <div class="countdown-time display tnum">04:18:42</div>
        <div class="countdown-sub">Mercado cierra hoy</div>
      </div>

      <button type="button" class="btn btn-ghost logout-btn" @click="onLogout">
<<<<<<< Updated upstream
        <svg
          class="logout-icon"
          width="15"
          height="15"
          viewBox="0 0 16 16"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path d="M6 14 H3 a1 1 0 0 1 -1 -1 V3 a1 1 0 0 1 1 -1 h3 M10.5 11 L13.5 8 L10.5 5 M13.5 8 H6" />
=======
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
          <path d="M6 2 H3 V14 H6" />
          <path d="M10 5 L13 8 L10 11" />
          <path d="M13 8 H6" />
>>>>>>> Stashed changes
        </svg>
        Cerrar sesión
      </button>
    </aside>

    <header class="mobile-topbar">
      <Logo :size="22" />
      <button type="button" class="btn btn-ghost logout-btn logout-btn-mobile" @click="onLogout">
        <svg
          class="logout-icon"
          width="14"
          height="14"
          viewBox="0 0 16 16"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path d="M6 14 H3 a1 1 0 0 1 -1 -1 V3 a1 1 0 0 1 1 -1 h3 M10.5 11 L13.5 8 L10.5 5 M13.5 8 H6" />
        </svg>
        Salir
      </button>
    </header>

    <main class="main">
      <slot />
    </main>

    <TabBar
      class="mobile-tabbar"
      :active="activeTabId ?? undefined"
      @select="onTabSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import Logo from "./primitives/Logo.vue";
import TabBar from "./primitives/TabBar.vue";
import { logout } from "@/lib/auth";

type NavId = "ligas" | "clasif" | "equipo" | "mercado";

interface NavItem {
  id: NavId;
  label: string;
  icon: string;
  route: string;
}

const route = useRoute();
const router = useRouter();

const navItems: readonly NavItem[] = [
  { id: "ligas", label: "Mis Ligas", icon: "trophy", route: "/leagues" },
  { id: "clasif", label: "Clasificación", icon: "chart", route: "/standings" },
  { id: "equipo", label: "Mi Equipo", icon: "shirt", route: "/team" },
  { id: "mercado", label: "Mercado", icon: "money", route: "/market" },
];

const activeTabId = computed<NavId | null>(() => {
  const path = route.path;
  if (path.startsWith("/leagues")) return "ligas";
  if (path.startsWith("/standings")) return "clasif";
  if (path.startsWith("/team") || path.startsWith("/squad")) return "equipo";
  if (path.startsWith("/market")) return "mercado";
  return null;
});

function onNavClick(item: NavItem): void {
  if (item.route !== route.path) router.push(item.route);
}

function onTabSelect(id: NavId): void {
  const item = navItems.find((n) => n.id === id);
  if (item && item.route !== route.path) router.push(item.route);
}

<<<<<<< Updated upstream
function onLogout(): void {
  void logout();
=======
// No dedicated auth store: the app's JWT lives in localStorage["token"]
// (see lib/api.ts, which clears it the same way on a 401). Logout drops the
// token and returns the user to the auth screen.
function onLogout(): void {
  localStorage.removeItem("token");
  router.push("/auth");
>>>>>>> Stashed changes
}
</script>

<style scoped>
.app-shell {
  height: 100vh;
  display: grid;
  grid-template-columns: 240px 1fr;
  background: var(--ink-900);
  color: var(--ink-100);
  font-family: var(--f-ui);
}

.sidebar {
  background: var(--ink-850);
  border-right: 1px solid var(--ink-700);
  padding: 32px 20px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  overflow-y: auto;
}

.sidebar-gap {
  height: 32px;
}

.nav {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: var(--r-sm);
  background: transparent;
  border: none;
  color: var(--ink-200);
  font-family: var(--f-ui);
  font-size: 13px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.nav-item:hover {
  background: var(--ink-800);
}

.nav-item.active {
  background: var(--ink-700);
  color: var(--lime);
}

.nav-icon {
  width: 16px;
  height: 16px;
  border-radius: 2px;
  background: var(--ink-500);
  opacity: 0.5;
  flex-shrink: 0;
}

.nav-icon.active {
  background: var(--lime);
  opacity: 1;
}

.nav-label {
  flex: 1;
}

.spacer {
  flex: 1;
}

.countdown {
  padding: 14px;
  background: var(--ink-800);
}

.countdown-label {
  font-size: 9px;
  color: var(--ink-400);
  letter-spacing: 0.15em;
}

.countdown-time {
  font-size: 26px;
  margin-top: 4px;
  color: var(--lime);
}

.countdown-sub {
  font-size: 11px;
  color: var(--ink-300);
  margin-top: 2px;
}

/* Desktop logout — sits at the very bottom of the sidebar, visually distinct
   from the nav items. */
.logout-btn {
  margin-top: 12px;
  width: 100%;
  font-size: 13px;
  color: var(--ink-300);
}

.logout-btn:hover {
  background: transparent;
  color: var(--down);
  border-color: var(--down);
}

/* Mobile top bar (hidden on desktop) */
.mobile-topbar {
  display: none;
}

.logout-mobile {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 6px;
  background: transparent;
  border: none;
  font-family: var(--f-ui);
  font-size: 12px;
  font-weight: 500;
  color: var(--ink-300);
  cursor: pointer;
  transition: color 0.15s;
}

.logout-mobile:hover {
  color: var(--down);
}

.main {
  padding: 40px 56px;
  overflow-y: auto;
}

.mobile-tabbar {
  display: none;
}

/* Logout — reuses the .btn / .btn-ghost tokens; only layout is set here.
   Full width at the bottom of the sidebar, aligned with the nav items. */
.logout-btn {
  width: 100%;
  margin-top: 10px;
  justify-content: flex-start;
  gap: 12px;
  padding: 10px 12px;
  font-size: 13px;
  font-weight: 500;
}

.logout-icon {
  flex-shrink: 0;
}

/* Shell-owned top bar, only used on mobile (sidebar is hidden there). */
.mobile-topbar {
  display: none;
}

@media (max-width: 767px) {
  .app-shell {
    display: flex;
    flex-direction: column;
  }

  .sidebar {
    display: none;
  }

  .mobile-topbar {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    padding: 10px 16px;
    background: var(--ink-850);
    border-bottom: 1px solid var(--ink-700);
    flex-shrink: 0;
  }

  .main {
    flex: 1;
    padding: 20px;
    overflow-y: auto;
  }

  .mobile-tabbar {
    display: flex;
    flex-shrink: 0;
  }

  .mobile-topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 20px;
    background: var(--ink-850);
    border-bottom: 1px solid var(--ink-700);
    flex-shrink: 0;
  }

  /* Compact variant for the mobile top bar (overrides the sidebar layout). */
  .logout-btn-mobile {
    width: auto;
    margin-top: 0;
    padding: 8px 12px;
    font-size: 12px;
    gap: 8px;
  }
}
</style>
