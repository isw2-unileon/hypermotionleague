<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  initials: string
  size?: number
  color?: string
}

const props = withDefaults(defineProps<Props>(), {
  size: 36,
  color: 'var(--lime)',
})

const rootStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
  border: `1.5px solid ${props.color}`,
  fontSize: `${props.size * 0.32}px`,
}))
</script>

<template>
  <div class="avatar" :style="rootStyle">{{ initials }}</div>
</template>

<style scoped>
.avatar {
  border-radius: 50%;
  background: var(--ink-700);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--f-mono);
  font-weight: 600;
  color: var(--ink-100);
  letter-spacing: 0.05em;
  flex-shrink: 0;
}
</style>
