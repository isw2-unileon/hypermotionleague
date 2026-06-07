<template>
  <div class="app-shell">
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

      <MarketCountdown :label="label" :time-text="timeText" :subtitle="subtitle" />

      <button type="button" class="btn btn-ghost logout-btn" @click="onLogout">
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
import MarketCountdown from "./MarketCountDown.vue";
import { useMarketCountdown } from "@/composables/useMarketCountdown";
import { activeLeagueId } from "@/lib/activeLeague";
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

// Selected league: read from route params/query first; fall back to the
// reactive activeLeagueId that individual pages write when they resolve a league.
const selectedLeagueId = computed<number | null>(() => {
  const params = route.params as Record<string, string | undefined>;
  const fromParams = params.leagueId ?? params.id;
  if (fromParams != null) {
    const n = Number(fromParams);
    if (!Number.isNaN(n)) return n;
  }
  const query = route.query.league;
  const raw = Array.isArray(query) ? query[0] : query;
  if (raw != null) {
    const n = Number(raw);
    if (!Number.isNaN(n)) return n;
  }
  return activeLeagueId.value;
});

const { timeText, label, subtitle } = useMarketCountdown(() => selectedLeagueId.value);

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


function onLogout(): void {
  void logout();
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


/* Logout — reuses the .btn / .btn-ghost tokens; only layout/colour set here.
   Full width at the bottom of the sidebar, aligned with the nav items. */
.logout-btn {
  width: 100%;
  margin-top: 10px;
  justify-content: flex-start;
  gap: 12px;
  padding: 10px 12px;
  font-size: 13px;
  font-weight: 500;
  color: var(--ink-300);
}

.logout-btn:hover {
  background: transparent;
  color: var(--down);
  border-color: var(--down);
}

.logout-icon {
  flex-shrink: 0;
}

.main {
  padding: 40px 56px;
  overflow-y: auto;
}

/* Hidden on desktop: the sidebar carries its own logout there. */
.mobile-topbar {
  display: none;
}

.mobile-tabbar {
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

  /* Single mobile header: Logo on the left, "Salir" on the right. */
  .mobile-topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 20px;
    background: var(--ink-850);
    border-bottom: 1px solid var(--ink-700);
    flex-shrink: 0;
  }

  /* Compact logout for the mobile top bar (overrides the sidebar layout). */
  .logout-btn-mobile {
    width: auto;
    margin-top: 0;
    padding: 8px 12px;
    font-size: 12px;
    gap: 8px;
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
}
</style>
