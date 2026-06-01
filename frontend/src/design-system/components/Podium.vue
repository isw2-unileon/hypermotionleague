<script setup lang="ts">
import { computed } from 'vue'
import type { StandingsRow } from '@/types/standings'
import Avatar from '../primitives/Avatar.vue'

interface Props {
  /** Exactly the top three rows, ordered 1st, 2nd, 3rd. */
  top3: StandingsRow[]
  mobile?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  mobile: false,
})

const emit = defineEmits<{ (e: 'click', userId: number): void }>()

interface PodiumSlot {
  row: StandingsRow
  place: number
  emphasis: boolean
}

// Visual order places 1st in the center, 2nd on the left, 3rd on the right —
// the real-podium arrangement for both layouts.
const slots = computed<PodiumSlot[]>(() => {
  const order = [
    { index: 1, place: 2, emphasis: false },
    { index: 0, place: 1, emphasis: true },
    { index: 2, place: 3, emphasis: false },
  ]
  return order
    .map(({ index, place, emphasis }) => {
      const row = props.top3[index]
      return row ? { row, place, emphasis } : null
    })
    .filter((slot): slot is PodiumSlot => slot !== null)
})

const MEDALS: Record<number, string> = {
  1: '#E5C547',
  2: '#C0C5CE',
  3: '#CD7F32',
}

function medalColor(place: number): string {
  return MEDALS[place] ?? 'var(--ink-400)'
}

function barHeight(place: number): number {
  if (place === 1) return 70
  if (place === 2) return 50
  return 40
}

function barStyle(place: number): Record<string, string> {
  const color = medalColor(place)
  return {
    height: `${barHeight(place)}px`,
    background: `linear-gradient(180deg, ${color}40 0%, ${color}10 100%)`,
    borderTop: `2px solid ${color}`,
  }
}

function formatPoints(points: number): string {
  return points.toLocaleString('es')
}

function firstName(name: string): string {
  return name.split(' ')[0] ?? name
}
</script>

<template>
  <!-- Mobile: bar-chart podium -->
  <div v-if="mobile" class="podium podium-mobile">
    <div
      v-for="slot in slots"
      :key="slot.row.userId"
      class="bar-col"
      role="button"
      tabindex="0"
      @click="emit('click', slot.row.userId)"
      @keydown.enter="emit('click', slot.row.userId)"
    >
      <Avatar
        :initials="slot.row.initials"
        :size="36"
        :color="medalColor(slot.place)"
      />
      <div class="bar-name">{{ firstName(slot.row.name) }}</div>
      <div class="bar-points display tnum" :style="{ color: medalColor(slot.place) }">
        {{ formatPoints(slot.row.totalPoints) }}
      </div>
      <div class="bar" :style="barStyle(slot.place)">
        <span class="bar-place display" :style="{ color: medalColor(slot.place) }">
          {{ slot.place }}
        </span>
      </div>
    </div>
  </div>

  <!-- Desktop: hero strip of three cards -->
  <div v-else class="podium podium-desktop">
    <div
      v-for="slot in slots"
      :key="slot.row.userId"
      class="card podium-card"
      :class="{ emphasis: slot.emphasis }"
      :style="{ borderTopColor: medalColor(slot.place) }"
      role="button"
      tabindex="0"
      @click="emit('click', slot.row.userId)"
      @keydown.enter="emit('click', slot.row.userId)"
    >
      <div class="card-place display" :style="{ color: medalColor(slot.place) }">
        {{ slot.place }}
      </div>
      <Avatar
        :initials="slot.row.initials"
        :size="48"
        :color="medalColor(slot.place)"
      />
      <div class="card-text">
        <div class="card-name">{{ slot.row.name }}</div>
        <div class="card-squad mono">{{ slot.row.squadName }}</div>
      </div>
      <div class="card-points display tnum" :style="{ color: medalColor(slot.place) }">
        {{ formatPoints(slot.row.totalPoints) }}
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Mobile bar chart */
.podium-mobile {
  display: flex;
  justify-content: space-around;
  align-items: flex-end;
  gap: 8px;
}

.bar-col {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  cursor: pointer;
}

.bar-name {
  font-size: 11px;
  font-weight: 600;
  color: var(--ink-100);
  margin-top: 6px;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bar-points {
  font-size: 14px;
  line-height: 1;
  margin-top: 2px;
}

.bar {
  width: 100%;
  margin-top: 6px;
  border-radius: 3px 3px 0 0;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 6px;
}

.bar-place {
  font-size: 18px;
  font-weight: 400;
}

/* Desktop hero strip */
.podium-desktop {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
  align-items: center;
}

.podium-card {
  border-top: 3px solid var(--ink-700);
  padding: 18px;
  display: flex;
  align-items: center;
  gap: 16px;
  cursor: pointer;
  transition: transform 0.15s, box-shadow 0.15s;
}

.podium-card.emphasis {
  transform: translateY(-8px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.35);
}

.card-place {
  font-size: 56px;
  line-height: 1;
  width: 56px;
  text-align: center;
  flex-shrink: 0;
}

.card-text {
  flex: 1;
  min-width: 0;
}

.card-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--ink-100);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-squad {
  font-size: 10px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--ink-400);
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-points {
  font-size: 28px;
  line-height: 1;
  text-align: right;
  flex-shrink: 0;
}
</style>
