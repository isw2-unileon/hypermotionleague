<template>
  <AppShell>
    <div class="lineup">

      <!-- Header -->
      <header class="header">
        <button type="button" class="back mono" @click="router.back()">← Volver</button>
        <h1 class="title display">Alineación</h1>
      </header>

      <!-- Matchday info -->
      <div v-if="matchday" class="card matchday-card">
        <p class="matchday-name">Jornada {{ matchday.number }} — {{ matchday.name }}</p>
        <p class="matchday-close mono">Cierre: {{ formatDate(matchday.end_date) }}</p>
      </div>

      <!-- Formation selector -->
      <div class="section">
        <p class="label">Formación</p>
        <div class="formation-row">
          <button
            v-for="f in formationKeys"
            :key="f"
            type="button"
            class="btn formation-btn"
            :class="formation === f ? 'btn-primary' : 'btn-ghost'"
            @click="selectFormation(f)"
          >
            {{ f }}
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="state state-muted">Cargando...</div>

      <!-- Error -->
      <div v-else-if="error" class="state state-error">{{ error }}</div>

      <!-- No players -->
      <div v-else-if="players.length === 0" class="state state-muted">
        <p class="empty-icon">👕</p>
        <p class="empty-title">No tienes jugadores en esta liga</p>
        <RouterLink to="/market" class="empty-link">Ir al mercado</RouterLink>
      </div>

      <template v-else>
        <!-- Position slot progress -->
        <div class="slot-row">
          <div
            v-for="pos in POSITIONS"
            :key="pos"
            class="slot-cell"
            :class="slotsFilled(pos) === formationSlots(pos) ? 'slot-full' : ''"
          >
            <span class="pos" :class="posClass(pos)">{{ pos }}</span>
            <span class="slot-count mono">{{ slotsFilled(pos) }}/{{ formationSlots(pos) }}</span>
          </div>
        </div>

        <!-- Save button -->
        <button
          type="button"
          class="btn save-btn"
          :class="isLineupValid ? 'btn-primary' : 'btn-disabled'"
          :disabled="!isLineupValid || saving"
          @click="saveLineup"
        >
          {{
            saving
              ? 'Guardando...'
              : isLineupValid
              ? 'Guardar alineación'
              : `Faltan ${11 - startersCount} titulares`
          }}
        </button>

        <!-- Save feedback -->
        <div v-if="saveError" class="feedback feedback-error">{{ saveError }}</div>
        <div v-if="saveSuccess" class="feedback feedback-ok">¡Alineación guardada correctamente!</div>

        <!-- Players grouped by position -->
        <div v-for="pos in POSITIONS" :key="pos" class="pos-group">
          <p class="pos-group-label label">{{ positionLabel(pos) }}</p>
          <div v-if="playersByPosition(pos).length > 0" class="player-list">
            <div
              v-for="player in playersByPosition(pos)"
              :key="player.player_id"
              class="player-row"
              :class="[
                starterMap[player.player_id] ? 'player-row-active' : '',
                !starterMap[player.player_id] && !canAddStarter(pos) ? 'player-row-disabled' : '',
              ]"
              @click="toggleStarter(player)"
            >
              <!-- Checkbox -->
              <div
                class="checkbox"
                :class="starterMap[player.player_id] ? 'checkbox-on' : ''"
              >
                <span v-if="starterMap[player.player_id]" class="checkbox-tick">✓</span>
              </div>

              <!-- Info -->
              <div class="player-info">
                <p class="player-name">
                  {{ player.player.first_name }} {{ player.player.last_name }}
                </p>
                <p class="player-team mono">{{ player.player.team_name }}</p>
              </div>

              <!-- Badge -->
              <span class="tag" :class="starterMap[player.player_id] ? 'tag-lime' : ''">
                {{ starterMap[player.player_id] ? 'Titular' : 'Suplente' }}
              </span>
            </div>
          </div>
          <p v-else class="no-players">No tienes jugadores en esta posición</p>
        </div>
      </template>

    </div>
  </AppShell>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter, RouterLink } from 'vue-router';
import AppShell from '@/design-system/AppShell.vue';
import api from '@/lib/api';

type PlayerPosition = 'GK' | 'DEF' | 'MID' | 'FWD';

const POSITIONS: PlayerPosition[] = ['GK', 'DEF', 'MID', 'FWD'];

const FORMATIONS: Record<string, Record<PlayerPosition, number>> = {
  '4-3-3': { GK: 1, DEF: 4, MID: 3, FWD: 3 },
  '4-4-2': { GK: 1, DEF: 4, MID: 4, FWD: 2 },
  '3-5-2': { GK: 1, DEF: 3, MID: 5, FWD: 2 },
  '4-5-1': { GK: 1, DEF: 4, MID: 5, FWD: 1 },
  '5-3-2': { GK: 1, DEF: 5, MID: 3, FWD: 2 },
};

const formationKeys = Object.keys(FORMATIONS);

interface Player {
  id: number;
  first_name: string;
  last_name: string;
  position: PlayerPosition;
  team_name: string;
}

interface TeamPlayer {
  player_id: number;
  player: Player;
}

interface Matchday {
  id: number;
  number: number;
  name: string;
  end_date: string;
  is_current: boolean;
}

interface LineupPlayerDetail {
  player_id: number;
  is_starter: boolean;
  position: PlayerPosition;
}

interface LineupWithPlayers {
  id: number;
  matchday_id: number;
  players: LineupPlayerDetail[];
}

const route = useRoute();
const router = useRouter();

const leagueId = route.params.leagueId as string;
const matchdayNumber = route.params.matchdayNumber as string;

const players = ref<TeamPlayer[]>([]);
const matchday = ref<Matchday | null>(null);
const formation = ref('4-3-3');
const starterMap = ref<Record<number, boolean>>({});
const loading = ref(true);
const error = ref('');
const saving = ref(false);
const saveError = ref('');
const saveSuccess = ref(false);

function positionLabel(pos: PlayerPosition): string {
  return { GK: 'Porteros', DEF: 'Defensas', MID: 'Centrocampistas', FWD: 'Delanteros' }[pos];
}

function posClass(pos: PlayerPosition): string {
  return { GK: 'pos-gk', DEF: 'pos-def', MID: 'pos-mid', FWD: 'pos-fwd' }[pos];
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('es-ES', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function playersByPosition(pos: PlayerPosition): TeamPlayer[] {
  return players.value.filter(p => p.player.position === pos);
}

function formationSlots(pos: PlayerPosition): number {
  return FORMATIONS[formation.value]?.[pos] ?? 0;
}

function slotsFilled(pos: PlayerPosition): number {
  return playersByPosition(pos).filter(p => starterMap.value[p.player_id]).length;
}

function canAddStarter(pos: PlayerPosition): boolean {
  return slotsFilled(pos) < formationSlots(pos);
}

const startersCount = computed(() =>
  Object.values(starterMap.value).filter(Boolean).length
);

const isLineupValid = computed(() => {
  if (startersCount.value !== 11) return false;
  return POSITIONS.every(pos => slotsFilled(pos) === formationSlots(pos));
});

function selectFormation(f: string) {
  formation.value = f;
  for (const pos of POSITIONS) {
    const max = FORMATIONS[f]?.[pos] ?? 0;
    const starters = playersByPosition(pos).filter(p => starterMap.value[p.player_id]);
    starters.slice(max).forEach(p => {
      starterMap.value[p.player_id] = false;
    });
  }
}

function toggleStarter(player: TeamPlayer) {
  const pos = player.player.position;
  if (starterMap.value[player.player_id]) {
    starterMap.value[player.player_id] = false;
    return;
  }
  if (!canAddStarter(pos)) return;
  starterMap.value[player.player_id] = true;
}

async function saveLineup() {
  if (!isLineupValid.value || !matchday.value) return;
  saving.value = true;
  saveError.value = '';
  saveSuccess.value = false;

  const lineupPlayers = players.value.map(p => ({
    player_id: p.player_id,
    position: p.player.position,
    is_starter: starterMap.value[p.player_id] ?? false,
  }));

  try {
    await api.put(`/api/v1/leagues/${leagueId}/matchdays/${matchdayNumber}/lineup`, {
      matchday_id: matchday.value.id,
      players: lineupPlayers,
    });
    saveSuccess.value = true;
    setTimeout(() => { saveSuccess.value = false; }, 3000);
  } catch (e) {
    saveError.value = e instanceof Error ? e.message : 'Error al guardar la alineación';
  } finally {
    saving.value = false;
  }
}

onMounted(async () => {
  loading.value = true;
  try {
    const [teamData, matchdaysData] = await Promise.all([
      api.get<{ players: TeamPlayer[] }>(`/api/v1/leagues/${leagueId}/team`),
      api.get<{ matchdays: Matchday[] }>(`/api/v1/leagues/${leagueId}/matchdays`),
    ]);

    players.value = teamData.players ?? [];
    const matchdays = matchdaysData.matchdays ?? [];
    matchday.value = matchdays.find(m => m.number === parseInt(matchdayNumber)) ?? null;

    try {
      const existing = await api.get<LineupWithPlayers>(
        `/api/v1/leagues/${leagueId}/matchdays/${matchdayNumber}/lineup`
      );
      if (existing?.players) {
        const map: Record<number, boolean> = {};
        existing.players.forEach(p => { map[p.player_id] = p.is_starter; });
        starterMap.value = map;
      }
    } catch {
      // No lineup yet — start from scratch
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Error al cargar los datos';
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.lineup {
  display: flex;
  flex-direction: column;
  gap: 20px;
  color: var(--ink-100);
}

/* Header */
.header {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.back {
  align-self: flex-start;
  background: transparent;
  border: none;
  padding: 0;
  color: var(--ink-300);
  font-size: 11px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  cursor: pointer;
  transition: color 0.15s;
}

.back:hover { color: var(--lime); }

.title {
  font-size: 56px;
  line-height: 0.9;
  text-transform: uppercase;
  margin: 0;
}

/* Matchday card */
.matchday-card {
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.matchday-name {
  font-size: 14px;
  font-weight: 600;
  margin: 0;
}

.matchday-close {
  font-size: 11px;
  color: var(--ink-400);
  letter-spacing: 0.06em;
  margin: 0;
}

/* Formation */
.section { display: flex; flex-direction: column; gap: 8px; }

.formation-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.formation-btn {
  padding: 8px 14px;
  font-size: 13px;
}

/* Slot progress */
.slot-row {
  display: flex;
  gap: 8px;
}

.slot-cell {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 10px 6px;
  border-radius: var(--r-sm);
  border: 1px solid var(--ink-700);
  background: var(--ink-800);
  transition: border-color 0.15s, background 0.15s;
}

.slot-full {
  border-color: rgba(199, 255, 61, 0.35);
  background: rgba(199, 255, 61, 0.06);
}

.slot-count {
  font-size: 12px;
  color: var(--ink-200);
}

/* Save button */
.save-btn {
  width: 100%;
  padding: 14px;
  font-size: 14px;
}

.btn-disabled {
  background: var(--ink-700);
  color: var(--ink-400);
  border: 1px solid var(--ink-600);
  cursor: not-allowed;
}

/* Feedback banners */
.feedback {
  padding: 12px 16px;
  border-radius: var(--r-sm);
  font-size: 13px;
}

.feedback-error {
  background: rgba(255, 98, 98, 0.12);
  border: 1px solid rgba(255, 98, 98, 0.3);
  color: var(--down);
}

.feedback-ok {
  background: rgba(91, 227, 138, 0.1);
  border: 1px solid rgba(91, 227, 138, 0.25);
  color: var(--up);
}

/* States */
.state {
  text-align: center;
  padding: 48px 0;
  font-size: 14px;
}

.state-muted { color: var(--ink-400); }
.state-error { color: var(--down); }

.empty-icon { font-size: 40px; margin: 0 0 12px; }
.empty-title { font-size: 15px; font-weight: 600; margin: 0 0 8px; color: var(--ink-100); }
.empty-link { color: var(--lime); font-size: 13px; }

/* Position groups */
.pos-group { display: flex; flex-direction: column; gap: 6px; }

.pos-group-label {
  margin: 0;
}

.player-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.player-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--ink-800);
  border: 1px solid var(--ink-700);
  border-radius: var(--r-sm);
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.player-row:hover { background: var(--ink-700); }

.player-row-active {
  background: rgba(199, 255, 61, 0.07);
  border-color: rgba(199, 255, 61, 0.25);
}

.player-row-active:hover {
  background: rgba(199, 255, 61, 0.1);
}

.player-row-disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* Checkbox */
.checkbox {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid var(--ink-500);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: border-color 0.15s, background 0.15s;
}

.checkbox-on {
  border-color: var(--lime);
  background: var(--lime);
}

.checkbox-tick {
  font-size: 11px;
  color: var(--ink-900);
  line-height: 1;
  font-weight: 700;
}

/* Player info */
.player-info {
  flex: 1;
  min-width: 0;
}

.player-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink-100);
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.player-team {
  font-size: 10px;
  color: var(--ink-400);
  letter-spacing: 0.08em;
  margin: 2px 0 0;
}

.no-players {
  font-size: 12px;
  color: var(--ink-500);
  margin: 0;
  padding: 4px 2px;
}

@media (max-width: 767px) {
  .title { font-size: 36px; }
  .slot-row { flex-wrap: wrap; }
  .slot-cell { min-width: calc(50% - 4px); }
}
</style>
