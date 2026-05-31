<template>
  <!--
    PlayerPhoto — "silhouette card with team tag" (F3.2 primitive).
    primitives.jsx was never committed, so this is reconstructed from the brief's
    description and its usage in Mercado.jsx: a rounded card tinted by the team
    color, a generic player silhouette, and the short team code tagged at the
    bottom. The backend has no player photos, so the silhouette is intentional.
  -->
  <div class="player-photo" :style="rootStyle">
    <svg class="silhouette" viewBox="0 0 64 64" fill="none" aria-hidden="true">
      <circle cx="32" cy="23" r="11" fill="rgba(0,0,0,0.28)" />
      <path d="M11 62 C11 45 21 39 32 39 C43 39 53 45 53 62 Z" fill="rgba(0,0,0,0.28)" />
    </svg>
    <span class="team-tag mono">{{ team }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  team: string; // short team code, e.g. "LIV"
  color: string; // team accent color (any CSS color)
  size?: number;
}

const props = withDefaults(defineProps<Props>(), {
  size: 64,
});

const rootStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
  background: `linear-gradient(165deg, ${props.color}, var(--ink-800))`,
}));
</script>

<style scoped>
.player-photo {
  position: relative;
  flex-shrink: 0;
  border-radius: var(--r-md);
  border: 1px solid var(--ink-700);
  overflow: hidden;
}

.silhouette {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.team-tag {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  text-align: center;
  font-size: 8px;
  letter-spacing: 0.12em;
  padding: 2px 0;
  color: var(--ink-100);
  background: rgba(8, 9, 11, 0.55);
}
</style>