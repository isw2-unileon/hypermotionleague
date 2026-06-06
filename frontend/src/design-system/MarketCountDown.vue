<script setup lang="ts">
interface Props {
  label?: string
  timeText: string
  subtitle?: string
}

withDefaults(defineProps<Props>(), {
  label: "PRÓXIMO CIERRE",
  subtitle: "Mercado cierra hoy",
})
</script>

<template>
  <div class="countdown card">
    <div class="countdown-label mono">{{ label }}</div>
    <div class="countdown-time display tnum">{{ timeText }}</div>
    <div class="countdown-sub">{{ subtitle }}</div>
  </div>
</template>

<style scoped>
.countdown {
  padding: 14px;
  background: var(--ink-800);
}

.countdown-label {
  font-size: 9px;
  color: var(--ink-400);
  letter-spacing: 0.15em;
}

.countdown-time {
  font-size: 26px;
  margin-top: 4px;
  color: var(--lime);
}

.countdown-sub {
  font-size: 11px;
  color: var(--ink-300);
  margin-top: 2px;
}
</style>