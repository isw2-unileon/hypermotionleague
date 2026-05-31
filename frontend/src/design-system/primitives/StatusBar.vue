<template>
  <div class="status-bar" :style="{ color: dark ? 'var(--ink-100)' : 'var(--bone-900)' }">
    <span class="time">{{ time }}</span>
    <div class="icons">
      <svg width="18" height="10" viewBox="0 0 18 10" fill="currentColor">
        <rect x="0" y="6" width="3" height="4" rx="0.5" />
        <rect x="5" y="4" width="3" height="6" rx="0.5" />
        <rect x="10" y="2" width="3" height="8" rx="0.5" />
        <rect x="15" y="0" width="3" height="10" rx="0.5" />
      </svg>
      <svg width="16" height="11" viewBox="0 0 16 11" fill="none" stroke="currentColor" stroke-width="1.4">
        <path d="M1 4 Q8 -1 15 4 M3 6 Q8 2 13 6 M5.5 8 Q8 6.5 10.5 8" />
      </svg>
      <svg width="24" height="11" viewBox="0 0 24 11" fill="none" stroke="currentColor" stroke-width="1.2">
        <rect x="0.5" y="0.5" width="20" height="10" rx="2" />
        <rect x="2" y="2" width="14" height="7" rx="1" fill="currentColor" />
        <rect x="21.5" y="3.5" width="2" height="4" rx="0.5" fill="currentColor" />
      </svg>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  time?: string;
  dark?: boolean;
}

withDefaults(defineProps<Props>(), {
  time: "20:47",
  dark: true,
});
</script>

<style scoped>
.time {
  font-weight: 600;
}

.icons {
  display: flex;
  gap: 6px;
  align-items: center;
}
</style>
