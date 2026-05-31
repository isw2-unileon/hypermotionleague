<!--
  SPRINT 2 WIRING — DEV 3
  =======================
  This page is a STATIC SHELL. To wire it:
  1. Fetch current lineup:
       GET /api/v1/leagues/:id/lineups/me  (current matchday)
  2. Save lineup:
       PUT /api/v1/leagues/:id/lineups/me  with body
       { players: [{ playerId, positionSlot, isStarter }] }
       The backend uses ReplaceLineupPlayers (Sprint 1.6 task 3.1)
       so saves are atomic and stale players are dropped.
  3. Deadline: bind countdown to matchday.startDate from
       GET /api/v1/leagues/:id/current-matchday
  4. Substitution modal: list bench players from the same lineup
       response and emit position swaps.
-->
<template>
  <AppLayout>
    <div class="equipo-page">
<!-- ── Header ── -->
      <div class="equipo-header">
        <div>
          <span class="meta-strip">
            <span class="meta-dot"></span>
            MI EQUIPO · J·32
          </span>
          <h1 class="display uc equipo-title">
            {{ selectedLeagueName || 'Mi Equipo' }}
          </h1>
        </div>
        <!-- TODO Sprint 2 (Dev 3): bind countdown to matchday.startDate from
             GET /api/v1/leagues/:id/current-matchday -->
        <span class="tag tag-lime tnum">CIERRA EN 04:18:42</span>
      </div>

      <!-- ── League selector ── -->
      <div class="league-select-wrap">
        <select v-model="selectedLeagueId" @change="onLeagueChange" class="input">
          <option disabled value="">Selecciona una liga</option>
          <option v-for="league in leagues" :key="league.id" :value="league.id">
            {{ league.name }}
          </option>
        </select>
      </div>

      <!-- ── Formation chips ── -->
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

      <!-- ── States ── -->
      <div v-if="loading" class="state-msg mono">Cargando equipo…</div>
      <div v-else-if="error" class="state-msg state-error">{{ error }}</div>
      <div v-else-if="!selectedLeagueId" class="state-msg mono">
        Selecciona una liga para ver tu equipo
      </div>
      <div v-else-if="team && team.players.length === 0" class="state-msg mono">
        Tu plantilla está vacía — ve al mercado
      </div>

      <!-- ── Pitch ── -->
      <div v-else-if="team" class="pitch-wrap">
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
                  <span class="slot-name mono">{{ slot.player.last_name }}</span>
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
                  <span class="slot-name mono">{{ slot.player.last_name }}</span>
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
                  <span class="slot-name mono">{{ slot.player.last_name }}</span>
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
                  <span class="slot-name mono">{{ slot.player.last_name }}</span>
                </template>
                <div v-else class="slot-empty">
                  <span class="pos pos-gk">POR</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- ── Save bar ── -->
        <div class="save-bar">
          <!-- TODO Sprint 2 (Dev 3): wire to PUT /api/v1/leagues/:id/lineups/me -->
          <button
            class="btn btn-primary save-btn"
            :disabled="!hasEdits"
            @click="saveLineup"
          >
            Guardar alineación
          </button>
        </div>
      </div>

      <!-- ── Substitution modal (stub) ── -->
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
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue';
import AppLayout from '@/layouts/AppLayout.vue';
import PitchSVG from '@/design-system/primitives/PitchSVG.vue';
import PlayerPhoto from '@/design-system/primitives/PlayerPhoto.vue';
import api from '@/lib/api';

// ── Types ──────────────────────────────────────────────────────────────────

type PlayerPosition = 'GK' | 'DEF' | 'MID' | 'FWD';

interface League {
  id: number;
  name: string;
}

interface Player {
  id: number;
  first_name: string;
  last_name: string;
  position: PlayerPosition;
  team_name: string;
  market_value: number;
}

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

// ── Formations ─────────────────────────────────────────────────────────────

const FORMATIONS: Record<string, { gk: number; def: number; mid: number; fwd: number }> = {
  '4-4-2': { gk: 1, def: 4, mid: 4, fwd: 2 },
  '4-3-3': { gk: 1, def: 4, mid: 3, fwd: 3 },
  '3-5-2': { gk: 1, def: 3, mid: 5, fwd: 2 },
  '5-3-2': { gk: 1, def: 5, mid: 3, fwd: 2 },
  '3-4-3': { gk: 1, def: 3, mid: 4, fwd: 3 },
};
const formationKeys = Object.keys(FORMATIONS);

// ── State ──────────────────────────────────────────────────────────────────

const leagues = ref<League[]>([]);
const selectedLeagueId = ref<number | ''>('');
const team = ref<UserTeam | null>(null);
const loading = ref(false);
const error = ref('');
const formation = ref('4-3-3');
const hasEdits = ref(false); // TODO Sprint 2 (Dev 3): set true on any slot change

const subModal = reactive<{
  open: boolean;
  position: PlayerPosition;
  index: number;
  current: TeamPlayerWithDetails | null;
}>({ open: false, position: 'FWD', index: 0, current: null });

// ── Derived ────────────────────────────────────────────────────────────────

const selectedLeagueName = computed(
  () => leagues.value.find(l => l.id === selectedLeagueId.value)?.name ?? '',
);

const slottedPlayers = computed(() => {
  const config = FORMATIONS[formation.value];
  const take = (pos: PlayerPosition, n: number): (TeamPlayerWithDetails | null)[] => {
    const pool = team.value?.players.filter(p => p.player.position === pos) ?? [];
    return Array.from({ length: n }, (_, i) => pool[i] ?? null);
  };
  return {
    gk:  take('GK',  config.gk),
    def: take('DEF', config.def),
    mid: take('MID', config.mid),
    fwd: take('FWD', config.fwd),
  };
});

// ── Helpers ────────────────────────────────────────────────────────────────

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
  return TEAM_PALETTE[Math.abs(hash) % TEAM_PALETTE.length];
}

function posLabel(pos: PlayerPosition): string {
  return { GK: 'POR', DEF: 'DEF', MID: 'MED', FWD: 'DEL' }[pos];
}

function posBadgeClass(pos: PlayerPosition): string {
  return { GK: 'pos-gk', DEF: 'pos-def', MID: 'pos-mid', FWD: 'pos-fwd' }[pos];
}

function benchPlayers(pos: PlayerPosition): TeamPlayerWithDetails[] {
  if (!team.value) return [];
  const keyMap = { GK: 'gk', DEF: 'def', MID: 'mid', FWD: 'fwd' } as const;
  const n = FORMATIONS[formation.value][keyMap[pos]];
  return team.value.players.filter(p => p.player.position === pos).slice(n);
}

// ── Actions ────────────────────────────────────────────────────────────────

function selectFormation(f: string) {
  formation.value = f;
}

function openSubModal(slot: TeamPlayerWithDetails | null, pos: PlayerPosition, idx: number) {
  subModal.open = true;
  subModal.position = pos;
  subModal.index = idx;
  subModal.current = slot;
}

function selectBenchPlayer(_p: TeamPlayerWithDetails) {
  // TODO Sprint 2 (Dev 3): swap slot with bench player and set hasEdits = true
  console.log('TODO Sprint 2 (Dev 3): swap slot', subModal.index, 'with', _p.player.last_name);
  subModal.open = false;
}

function saveLineup() {
  // TODO Sprint 2 (Dev 3): wire to PUT /api/v1/leagues/:id/lineups/me
  // Body: { players: [{ playerId, positionSlot, isStarter }] }
  // Backend uses ReplaceLineupPlayers — atomic, stale players are dropped.
  console.log('TODO Sprint 2 (Dev 3): save lineup for league', selectedLeagueId.value);
}

async function onLeagueChange() {
  if (!selectedLeagueId.value) return;
  team.value = null;
  error.value = '';
  loading.value = true;
  try {
    team.value = await api.get<UserTeam>(`/api/v1/leagues/${selectedLeagueId.value}/team`);
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
      selectedLeagueId.value = leagues.value[0].id;
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

/* ── Header ── */
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
  font-size: 9px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--ink-100);
  max-width: 52px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: center;
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
