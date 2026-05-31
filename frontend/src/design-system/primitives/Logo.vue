<template>
  <div class="logo">
    <svg :width="size" :height="size" viewBox="0 0 32 32" fill="none">
      <rect x="1" y="1" width="30" height="30" rx="3" :stroke="color" stroke-width="2" />
      <path
        d="M8 8 L8 24 M14 8 L14 24 M20 8 L20 24 M26 8 L26 24"
        :stroke="color"
        stroke-width="1.5"
        opacity="0.4"
      />
      <path d="M6 16 L26 16" :stroke="color" stroke-width="2" />
      <circle cx="16" cy="16" r="3" :fill="color" />
    </svg>
    <span class="display wordmark" :style="{ fontSize: size * 0.85 + 'px', color }">
      HM/<span class="suffix">LEAGUE</span>
    </span>
  </div>
</template>

<script setup lang="ts">
interface Props {
  size?: number;
  color?: string;
}

withDefaults(defineProps<Props>(), {
  size: 28,
  color: "var(--lime)",
});
</script>

<style scoped>
.logo {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.wordmark {
  line-height: 1;
  letter-spacing: 0.04em;
}

.suffix {
  color: var(--ink-100);
}
</style>
