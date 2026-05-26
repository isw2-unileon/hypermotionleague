<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  team: string
  color: string
  size?: number
}

const props = withDefaults(defineProps<Props>(), {
  size: 56,
})

const rootStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
  background: `linear-gradient(160deg, ${props.color}33, var(--ink-700))`,
}))

const svgStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
}))
</script>

<template>
  <div class="player-photo" :style="rootStyle">
    <svg
      class="player-photo__silhouette"
      :style="svgStyle"
      viewBox="0 0 56 56"
      aria-hidden="true"
    >
      <circle cx="28" cy="20" r="9" fill="rgba(255,255,255,0.25)" />
      <path d="M10 56 Q10 38 28 38 Q46 38 46 56 Z" fill="rgba(255,255,255,0.2)" />
    </svg>
    <div class="player-photo__tag">{{ team }}</div>
  </div>
</template>

<style scoped>
.player-photo {
  border-radius: 8px;
  border: 1px solid var(--ink-600);
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  flex-shrink: 0;
}
.player-photo__silhouette {
  position: absolute;
  bottom: 0;
}
.player-photo__tag {
  position: absolute;
  top: 4px;
  left: 4px;
  font-family: var(--f-mono);
  font-size: 8px;
  letter-spacing: 0.1em;
  color: rgba(255, 255, 255, 0.7);
}
</style>
