<script setup lang="ts">
interface Props {
  message: string
  variant?: "success" | "error" | "info"
  visible: boolean
}

withDefaults(defineProps<Props>(), {
  variant: "success",
});

defineEmits<{ (e: "close"): void }>()
</script>

<template>
  <Transition name="toast">
    <div v-if="visible" class="toast" :class="['toast-' + variant]" role="status">
      <span class="toast-icon" aria-hidden="true">
        <span v-if="variant === 'success'">✓</span>
        <span v-else-if="variant === 'error'">⚠</span>
        <span v-else>ℹ</span>
      </span>
      <span class="toast-msg">{{ message }}</span>
      <button type="button" class="toast-close" @click="$emit('close')">×</button>
    </div>
  </Transition>
</template>

<style scoped>
.toast {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 20px;
  border-radius: var(--r-md);
  background: var(--ink-800);
  border: 1px solid var(--ink-700);
  color: var(--ink-100);
  font-family: var(--f-ui);
  font-size: 13px;
  max-width: 480px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  z-index: 1000;
}

.toast-success {
  border-color: var(--lime);
}

.toast-success .toast-icon {
  color: var(--lime);
}

.toast-error {
  border-color: var(--down);
}

.toast-error .toast-icon {
  color: var(--down);
}

.toast-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.toast-msg {
  flex: 1;
  line-height: 1.4;
}

.toast-close {
  background: transparent;
  border: none;
  color: var(--ink-400);
  font-size: 20px;
  line-height: 1;
  cursor: pointer;
  padding: 0;
  margin-left: 4px;
}

.toast-close:hover {
  color: var(--ink-100);
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translate(-50%, 10px);
}
</style>