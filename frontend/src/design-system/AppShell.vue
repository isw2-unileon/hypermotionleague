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

      <div class="countdown card">
        <div class="countdown-label mono">PRÓXIMO CIERRE</div>
        <div class="countdown-time display tnum">04:18:42</div>
        <div class="countdown-sub">Mercado cierra hoy</div>
      </div>
    </aside>

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

.main {
  padding: 40px 56px;
  overflow-y: auto;
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
