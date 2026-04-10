<template>
  <div class="min-h-screen flex flex-col bg-gradient-to-br from-green-900 via-green-800 to-emerald-900">
    <!-- Page content -->
    <main class="flex-1 overflow-y-auto pb-20">
      <slot />
    </main>

    <!-- Bottom navigation -->
    <nav class="fixed bottom-0 inset-x-0 bg-green-950/90 backdrop-blur-md border-t border-white/10">
      <div class="flex justify-around items-center h-16 max-w-lg mx-auto">
        <RouterLink
          v-for="tab in tabs"
          :key="tab.to"
          :to="tab.to"
          class="flex flex-col items-center gap-1 px-3 py-2 rounded-lg transition-colors"
          :class="isActive(tab.to) ? 'text-amber-400' : 'text-green-300/60 hover:text-green-200'"
        >
          <span class="text-xl" v-html="tab.icon"></span>
          <span class="text-[10px] font-semibold tracking-wide">{{ tab.label }}</span>
        </RouterLink>
      </div>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, useRoute } from "vue-router";

const route = useRoute();

const tabs = [
  { to: "/leagues", label: "Mis Ligas", icon: "🏆" },
  { to: "/standings", label: "Clasificación", icon: "📊" },
  { to: "/team", label: "Equipo", icon: "👕" },
  { to: "/market", label: "Mercado", icon: "💰" },
];

function isActive(path: string): boolean {
  return route.path.startsWith(path);
}
</script>
