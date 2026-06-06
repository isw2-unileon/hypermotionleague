<template>
  <AppShell>
    <div class="squad">
      <PitchSVG class="pitch-bg" />

      <div class="content">
        <!-- Header -->
        <header class="header">
          <button type="button" class="back mono" @click="goBack">← Volver</button>
          <div class="meta mono">{{ matchdayLabel }}</div>
          <h1 class="title display">{{ userName }}</h1>
          <div v-if="squadName" class="subtitle">{{ squadName }}</div>
        </header>

        <!-- View toggle: standings vs full squad -->
        <div class="view-tabs">
          <button
            class="vtab mono"
            :class="{ active: viewMode === 'standings' }"
            @click="viewMode = 'standings'"
          >PUNTOS POR JORNADA</button>
          <button
            class="vtab mono"
            :class="{ active: viewMode === 'squad' }"
            @click="viewMode = 'squad'; fetchRivalSquad()"
          >PLANTILLA COMPLETA</button>
        </div>

        <!-- ═══ STANDINGS VIEW ═══ -->
        <template v-if="viewMode === 'standings'">
          <!-- Matchday selector -->
          <select
            v-model="selectedMatchdayNumber"
            class="input md-select"
            @change="onMatchdayChange"
          >
            <option :value="null">General (Total)</option>
            <option v-for="md in matchdays" :key="md.id" :value="md.number">
              Jornada {{ md.number }} — {{ md.name }}
            </option>
          </select>

          <!-- Loading -->
          <div v-if="loading" class="state state-muted">Cargando equipo...</div>

          <!-- Error -->
          <div v-else-if="error" class="state state-error">{{ error }}</div>

          <!-- User summary -->
          <div v-else-if="userStanding && selectedMatchdayNumber" class="card summary">
            <div class="summary-block">
              <div class="summary-label mono">POSICIÓN</div>
              <div class="summary-value display tnum" :style="{ color: rankColor }">
                #{{ userStanding.rank }}
              </div>
            </div>
            <div class="summary-block summary-right">
              <div class="summary-label mono">PUNTOS</div>
              <div class="summary-value display tnum summary-points">
                {{ userStanding.total_points }}
              </div>
            </div>
          </div>

          <!-- Empty states -->
          <div v-if="!loading && matchdays.length === 0" class="state state-muted">
            No hay jornadas creadas en esta liga todavía.
          </div>
          <div v-else-if="!selectedMatchdayNumber" class="state state-muted">
            Selecciona una jornada para ver los jugadores
          </div>
          <div v-else-if="!userStanding && !loading" class="state state-muted">
            No hay datos para esta jornada
          </div>

          <!-- Players grouped by position -->
          <div v-if="userStanding && userStanding.players && userStanding.players.length > 0" class="groups">
            <section v-for="group in playerGroups" :key="group.key" class="group">
              <div v-for="player in group.players" :key="player.player_id" class="player-row">
                <span class="pos" :class="group.cls">{{ group.label }}</span>
                <div class="player-info">
                  <div class="player-name">{{ player.first_name }} {{ player.last_name }}</div>
                  <div class="player-team mono">{{ player.team_name }}</div>
                </div>
                <span class="player-points display tnum">{{ player.points }}</span>
              </div>
            </section>
          </div>

          <div v-else-if="userStanding && selectedMatchdayNumber && !loading" class="state state-muted">
            Este usuario no tiene jugadores en su alineación para esta jornada.
          </div>
        </template>

        <!-- ═══ SQUAD VIEW (with buy clause) ═══ -->
        <template v-if="viewMode === 'squad'">
          <div v-if="squadLoading" class="state state-muted">Cargando plantilla...</div>
          <div v-else-if="squadError" class="state state-error">{{ squadError }}</div>
          <div v-else-if="rivalPlayers.length === 0" class="state state-muted">
            Este usuario no tiene jugadores en su equipo.
          </div>
          <div v-else class="groups">
            <section v-for="group in rivalPlayerGroups" :key="group.key" class="group">
              <div v-for="tp in group.players" :key="tp.player_id" class="player-row player-row-buy">
                <span class="pos" :class="group.cls">{{ group.label }}</span>
                <div class="player-info">
                  <div class="player-name">{{ tp.player.first_name }} {{ tp.player.last_name }}</div>
                  <div class="player-team mono">{{ tp.player.team_name }}</div>
                </div>
                <div class="clause-col">
                  <div class="mono clause-label">CLÁUSULA</div>
                  <div class="display tnum clause-value">{{ millions(tp.player.market_value) }}</div>
                </div>
                <button
                  class="btn btn-secondary buy-btn"
                  :disabled="buying === tp.player_id"
                  @click="confirmBuy(tp)"
                >
                  {{ buying === tp.player_id ? 'Comprando…' : 'Fichar' }}
                </button>
              </div>
            </section>
          </div>
        </template>

        <!-- Buy result message -->
        <div v-if="buySuccess" class="buy-msg buy-msg-ok mono">{{ buySuccess }}</div>
        <div v-if="buyError" class="buy-msg buy-msg-err mono">{{ buyError }}</div>
      </div>
    </div>
  </AppShell>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import AppShell from "@/design-system/AppShell.vue";
import PitchSVG from "@/design-system/primitives/PitchSVG.vue";
import api from "@/lib/api";

interface Matchday {
  id: number;
  number: number;
  name: string;
}

interface StandingPlayer {
  player_id: number;
  first_name: string;
  last_name: string;
  position: string;
  team_name: string;
  points: number;
}

interface UserStanding {
  rank: number;
  user_id: number;
  username: string;
  display_name: string;
  total_points: number;
  players: StandingPlayer[];
}

interface Standings {
  league_id: number;
  matchday_id?: number;
  rankings: UserStanding[];
}

interface Player {
  id: number;
  first_name: string;
  last_name: string;
  position: string;
  team_name: string;
  market_value: number;
}

interface TeamPlayerWithDetails {
  id: number;
  league_id: number;
  user_id: number;
  player_id: number;
  purchase_price: number;
  player: Player;
}

interface UserTeam {
  league_id: number;
  user_id: number;
  username: string;
  display_name: string;
  budget: number;
  players: TeamPlayerWithDetails[];
  total_value: number;
}

const route = useRoute();
const router = useRouter();

const leagueId = computed(() => Number(route.params.leagueId));
const userId = computed(() => Number(route.params.userId));

const matchdays = ref<Matchday[]>([]);
const selectedMatchdayNumber = ref<number | null>(null);
const userStanding = ref<UserStanding | null>(null);
const userName = ref("");
const loading = ref(false);
const error = ref("");

const viewMode = ref<"standings" | "squad">("standings");
const rivalPlayers = ref<TeamPlayerWithDetails[]>([]);
const squadLoading = ref(false);
const squadError = ref("");
const buying = ref<number | null>(null);
const buySuccess = ref("");
const buyError = ref("");

// Position groups
const POSITION_META = [
  { key: "GK", label: "POR", cls: "pos-gk" },
  { key: "DEF", label: "DEF", cls: "pos-def" },
  { key: "MID", label: "MED", cls: "pos-mid" },
  { key: "FWD", label: "DEL", cls: "pos-fwd" },
] as const;

interface PositionGroup<T> {
  key: string;
  label: string;
  cls: string;
  players: T[];
}

const matchdayLabel = computed(() =>
  selectedMatchdayNumber.value !== null
    ? `EQUIPO · J·${selectedMatchdayNumber.value}`
    : "EQUIPO",
);

const squadName = computed(() => {
  const username = userStanding.value?.username;
  return username ? `@${username}` : "";
});

const rankColor = computed(() => {
  const rank = userStanding.value?.rank;
  if (rank === undefined) return "var(--ink-100)";
  if (rank === 1) return "var(--lime)";
  if (rank <= 3) return "var(--lime-soft)";
  return "var(--ink-100)";
});

const playerGroups = computed<PositionGroup<StandingPlayer>[]>(() => {
  const players = userStanding.value?.players ?? [];
  const known = new Set<string>(POSITION_META.map((m) => m.key));
  const groups: PositionGroup<StandingPlayer>[] = POSITION_META.map((m) => ({
    key: m.key, label: m.label, cls: m.cls,
    players: players.filter((p) => p.position === m.key),
  })).filter((g) => g.players.length > 0);
  const others = players.filter((p) => !known.has(p.position));
  if (others.length > 0) {
    groups.push({ key: "OTHER", label: "—", cls: "pos-other", players: others });
  }
  return groups;
});

const rivalPlayerGroups = computed<PositionGroup<TeamPlayerWithDetails>[]>(() => {
  const players = rivalPlayers.value;
  return POSITION_META.map((m) => ({
    key: m.key, label: m.label, cls: m.cls,
    players: players.filter((p) => p.player.position === m.key),
  })).filter((g) => g.players.length > 0);
});

function millions(amount: number): string {
  const text = (amount / 1_000_000).toFixed(1).replace(/\.0$/, "");
  return `€${text}M`;
}

function goBack() {
  if (window.history.length > 1) {
    router.back();
  } else {
    router.push({ path: "/standings", query: route.query });
  }
}

// ── Data fetching ──

onMounted(async () => {
  await Promise.all([fetchUserName(), fetchMatchdays()]);
  const firstMatchday = matchdays.value[0];
  if (firstMatchday) {
    selectedMatchdayNumber.value = firstMatchday.number;
    await fetchUserStanding();
  }
});

watch(leagueId, async () => {
  await Promise.all([fetchUserName(), fetchMatchdays()]);
  const firstMatchday = matchdays.value[0];
  if (firstMatchday) {
    selectedMatchdayNumber.value = firstMatchday.number;
    await fetchUserStanding();
  }
});

async function fetchUserName() {
  try {
    const team = await api.get<{ username: string; display_name: string }>(
      `/api/v1/leagues/${leagueId.value}/users/${userId.value}/team`,
    );
    userName.value = team.display_name || team.username || "Usuario";
  } catch {
    userName.value = "Usuario";
  }
}

async function onMatchdayChange() {
  await fetchUserStanding();
}

async function fetchMatchdays() {
  try {
    const data = await api.get<{ matchdays: Matchday[] }>(
      `/api/v1/leagues/${leagueId.value}/matchdays`,
    );
    matchdays.value = data.matchdays ?? [];
  } catch {
    // non-blocking
  }
}

async function fetchUserStanding() {
  if (!selectedMatchdayNumber.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const path = `/api/v1/leagues/${leagueId.value}/matchdays/${selectedMatchdayNumber.value}/standings`;
    const data = await api.get<Standings>(path);
    const userRank = data.rankings.find((r) => r.user_id === userId.value);
    if (userRank) {
      userStanding.value = userRank;
    } else {
      error.value = "Usuario no encontrado en la clasificación";
      userStanding.value = null;
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "Error al cargar el equipo";
    userStanding.value = null;
  } finally {
    loading.value = false;
  }
}

async function fetchRivalSquad() {
  if (rivalPlayers.value.length > 0) return; // already loaded
  squadLoading.value = true;
  squadError.value = "";
  try {
    const team = await api.get<UserTeam>(
      `/api/v1/leagues/${leagueId.value}/users/${userId.value}/team`,
    );
    rivalPlayers.value = team.players ?? [];
    if (!userName.value || userName.value === "Usuario") {
      userName.value = team.display_name || team.username || "Usuario";
    }
  } catch (e: unknown) {
    squadError.value = e instanceof Error ? e.message : "Error al cargar la plantilla";
  } finally {
    squadLoading.value = false;
  }
}

async function confirmBuy(tp: TeamPlayerWithDetails) {
  const name = `${tp.player.first_name} ${tp.player.last_name}`;
  const price = millions(tp.player.market_value);
  const ok = window.confirm(
    `¿Pagar la cláusula de ${price} por ${name}? Se descontará de tu saldo.`,
  );
  if (!ok) return;

  buying.value = tp.player_id;
  buySuccess.value = "";
  buyError.value = "";
  try {
    await api.post(
      `/api/v1/leagues/${leagueId.value}/users/${userId.value}/team/${tp.player_id}/buy`,
    );
    buySuccess.value = `Has fichado a ${name} por ${price}`;
    // Refresh the squad to reflect the change
    rivalPlayers.value = [];
    await fetchRivalSquad();
  } catch (e: unknown) {
    buyError.value = e instanceof Error ? e.message : "No se pudo completar la compra";
  } finally {
    buying.value = null;
  }
}
</script>

<style scoped>
.squad {
  position: relative;
  min-height: 100%;
  color: var(--ink-100);
}

.pitch-bg {
  opacity: 0.08;
}

.content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Header */
.header {
  display: flex;
  flex-direction: column;
  gap: 8px;
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

.back:hover {
  color: var(--lime);
}

.meta {
  font-size: 11px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--lime);
}

.title {
  font-size: 64px;
  line-height: 0.9;
  letter-spacing: 0.01em;
  text-transform: uppercase;
  margin: 0;
}

.subtitle {
  font-size: 13px;
  color: var(--ink-300);
}

/* View tabs */
.view-tabs {
  display: flex;
  border-bottom: 1px solid var(--ink-700);
}

.vtab {
  flex: 1;
  padding: 10px 0;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--ink-400);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.12em;
  cursor: pointer;
  text-align: center;
}

.vtab.active {
  color: var(--lime);
  border-bottom-color: var(--lime);
}

/* Matchday select */
.md-select {
  max-width: 360px;
  cursor: pointer;
}

/* Summary card */
.summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px;
}

.summary-right {
  text-align: right;
}

.summary-label {
  font-size: 10px;
  letter-spacing: 0.15em;
  color: var(--ink-400);
}

.summary-value {
  font-size: 32px;
  line-height: 1;
  margin-top: 4px;
}

.summary-points {
  color: var(--lime);
}

/* States */
.state {
  text-align: center;
  padding: 48px 0;
  font-size: 14px;
}

.state-muted {
  color: var(--ink-400);
}

.state-error {
  color: var(--down);
}

/* Player groups */
.groups {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.player-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 16px;
  background: var(--ink-800);
  border: 1px solid var(--ink-700);
  border-radius: var(--r-sm);
}

.player-row-buy {
  gap: 10px;
}

.player-info {
  flex: 1;
  min-width: 0;
}

.player-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink-100);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.player-team {
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--ink-400);
  margin-top: 2px;
}

.player-points {
  font-size: 20px;
  line-height: 1;
  color: var(--ink-100);
  text-align: right;
}

/* Clause column */
.clause-col {
  text-align: right;
  flex-shrink: 0;
}

.clause-label {
  font-size: 9px;
  letter-spacing: 0.15em;
  color: var(--ink-400);
}

.clause-value {
  font-size: 16px;
  line-height: 1;
  color: var(--lime);
  margin-top: 2px;
}

/* Buy button */
.buy-btn {
  flex-shrink: 0;
  height: 36px;
  font-size: 12px;
  padding: 0 14px;
}

.buy-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Buy result messages */
.buy-msg {
  text-align: center;
  padding: 12px;
  border-radius: var(--r-sm);
  font-size: 12px;
  letter-spacing: 0.08em;
}

.buy-msg-ok {
  background: rgba(199, 255, 61, 0.08);
  color: var(--lime);
  border: 1px solid var(--lime);
}

.buy-msg-err {
  background: rgba(255, 98, 98, 0.06);
  color: var(--down);
  border: 1px solid var(--down);
}

/* Fallback badge for unexpected position codes */
.pos-other {
  background: var(--ink-500);
  color: var(--ink-100);
}

@media (max-width: 767px) {
  .title {
    font-size: 40px;
  }

  .player-row-buy {
    flex-wrap: wrap;
  }

  .clause-value {
    font-size: 14px;
  }
}
</style>
