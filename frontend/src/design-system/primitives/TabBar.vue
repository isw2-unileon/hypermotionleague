<template>
  <div class="tab-bar" :class="{ light }">
    <div
      v-for="t in tabs"
      :key="t.id"
      class="tab-item"
      :class="{ active: active === t.id }"
      role="button"
      :aria-pressed="active === t.id"
      @click="emit('select', t.id)"
    >
      <!-- Trophy -->
      <svg
        v-if="t.icon === 'trophy'"
        width="20"
        height="20"
        viewBox="0 0 20 20"
        fill="none"
        :stroke="iconColor(active === t.id)"
        stroke-width="1.5"
      >
        <path d="M5 3 H15 V8 A5 5 0 0 1 5 8 Z" />
        <path d="M5 5 H2 V7 A2 2 0 0 0 5 8" />
        <path d="M15 5 H18 V7 A2 2 0 0 1 15 8" />
        <path d="M8 14 H12 V17 H8 Z" />
        <path d="M10 12 V14" />
      </svg>

      <!-- Chart -->
      <svg
        v-else-if="t.icon === 'chart'"
        width="20"
        height="20"
        viewBox="0 0 20 20"
        fill="none"
        :stroke="iconColor(active === t.id)"
        stroke-width="1.5"
      >
        <rect x="3" y="11" width="3" height="6" />
        <rect x="8.5" y="6" width="3" height="11" />
        <rect x="14" y="3" width="3" height="14" />
      </svg>

      <!-- Shirt -->
      <svg
        v-else-if="t.icon === 'shirt'"
        width="20"
        height="20"
        viewBox="0 0 20 20"
        fill="none"
        :stroke="iconColor(active === t.id)"
        stroke-width="1.5"
      >
        <path d="M5 3 L7 5 H13 L15 3 L18 5 V8 L16 8 V17 H4 V8 L2 8 V5 Z" />
      </svg>

      <!-- Money -->
      <svg
        v-else-if="t.icon === 'money'"
        width="20"
        height="20"
        viewBox="0 0 20 20"
        fill="none"
        :stroke="iconColor(active === t.id)"
        stroke-width="1.5"
      >
        <circle cx="10" cy="10" r="7" />
        <path d="M10 6 V14 M8 8 H11.5 A1.5 1.5 0 0 1 11.5 11 H8.5 A1.5 1.5 0 0 0 8.5 14 H12" />
      </svg>

      <span>{{ t.label }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
export type TabId = "ligas" | "clasif" | "equipo" | "mercado";
type IconName = "trophy" | "chart" | "shirt" | "money";

interface Tab {
  id: TabId;
  label: string;
  icon: IconName;
}

interface Props {
  active?: TabId;
  light?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  active: "ligas",
  light: false,
});

const emit = defineEmits<{ (e: "select", id: TabId): void }>();

const tabs: readonly Tab[] = [
  { id: "ligas", label: "Ligas", icon: "trophy" },
  { id: "clasif", label: "Clasif.", icon: "chart" },
  { id: "equipo", label: "Equipo", icon: "shirt" },
  { id: "mercado", label: "Mercado", icon: "money" },
];

function iconColor(isActive: boolean): string {
  if (isActive) return "var(--lime)";
  return props.light ? "var(--bone-300)" : "var(--ink-400)";
}
</script>
