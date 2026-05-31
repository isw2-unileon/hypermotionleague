<template>
  <!--
    FormSpark — mini recent-form bar chart, ported from Mercado.jsx.
    NOTE: no market endpoint currently exposes a per-player recent-form array
    (see Phase 0 #4), so this component is intentionally NOT rendered by
    PlayerCard yet. It is kept here, ready to wire, for when the backend adds a
    points-per-matchday array to the players/listings response.
    TODO: render in PlayerCard once the API returns recent form.
  -->
  <svg :width="width" :height="height" :viewBox="`0 0 ${width} ${height}`">
    <rect
      v-for="(v, i) in data"
      :key="i"
      :x="i * (width / data.length) + 1"
      :y="height - barHeight(v)"
      :width="width / data.length - 2"
      :height="barHeight(v)"
      :fill="barColor(v)"
      rx="1"
    />
  </svg>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  data: number[];
  big?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  big: false,
});

const width = computed(() => (props.big ? 80 : 56));
const height = computed(() => (props.big ? 24 : 18));

function barHeight(v: number): number {
  return (v / 12) * (props.big ? 22 : 16);
}

function barColor(v: number): string {
  if (v >= 8) return "var(--lime)";
  if (v >= 5) return "var(--ink-200)";
  return "var(--ink-500)";
}
</script>
