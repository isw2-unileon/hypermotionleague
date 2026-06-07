<template>
  <AppShell>
    <div class="standings">
      <!-- Header -->
      <header class="header">
        <div class="meta mono">◆ CLASIFICACIÓN</div>
        <h1 class="title display">Clasificación.</h1>
      </header>

      <!-- League selector -->
      <select
        v-model="selectedLeagueId"
        class="input league-select"
        @change="onLeagueChange"
      >
        <option disabled value="">Selecciona una liga</option>
        <option v-for="league in leagues" :key="league.id" :value="league.id">
          {{ league.name }}
        </option>
      </select>

      <!-- Matchday filter pills -->
      <div v-if="matchdays.length > 0" class="pills">
        <button
          type="button"
          class="pill"
          :class="{ active: selectedMatchdayNumber === null }"
          @click="selectMatchday(null)"
        >
          General
        </button>
        <button
          v-for="md in matchdays"
          :key="md.id"
          type="button"
          class="pill"
          :class="{ active: selectedMatchdayNumber === md.number }"
          @click="selectMatchday(md.number)"
        >
          J·{{ md.number }}
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="state state-muted">
        Cargando clasificación...
      </div>

      <!-- Error -->
      <div v-else-if="error" class="state state-error">
        {{ error }}
      </div>

      <!-- Empty state -->
      <div v-else-if="!selectedLeagueId" class="state state-muted">
        Selecciona una liga para ver la clasificación
      </div>

      <!-- Standings -->
      <div v-else-if="standings" class="results">
        <div v-if="standings.rankings.length === 0" class="state state-muted">
          No hay datos para esta jornada
        </div>

        <template v-else>
          <Podium
            v-if="podium.length > 0"
            :top3="podium"
            :mobile="isMobile"
            @click="goToUserSquad"
          />

          <div class="rows">
            <StandingsRow
              v-for="row in restRows"
              :key="row.userId"
              :row="row"
              :mobile="isMobile"
              @click="goToUserSquad"
            />
          </div>
        </template>
      </div>
    </div>
  </AppShell>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import AppShell from "@/design-system/AppShell.vue";
import Podium from "@/design-system/components/Podium.vue";
import StandingsRow from "@/design-system/components/StandingsRow.vue";
import type { StandingsRow as StandingsRowData } from "@/types/standings";
import { currentUserId } from "@/lib/auth";
import api from "@/lib/api";
import { lastStandingsLeagueId } from "@/lib/standings-memory";
import { activeLeagueId } from "@/lib/activeLeague";

interface League {
  id: number;
  name: string;
}

interface Matchday {
  id: number;
  number: number;
  name: string;
}

interface UserStanding {
  rank: number;
  user_id: number;
  username: string;
  display_name: string;
  total_points: number;
}

interface Standings {
  league_id: number;
  matchday_id?: number;
  rankings: UserStanding[];
}

const route = useRoute();
const router = useRouter();

const leagues = ref<League[]>([]);
const matchdays = ref<Matchday[]>([]);
const standings = ref<Standings | null>(null);
const selectedLeagueId = ref<number | "">("");
const selectedMatchdayNumber = ref<number | null>(null);
const loading = ref(false);
const error = ref("");
const isMobile = ref(false);

// Accent color cycled by row index: mid → def → fwd → gk.
function cycleColor(index: number): string {
  switch (index % 4) {
    case 0:
      return "var(--pos-mid)";
    case 1:
      return "var(--pos-def)";
    case 2:
      return "var(--pos-fwd)";
    default:
      return "var(--pos-gk)";
  }
}

function initialsFrom(name: string): string {
  return name
    .split(" ")
    .filter((part) => part.length > 0)
    .map((part) => part.charAt(0).toUpperCase())
    .join("");
}

// Maps the backend rankings onto the design-system StandingsRow shape.
const standingsRows = computed<StandingsRowData[]>(() => {
  const rankings = standings.value?.rankings ?? [];
  const meId = currentUserId();
  return rankings.map((entry, index) => ({
    position: entry.rank,
    userId: entry.user_id,
    name: entry.display_name,
    initials: initialsFrom(entry.display_name),
    squadName: entry.username ?? "",
    totalPoints: entry.total_points,
    matchdayPoints: 0, // TODO(Sprint 2): backend doesn't send per-matchday points yet.
    deltaPosition: 0, // TODO(Sprint 2): backend doesn't send position delta yet.
    color: cycleColor(index),
    isCurrentUser: meId !== 0 && entry.user_id === meId,
  }));
});

const podium = computed<StandingsRowData[]>(() => standingsRows.value.slice(0, 3));
const restRows = computed<StandingsRowData[]>(() => standingsRows.value.slice(3));

function updateIsMobile(): void {
  isMobile.value = window.innerWidth < 768;
}

onMounted(async () => {
  updateIsMobile();
  window.addEventListener("resize", updateIsMobile);

  try {
    leagues.value = await api.get<League[]>("/api/v1/leagues");
  } catch {
    error.value = "No se pudieron cargar las ligas";
    return;
  }

  const queryLeagueId = Number(route.query.leagueId);
  const queryMatchday = route.query.matchday;

  // Priority: URL query param → last remembered league → first league of the user.
  let resolvedId: number | null = null;
  if (queryLeagueId && leagues.value.some((l) => l.id === queryLeagueId)) {
    resolvedId = queryLeagueId;
  } else if (
    lastStandingsLeagueId.value !== null &&
    leagues.value.some((l) => l.id === lastStandingsLeagueId.value)
  ) {
    resolvedId = lastStandingsLeagueId.value;
  } else if (leagues.value.length > 0) {
    resolvedId = leagues.value[0]!.id;
  }

  if (resolvedId !== null) {
    selectedLeagueId.value = resolvedId;
    lastStandingsLeagueId.value = resolvedId;
    activeLeagueId.value = resolvedId;
    await fetchMatchdays();
    if (queryMatchday !== undefined && queryMatchday !== null && queryMatchday !== "") {
      const num = Number(queryMatchday);
      if (matchdays.value.some((m) => m.number === num)) {
        selectedMatchdayNumber.value = num;
      }
    }
    await fetchStandings();
  }
});

onUnmounted(() => {
  window.removeEventListener("resize", updateIsMobile);
});

watch([selectedLeagueId, selectedMatchdayNumber], ([leagueId, matchday]) => {
  const query: Record<string, string> = {};
  if (leagueId !== "") query.leagueId = String(leagueId);
  if (matchday !== null) query.matchday = String(matchday);
  router.replace({ query });
});

async function onLeagueChange() {
  if (!selectedLeagueId.value) return;
  lastStandingsLeagueId.value = selectedLeagueId.value as number;
  activeLeagueId.value = selectedLeagueId.value as number;
  selectedMatchdayNumber.value = null;
  matchdays.value = [];
  standings.value = null;
  error.value = "";

  await Promise.all([fetchMatchdays(), fetchStandings()]);
}

async function onMatchdayChange() {
  await fetchStandings();
}

// Pill click → drive the same matchday selection the original <select> did.
async function selectMatchday(num: number | null): Promise<void> {
  if (selectedMatchdayNumber.value === num) return;
  selectedMatchdayNumber.value = num;
  await onMatchdayChange();
}

async function fetchMatchdays() {
  try {
    const data = await api.get<{ matchdays: Matchday[] }>(
      `/api/v1/leagues/${selectedLeagueId.value}/matchdays`,
    );
    matchdays.value = data.matchdays ?? [];
  } catch {
    // no blocking — standings can still load
  }
}

async function fetchStandings() {
  loading.value = true;
  error.value = "";
  try {
    const path =
      selectedMatchdayNumber.value !== null
        ? `/api/v1/leagues/${selectedLeagueId.value}/matchdays/${selectedMatchdayNumber.value}/standings`
        : `/api/v1/leagues/${selectedLeagueId.value}/standings`;

    standings.value = await api.get<Standings>(path);
  } catch (e: unknown) {
    error.value =
      e instanceof Error ? e.message : "Error al cargar la clasificación";
    standings.value = null;
  } finally {
    loading.value = false;
  }
}

function goToUserSquad(userId: number) {
  router.push(`/squad/${selectedLeagueId.value}/${userId}`);
}
</script>

<style scoped>
.standings {
  display: flex;
  flex-direction: column;
  gap: 20px;
  color: var(--ink-100);
}

/* Header */
.header {
  display: flex;
  flex-direction: column;
  gap: 8px;
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
  margin: 0;
}

/* League select — reuses the global .input token styling */
.league-select {
  max-width: 360px;
  cursor: pointer;
}

/* Matchday pills */
.pills {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.pill {
  padding: 6px 14px;
  border-radius: var(--r-xs);
  background: transparent;
  color: var(--ink-200);
  border: 1px solid var(--ink-700);
  font-family: var(--f-mono);
  font-size: 10px;
  letter-spacing: 0.12em;
  font-weight: 600;
  text-transform: uppercase;
  cursor: pointer;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
}

.pill:hover {
  border-color: var(--ink-500);
}

.pill.active {
  background: var(--lime);
  color: var(--ink-900);
  border-color: var(--lime);
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

/* Results */
.results {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.rows {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

@media (max-width: 767px) {
  .title {
    font-size: 40px;
  }
}
</style>
