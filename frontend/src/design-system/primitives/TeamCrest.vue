<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  team: string
  color: string
  size?: number
}

const props = withDefaults(defineProps<Props>(), {
  size: 22,
})

const rootStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
  background: props.color,
  fontSize: `${props.size * 0.42}px`,
}))
</script>

<template>
  <div class="team-crest" :style="rootStyle">{{ team }}</div>
</template>

<style scoped>
.team-crest {
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--f-display);
  color: #fff;
  letter-spacing: 0.04em;
  flex-shrink: 0;
}
</style>
