<template>
  <svg
    class="pitch"
    viewBox="0 0 500 800"
    preserveAspectRatio="xMidYMid slice"
  >
    <!-- Field outline -->
    <rect x="40" y="40" width="420" height="720" :fill="fill" :stroke="stroke" stroke-width="1.5" />
    <!-- Halfway line -->
    <line x1="40" y1="400" x2="460" y2="400" :stroke="stroke" stroke-width="1.5" />
    <!-- Center circle + spot -->
    <circle cx="250" cy="400" r="70" fill="none" :stroke="stroke" stroke-width="1.5" />
    <circle cx="250" cy="400" r="2" :fill="stroke" />
    <!-- Top penalty area (big box) -->
    <rect x="150" y="40" width="200" height="120" fill="none" :stroke="stroke" stroke-width="1.5" />
    <!-- Top goal area (small box) -->
    <rect x="200" y="40" width="100" height="50" fill="none" :stroke="stroke" stroke-width="1.5" />
    <!-- Top penalty arc (bulges downward into field) -->
    <path d="M 175 160 A 70 70 0 0 0 325 160" fill="none" :stroke="stroke" stroke-width="1.5" />
    <!-- Bottom penalty area (big box) -->
    <rect x="150" y="640" width="200" height="120" fill="none" :stroke="stroke" stroke-width="1.5" />
    <!-- Bottom goal area (small box) -->
    <rect x="200" y="710" width="100" height="50" fill="none" :stroke="stroke" stroke-width="1.5" />
    <!-- Bottom penalty arc (bulges upward into field) -->
    <path d="M 175 640 A 70 70 0 0 1 325 640" fill="none" :stroke="stroke" stroke-width="1.5" />
  </svg>
</template>

<script setup lang="ts">
interface Props {
  stroke?: string;
  fill?: string;
}

withDefaults(defineProps<Props>(), {
  stroke: "rgba(255,255,255,0.12)",
  fill: "transparent",
});
</script>

<style scoped>
/* PitchSVG anchors to the nearest positioned ancestor — parent must be
   position: relative (or absolute/fixed) for the lines to fill it. */
.pitch {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}
</style>
