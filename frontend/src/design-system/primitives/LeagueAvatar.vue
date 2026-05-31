<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  seed: number
  size?: number
  accent?: string
}

const props = withDefaults(defineProps<Props>(), {
  size: 40,
  accent: 'var(--lime)',
})

const patterns = computed<string[]>(() => {
  const a = props.accent
  return [
    // 0: diagonal stripes
    `<rect width="44" height="44" fill="#1B1F27" />
     <path d="M-10 30 L20 0 M-10 50 L40 0 M10 50 L60 0 M30 50 L70 10" style="stroke: ${a};" stroke-width="3" fill="none" />`,
    // 1: grid
    `<rect width="44" height="44" fill="#262B36" />
     <rect x="6" y="6" width="14" height="14" style="fill: ${a};" />
     <rect x="24" y="24" width="14" height="14" style="fill: ${a};" />
     <rect x="24" y="6" width="14" height="14" fill="#1B1F27" style="stroke: ${a};" />
     <rect x="6" y="24" width="14" height="14" fill="#1B1F27" style="stroke: ${a};" />`,
    // 2: chevron
    `<rect width="44" height="44" fill="#1B1F27" />
     <path d="M0 14 L22 0 L44 14 L44 22 L22 8 L0 22 Z" style="fill: ${a};" />
     <path d="M0 30 L22 16 L44 30 L44 38 L22 24 L0 38 Z" style="fill: ${a};" opacity="0.5" />`,
    // 3: dots
    `<rect width="44" height="44" fill="#262B36" />
     <circle cx="11" cy="11" r="3" style="fill: ${a};" />
     <circle cx="22" cy="11" r="3" style="fill: ${a};" opacity="0.6" />
     <circle cx="33" cy="11" r="3" style="fill: ${a};" opacity="0.3" />
     <circle cx="11" cy="22" r="3" style="fill: ${a};" opacity="0.6" />
     <circle cx="22" cy="22" r="3" style="fill: ${a};" />
     <circle cx="33" cy="22" r="3" style="fill: ${a};" opacity="0.6" />
     <circle cx="11" cy="33" r="3" style="fill: ${a};" opacity="0.3" />
     <circle cx="22" cy="33" r="3" style="fill: ${a};" opacity="0.6" />
     <circle cx="33" cy="33" r="3" style="fill: ${a};" />`,
    // 4: half pitch
    `<rect width="44" height="44" style="fill: ${a};" />
     <rect x="0" y="22" width="44" height="22" fill="#0E1014" />
     <line x1="0" y1="22" x2="44" y2="22" stroke="#fff" stroke-width="1" opacity="0.7" />
     <circle cx="22" cy="22" r="5" fill="none" stroke="#fff" stroke-width="1" opacity="0.7" />`,
  ]
})

const pattern = computed(() => patterns.value[props.seed % patterns.value.length])
</script>

<template>
  <svg
    class="league-avatar"
    :width="size"
    :height="size"
    viewBox="0 0 44 44"
    v-html="pattern"
  />
</template>

<style scoped>
.league-avatar {
  border-radius: var(--r-sm);
  display: block;
  flex-shrink: 0;
}
</style>
