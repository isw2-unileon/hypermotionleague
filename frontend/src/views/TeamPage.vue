<template>
  <AppShell>
    <div class="equipo-page">
<!-- Header -->
      <div class="equipo-header">
        <div>
          <span class="meta-strip">
            <span class="meta-dot"></span>
            MI EQUIPO · J·{{ currentMatchday?.number ?? '—' }}
          </span>
          <h1 class="display uc equipo-title">
            {{ selectedLeagueName || 'Mi Equipo' }}
          </h1>
        </div>
        <span class="tag tag-lime tnum">CIERRA EN {{ countdown }}</span>
      </div>

      <!-- League selector -->
      <div class="league-select-wrap">
        <select v-model="selectedLeagueId" @change="onLeagueChange" class="input">
          <option disabled value="">Selecciona una liga</option>
          <option v-for="league in leagues" :key="league.id" :value="league.id">
            {{ league.name }}
          </option>
        </select>
      </div>

      <!-- Formation chips  -->
      <div class="formation-chips">
        <button
          v-for="f in formationKeys"
          :key="f"
          class="btn formation-chip"
          :class="formation === f ? 'btn-primary' : 'btn-ghost'"
          @click="selectFormation(f)"
        >
          {{ f }}
        </button>
      </div>

      <!--  States  -->
      <div v-if="loading" class="state-msg mono">Cargando equipo…</div>
      <div v-else-if="error" class="state-msg state-error">{{ error }}</div>
      <div v-else-if="!selectedLeagueId" class="state-msg mono">
        Selecciona una liga para ver tu equipo
      </div>
      <div v-else-if="squadPlayers.length === 0" class="state-msg mono">
        Tu plantilla está vacía — ve al mercado
      </div>

      <!-- Pitch -->
      <div v-else-if="squadPlayers.length > 0" class="pitch-wrap">
        <div class="pitch-container">
          <PitchSVG />

          <div class="pitch-field">
            <!-- FWD row -->
            <div class="position-row">
              <div
                v-for="(slot, i) in slottedPlayers.fwd"
                :key="'fwd-' + i"
                class="player-slot"
                @click="openSubModal(slot, 'FWD', i)"
              >
                <template v-if="slot">
                  <PlayerPhoto
                    :team="shortTeam(slot.player.team_name)"
                    :color="teamColor(slot.player.team_name)"
                    :size="48"
                  />
                  <span class="pos pos-fwd">DEL</span>
                  <span class="slot-name mono">{{ slot.player.first_name.charAt(0) }}. {{ slot.player.last_name }}</span>
                </template>
                <div v-else class="slot-empty">
                  <span class="pos pos-fwd">DEL</span>
                </div>
              </div>
            </div>

            <!-- MID row -->
            <div class="position-row">
              <div
                v-for="(slot, i) in slottedPlayers.mid"
                :key="'mid-' + i"
                class="player-slot"
                @click="openSubModal(slot, 'MID', i)"
              >
                <template v-if="slot">
                  <PlayerPhoto
                    :team="shortTeam(slot.player.team_name)"
                    :color="teamColor(slot.player.team_name)"
                    :size="48"
                  />
                  <span class="pos pos-mid">MED</span>
                  <span class="slot-name mono">{{ slot.player.first_name.charAt(0) }}. {{ slot.player.last_name }}</span>
                </template>
                <div v-else class="slot-empty">
                  <span class="pos pos-mid">MED</span>
                </div>
              </div>
            </div>

            <!-- DEF row -->
            <div class="position-row">
              <div
                v-for="(slot, i) in slottedPlayers.def"
                :key="'def-' + i"
                class="player-slot"
                @click="openSubModal(slot, 'DEF', i)"
              >
                <template v-if="slot">
                  <PlayerPhoto
                    :team="shortTeam(slot.player.team_name)"
                    :color="teamColor(slot.player.team_name)"
                    :size="48"
                  />
                  <span class="pos pos-def">DEF</span>
                  <span class="slot-name mono">{{ slot.player.first_name.charAt(0) }}. {{ slot.player.last_name }}</span>
                </template>
                <div v-else class="slot-empty">
                  <span class="pos pos-def">DEF</span>
                </div>
              </div>
            </div>

            <!-- GK row -->
            <div class="position-row">
              <div
                v-for="(slot, i) in slottedPlayers.gk"
                :key="'gk-' + i"
                class="player-slot"
                @click="openSubModal(slot, 'GK', i)"
              >
                <template v-if="slot">
                  <PlayerPhoto
                    :team="shortTeam(slot.player.team_name)"
                    :color="teamColor(slot.player.team_name)"
                    :size="48"
                  />
                  <span class="pos pos-gk">POR</span>
                  <span class="slot-name mono">{{ slot.player.first_name.charAt(0) }}. {{ slot.player.last_name }}</span>
                </template>
                <div v-else class="slot-empty">
                  <span class="pos pos-gk">POR</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Save bar  -->
        <div class="save-bar">
          <button
            class="btn btn-primary save-btn"
            :disabled="!hasEdits || saving"
            @click="saveLineup"
          >
            {{ saving ? 'Guardando…' : 'Guardar alineación' }}
          </button>
          <p v-if="saveError" class="save-error mono">{{ saveError }}</p>
        </div>
      </div>

      <!--Substitution modal (stub) -->
      <Teleport to="body">
        <div
          v-if="subModal.open"
          class="modal-overlay"
          @click.self="subModal.open = false"
        >
          <div class="modal-panel card">
            <div class="meta-strip" style="margin-bottom: var(--s-4)">
              <span class="meta-dot"></span>
              CAMBIAR JUGADOR
            </div>
            <p class="modal-slot-info mono">
              {{ posLabel(subModal.position) }} · Slot {{ subModal.index + 1 }}
            </p>

            <div class="bench-list">
              <div
                v-for="p in benchPlayers(subModal.position)"
                :key="p.player_id"
                class="bench-row card-flat"
                @click="selectBenchPlayer(p)"
              >
                <span class="pos" :class="posBadgeClass(subModal.position)">
                  {{ posLabel(subModal.position) }}
                </span>
                <span class="bench-name">
                  {{ p.player.first_name }} {{ p.player.last_name }}
                </span>
                <span class="bench-team mono">{{ shortTeam(p.player.team_name) }}</span>
              </div>
              <p v-if="benchPlayers(subModal.position).length === 0" class="no-bench mono">
                Sin suplentes disponibles
              </p>
            </div>

            <button
              class="btn btn-ghost"
              style="width: 100%; margin-top: var(--s-4)"
              @click="subModal.open = false"
            >
              Cancelar
            </button>
          </div>
        </div>
      </Teleport>
</div>
</AppShell>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, reactive, watch } from 'vue';
import PitchSVG from '@/design-system/primitives/PitchSVG.vue';
import PlayerPhoto from '@/design-system/primitives/PlayerPhoto.vue';
import api from '@/lib/api';
import AppShell from '@/design-system/AppShell.vue';

// Types

type PlayerPosition = 'GK' | 'DEF' | 'MID' | 'FWD';

interface League {
  id: number;
  name: string;
}

interface Matchday {
  id: number;
  number: number;
  name: string;
  start_date: string;
  end_date: string;
  is_current: boolean;
}

interface Player {
  id: number;
  first_name: string;
  last_name: string;
  position: PlayerPosition;
  team_name: string;
  market_value: number;
}

interface LineupPlayer {
  id: number;
  lineup_id: number;
  player_id: number;
  position: PlayerPosition;
  is_starter: boolean;
  points: number;
  player: Player;
}

interface LineupWithPlayers {
  id: number;
  league_id: number;
  user_id: number;
  matchday_id: number | null;
  total_points: number;
  players: LineupPlayer[];
}

// Kept for squad fetch (bench source when building a new lineup from scratch)
interface TeamPlayerWithDetails {
  id: number;
  player_id: number;
  purchase_price: number;
  player: Player;
}

interface UserTeam {
  league_id: number;
  budget: number;
  players: TeamPlayerWithDetails[];
  total_value: number;
}

// Formations

const FORMATIONS: Record<string, { gk: number; def: number; mid: number; fwd: number }> = {
  '4-4-2': { gk: 1, def: 4, mid: 4, fwd: 2 },
  '4-3-3': { gk: 1, def: 4, mid: 3, fwd: 3 },
  '3-5-2': { gk: 1, def: 3, mid: 5, fwd: 2 },
  '5-3-2': { gk: 1, def: 5, mid: 3, fwd: 2 },
  '3-4-3': { gk: 1, def: 3, mid: 4, fwd: 3 },
};
const formationKeys = Object.keys(FORMATIONS);

// State

const leagues = ref<League[]>([]);
const selectedLeagueId = ref<number | ''>('');
const currentMatchday = ref<Matchday | null>(null);
const lineup = ref<LineupWithPlayers | null>(null);
const squadPlayers = ref<TeamPlayerWithDetails[]>([]);
const loading = ref(false);
const error = ref('');
const formation = ref('4-3-3');
const hasEdits = ref(false);
const saving = ref(false);
const saveError = ref('');
const countdown = ref('--:--:--');
let countdownTimer: ReturnType<typeof setInterval> | null = null;

const subModal = reactive<{
  open: boolean;
  position: PlayerPosition;
  index: number;
  current: LineupPlayer | null;
}>({ open: false, position: 'FWD', index: 0, current: null });

// Derived

const selectedLeagueName = computed(
  () => leagues.value.find(l => l.id === selectedLeagueId.value)?.name ?? '',
);

const slottedPlayers = computed(() => {
  const config = FORMATIONS[formation.value] ?? { gk: 1, def: 4, mid: 3, fwd: 3 };
  const take = (pos: PlayerPosition, n: number): (LineupPlayer | null)[] => {
    const starters = lineup.value?.players.filter(p => p.is_starter && p.position === pos) ?? [];
    return Array.from({ length: n }, (_, i) => starters[i] ?? null);
  };
  return {
    gk:  take('GK',  config.gk),
    def: take('DEF', config.def),
    mid: take('MID', config.mid),
    fwd: take('FWD', config.fwd),
  };
});

// Helpers

function shortTeam(name: string): string {
  return name.slice(0, 3).toUpperCase();
}

const TEAM_PALETTE = [
  '#3B82F6', '#EF4444', '#F59E0B', '#10B981',
  '#8B5CF6', '#EC4899', '#06B6D4', '#F97316',
];
function teamColor(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) | 0;
  return TEAM_PALETTE[Math.abs(hash) % TEAM_PALETTE.length] ?? '#3B82F6';
}

function posLabel(pos: PlayerPosition): string {
  return { GK: 'POR', DEF: 'DEF', MID: 'MED', FWD: 'DEL' }[pos];
}

function posBadgeClass(pos: PlayerPosition): string {
  return { GK: 'pos-gk', DEF: 'pos-def', MID: 'pos-mid', FWD: 'pos-fwd' }[pos];
}

function benchPlayers(pos: PlayerPosition): LineupPlayer[] {
  // Non-starters already in the lineup for this position
  const nonStarters = lineup.value?.players.filter(p => !p.is_starter && p.position === pos) ?? [];
  if (nonStarters.length > 0) return nonStarters;

  // Fallback: squad players not present in the lineup at all (new lineup from scratch)
  const inLineup = new Set(lineup.value?.players.map(p => p.player_id) ?? []);
  return squadPlayers.value
    .filter(p => p.player.position === pos && !inLineup.has(p.player_id))
    .map(p => ({
      id: 0,
      lineup_id: 0,
      player_id: p.player_id,
      position: pos,
      is_starter: false,
      points: 0,
      player: p.player,
    }));
}

// Countdown

function formatCountdown(ms: number): string {
  if (ms <= 0) return '00:00:00';
  const totalSecs = Math.floor(ms / 1000);
  const h = Math.floor(totalSecs / 3600);
  const m = Math.floor((totalSecs % 3600) / 60);
  const s = totalSecs % 60;
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

function startCountdown() {
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
  if (!currentMatchday.value) { countdown.value = '--:--:--'; return; }
  const deadline = new Date(currentMatchday.value.start_date).getTime();
  const tick = () => {
    const remaining = deadline - Date.now();
    countdown.value = formatCountdown(remaining);
    if (remaining <= 0 && countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
  };
  tick();
  countdownTimer = setInterval(tick, 1000);
}

watch(currentMatchday, () => startCountdown());

onUnmounted(() => { if (countdownTimer) clearInterval(countdownTimer); });

// Actions

function selectFormation(f: string) {
  formation.value = f;
}

function openSubModal(slot: LineupPlayer | null, pos: PlayerPosition, idx: number) {
  subModal.open = true;
  subModal.position = pos;
  subModal.index = idx;
  subModal.current = slot;
}

function selectBenchPlayer(p: LineupPlayer) {
  if (!lineup.value) {
    lineup.value = {
      id: 0,
      league_id: Number(selectedLeagueId.value) || 0,
      user_id: 0,
      matchday_id: currentMatchday.value?.id ?? null,
      total_points: 0,
      players: [],
    };
  }

  const players = lineup.value.players;

  // Move current slot occupant to bench
  if (subModal.current) {
    const cur = players.find(lp => lp.player_id === subModal.current!.player_id);
    if (cur) cur.is_starter = false;
  }

  // Promote incoming player to starter
  const existing = players.find(lp => lp.player_id === p.player_id);
  if (existing) {
    existing.is_starter = true;
  } else {
    players.push({ ...p, is_starter: true });
  }

  hasEdits.value = true;
  subModal.open = false;
}

async function saveLineup() {
  if (!selectedLeagueId.value || !hasEdits.value) return;
  saving.value = true;
  saveError.value = '';
  try {
    const reqPlayers = (lineup.value?.players ?? []).map(p => ({
      player_id: p.player_id,
      position: p.position,
      is_starter: p.is_starter,
    }));

    if (currentMatchday.value) {
      // Save to specific matchday
      await api.put(
        `/api/v1/leagues/${selectedLeagueId.value}/matchdays/${currentMatchday.value.number}/lineup`,
        { matchday_id: currentMatchday.value.id, players: reqPlayers },
      );
      hasEdits.value = false;
      lineup.value = await api.get<LineupWithPlayers>(
        `/api/v1/leagues/${selectedLeagueId.value}/matchdays/${currentMatchday.value.number}/lineup`,
      );
    } else {
      // No matchday — save as default lineup
      await api.put(
        `/api/v1/leagues/${selectedLeagueId.value}/lineup/default`,
        { players: reqPlayers },
      );
      hasEdits.value = false;
      lineup.value = await api.get<LineupWithPlayers>(
        `/api/v1/leagues/${selectedLeagueId.value}/lineup/default`,
      );
    }
  } catch (e) {
    saveError.value = e instanceof Error ? e.message : 'Error al guardar la alineación';
  } finally {
    saving.value = false;
  }
}

async function onLeagueChange() {
  if (!selectedLeagueId.value) return;
  lineup.value = null;
  squadPlayers.value = [];
  currentMatchday.value = null;
  error.value = '';
  loading.value = true;
  try {
    const leagueId = selectedLeagueId.value;

    // 1. Fetch current matchday
    const matchdaysData = await api.get<{ matchdays: Matchday[] }>(`/api/v1/leagues/${leagueId}/matchdays`);
    const matchdays = matchdaysData.matchdays ?? [];
    currentMatchday.value = matchdays.find(m => m.is_current) ?? matchdays[0] ?? null;

    // 2. Fetch squad (needed for bench when building a lineup from scratch)
    const teamData = await api.get<UserTeam>(`/api/v1/leagues/${leagueId}/team`);
    squadPlayers.value = teamData.players ?? [];

    // 3. Try to fetch existing lineup
    if (currentMatchday.value) {
      // Matchday exists — try matchday-specific lineup
      try {
        lineup.value = await api.get<LineupWithPlayers>(
          `/api/v1/leagues/${leagueId}/matchdays/${currentMatchday.value.number}/lineup`,
        );
      } catch {
        lineup.value = null;
      }
    }
    // Fallback: try default lineup if no matchday or no matchday-specific lineup
    if (!lineup.value) {
      try {
        lineup.value = await api.get<LineupWithPlayers>(
          `/api/v1/leagues/${leagueId}/lineup/default`,
        );
      } catch {
        lineup.value = null;
      }
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Error al cargar el equipo';
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  try {
    leagues.value = await api.get<League[]>('/api/v1/leagues');
    if (leagues.value.length === 1) {
      selectedLeagueId.value = leagues.value[0]!.id;
      await onLeagueChange();
    }
  } catch {
    error.value = 'No se pudieron cargar las ligas';
  }
});
</script>

<style scoped>
/* ── Page shell ── */
.equipo-page {
  padding: var(--s-4) var(--s-4) var(--s-16);
  min-height: 100vh;
  background: var(--ink-900);
  color: var(--ink-100);
}

/*Header */
.equipo-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--s-4);
  padding-top: var(--s-4);
}

.equipo-title {
  font-family: var(--f-display);
  font-size: 32px;
  line-height: 1;
  color: var(--ink-50);
  margin: var(--s-1) 0 0;
}

/* ── League selector ── */
.league-select-wrap {
  margin-bottom: var(--s-4);
}

/* ── Formation chips ── */
.formation-chips {
  display: flex;
  gap: var(--s-2);
  flex-wrap: wrap;
  margin-bottom: var(--s-5);
}

.formation-chip {
  font-family: var(--f-mono);
  font-size: 12px;
  letter-spacing: 0.06em;
  padding: 6px 12px;
  border-radius: var(--r-sm);
}

/* ── State messages ── */
.state-msg {
  text-align: center;
  color: var(--ink-400);
  padding: var(--s-12) 0;
  font-size: 13px;
  letter-spacing: 0.04em;
}

.state-error {
  color: var(--down);
}

/* ── Pitch wrapper ── */
.pitch-wrap {
  display: flex;
  flex-direction: column;
  gap: var(--s-4);
}

/* ── Pitch container ── */
.pitch-container {
  position: relative;
  width: 100%;
  aspect-ratio: 3 / 4;
  background: #1a2e1a;
  border-radius: var(--r-lg);
  overflow: hidden;
  border: 1px solid var(--ink-700);
}

/* ── Pitch field overlay ── */
.pitch-field {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column-reverse; /* GK at bottom */
  padding: var(--s-4) var(--s-2);
  z-index: 1;
}

/* ── Position rows ── */
.position-row {
  display: flex;
  justify-content: space-around;
  align-items: center;
  flex: 1;
}

/* ── Player slot ── */
.player-slot {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  cursor: pointer;
  padding: 4px;
  border-radius: var(--r-md);
  transition: background 0.15s;
}

.player-slot:hover {
  background: rgba(199, 255, 61, 0.08);
}

.slot-name {
  font-size: 10px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--ink-100);
  max-width: 70px;
  text-align: center;
  line-height: 1.2;
  word-break: break-word;
}

/* ── Empty slot ── */
.slot-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: var(--r-md);
  border: 1px dashed var(--ink-600);
  background: rgba(255, 255, 255, 0.04);
}

/* ── Save bar ── */
.save-bar {
  padding: var(--s-2) 0;
}

.save-btn {
  width: 100%;
  padding: 14px;
  font-size: 14px;
}

.save-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.save-error {
  margin-top: var(--s-2);
  font-size: 11px;
  letter-spacing: 0.06em;
  color: var(--down);
  text-align: center;
}

/* ── Modal overlay ── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(8, 9, 11, 0.75);
  display: flex;
  align-items: flex-end;
  z-index: 50;
}

.modal-panel {
  width: 100%;
  max-height: 70vh;
  border-radius: var(--r-xl) var(--r-xl) 0 0;
  padding: var(--s-6) var(--s-4) var(--s-8);
  overflow-y: auto;
}

.modal-slot-info {
  font-size: 11px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--ink-300);
  margin: 0 0 var(--s-4);
}

/* ── Bench list ── */
.bench-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.bench-row {
  display: flex;
  align-items: center;
  gap: var(--s-3);
  padding: var(--s-3);
  cursor: pointer;
  border-radius: var(--r-sm);
  transition: background 0.12s;
}

.bench-row:hover {
  background: var(--ink-700);
}

.bench-name {
  flex: 1;
  font-size: 13px;
  color: var(--ink-100);
}

.bench-team {
  font-size: 10px;
  letter-spacing: 0.1em;
  color: var(--ink-400);
  text-transform: uppercase;
}

.no-bench {
  font-size: 11px;
  letter-spacing: 0.08em;
  color: var(--ink-400);
  text-align: center;
  padding: var(--s-4) 0;
}

/* ── Desktop enhancement ── */
@media (min-width: 768px) {
  .equipo-page {
    max-width: 480px;
    margin: 0 auto;
    padding-top: var(--s-8);
  }

  .equipo-title {
    font-size: 42px;
  }

  .pitch-container {
    aspect-ratio: 2 / 3;
  }
}
</style>
