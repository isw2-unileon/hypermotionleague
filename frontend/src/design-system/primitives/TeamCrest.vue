<template>
  <!--
    TeamCrest — "solid square with short code" (F3.2 primitive). Reconstructed
    from the brief + usage in Mercado.jsx (primitives.jsx was never committed).
  -->
  <span class="team-crest mono" :style="rootStyle">{{ short }}</span>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  short: string; // short team code, e.g. "LIV"
  color: string; // team accent color (any CSS color)
  size?: number;
}

const props = withDefaults(defineProps<Props>(), {
  size: 18,
});

const rootStyle = computed(() => ({
  height: `${props.size}px`,
  minWidth: `${props.size}px`,
  background: props.color,
  fontSize: `${Math.max(7, Math.round(props.size * 0.42))}px`,
}));
</script>

<style scoped>
.team-crest {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
  border-radius: var(--r-xs);
  color: #fff;
  font-weight: 600;
  letter-spacing: 0.06em;
  line-height: 1;
}
</style>
