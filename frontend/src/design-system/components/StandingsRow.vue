<script setup lang="ts">
import { computed } from 'vue'
import type { StandingsRow } from '@/types/standings'
import Avatar from '../primitives/Avatar.vue'
import Trend from '../primitives/Trend.vue'

interface Props {
  row: StandingsRow
  mobile?: boolean
  light?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  mobile: false,
  light: false,
})

const emit = defineEmits<{ (e: 'click', userId: number): void }>()

const avatarSize = computed(() => (props.mobile ? 32 : 38))
const formattedPoints = computed(() => props.row.totalPoints.toLocaleString('es'))
</script>

<template>
  <div
    class="row"
    :class="{ mobile, current: row.isCurrentUser }"
    role="button"
    tabindex="0"
    @click="emit('click', row.userId)"
    @keydown.enter="emit('click', row.userId)"
  >
    <div class="position display tnum">#{{ row.position }}</div>

    <div class="manager">
      <Avatar :initials="row.initials" :size="avatarSize" :color="row.color" />
      <div class="manager-text">
        <div class="name">{{ row.name }}</div>
        <div class="squad mono">{{ row.squadName }}</div>
      </div>
    </div>

    <div class="trend">
      <Trend :delta="row.deltaPosition" />
    </div>

    <div class="points display tnum">{{ formattedPoints }}</div>
  </div>
</template>

<style scoped>
.row {
  display: grid;
  grid-template-columns: 56px 1fr auto 92px;
  gap: 12px;
  align-items: center;
  padding: 14px 18px;
  border-left: 3px solid transparent;
  border-radius: var(--r-sm);
  background: transparent;
  cursor: pointer;
  transition: background 0.15s;
}

.row:hover {
  background: var(--ink-850);
}

.row.current {
  border-left: 3px solid var(--lime);
  background: var(--lime-glow);
}

.row.mobile {
  grid-template-columns: 40px 1fr auto 64px;
  gap: 10px;
  padding: 12px 14px;
}

.position {
  font-size: 28px;
  line-height: 1;
  color: var(--ink-100);
  text-align: center;
}

.row.current .position {
  color: var(--lime);
}

.row.mobile .position {
  font-size: 22px;
}

.manager {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.manager-text {
  min-width: 0;
}

.name {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink-100);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.row.mobile .name {
  font-size: 13px;
}

.squad {
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--ink-300);
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.trend {
  text-align: right;
}

.points {
  font-size: 22px;
  line-height: 1;
  color: var(--ink-100);
  text-align: right;
}

.row.mobile .points {
  font-size: 18px;
}
</style>
