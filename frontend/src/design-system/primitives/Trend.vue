<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  delta: number
}

const props = defineProps<Props>()

const isZero = computed(() => props.delta === 0)
const isUp = computed(() => props.delta > 0)
const abs = computed(() => Math.abs(props.delta))
</script>

<template>
  <span v-if="isZero" class="trend trend-zero">—</span>
  <span v-else class="trend" :class="isUp ? 'trend-up' : 'trend-down'">
    <svg width="8" height="8" viewBox="0 0 8 8" aria-hidden="true">
      <path
        :d="isUp ? 'M4 1 L7 6 L1 6 Z' : 'M4 7 L7 2 L1 2 Z'"
        fill="currentColor"
      />
    </svg>
    {{ abs }}
  </span>
</template>

<style scoped>
.trend {
  font-family: var(--f-mono);
  font-size: 11px;
  display: inline-flex;
  align-items: center;
}
.trend-zero {
  color: var(--ink-400);
  gap: 3px;
}
.trend-up,
.trend-down {
  gap: 2px;
  font-weight: 600;
}
.trend-up {
  color: var(--up);
}
.trend-down {
  color: var(--down);
}
</style>
