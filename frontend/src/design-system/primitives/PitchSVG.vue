<template>
  <svg
    class="pitch"
    viewBox="0 0 800 500"
    preserveAspectRatio="xMidYMid slice"
  >
    <rect x="40" y="40" width="720" height="420" :fill="fill" :stroke="stroke" stroke-width="1.5" />
    <line x1="400" y1="40" x2="400" y2="460" :stroke="stroke" stroke-width="1.5" />
    <circle cx="400" cy="250" r="70" fill="none" :stroke="stroke" stroke-width="1.5" />
    <circle cx="400" cy="250" r="2" :fill="stroke" />
    <rect x="40" y="150" width="120" height="200" fill="none" :stroke="stroke" stroke-width="1.5" />
    <rect x="40" y="200" width="50" height="100" fill="none" :stroke="stroke" stroke-width="1.5" />
    <rect x="640" y="150" width="120" height="200" fill="none" :stroke="stroke" stroke-width="1.5" />
    <rect x="710" y="200" width="50" height="100" fill="none" :stroke="stroke" stroke-width="1.5" />
    <path d="M 160 175 A 70 70 0 0 1 160 325" fill="none" :stroke="stroke" stroke-width="1.5" />
    <path d="M 640 175 A 70 70 0 0 0 640 325" fill="none" :stroke="stroke" stroke-width="1.5" />
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
