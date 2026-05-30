<template>
  <AppLayout>
    <div class="px-4 pt-6 pb-4">
      <h1 class="text-xl font-semibold text-white mb-4">Clasificación</h1>

      <!-- League selector -->
      <select
        v-model="selectedLeagueId"
        class="w-full bg-green-950/60 border border-white/10 text-white rounded-lg px-3 py-2 mb-4 text-sm"
        @change="onLeagueChange"
      >
        <option disabled value="">Selecciona una liga</option>
        <option v-for="league in leagues" :key="league.id" :value="league.id">
          {{ league.name }}
        </option>
      </select>

      <!-- Matchday filter -->
      <select
        v-if="matchdays.length > 0"
        v-model="selectedMatchdayNumber"
        class="w-full bg-green-950/60 border border-white/10 text-white rounded-lg px-3 py-2 mb-4 text-sm"
        @change="onMatchdayChange"
      >
        <option :value="null">General</option>
        <option v-for="md in matchdays" :key="md.id" :value="md.number">
          Jornada {{ md.number }} — {{ md.name }}
        </option>
      </select>

      <!-- Loading -->
      <div v-if="loading" class="text-center text-green-300/60 py-12 text-sm">
        Cargando clasificación...
      </div>

      <!-- Error -->
      <div v-else-if="error" class="text-center text-red-400 py-12 text-sm">
        {{ error }}
      </div>

      <!-- Empty state -->
      <div
        v-else-if="!selectedLeagueId"
        class="text-center text-green-300/60 py-12 text-sm"
      >
        Selecciona una liga para ver la clasificación
      </div>

      <!-- Standings table -->
      <div
        v-else-if="standings"
        class="rounded-xl overflow-hidden border border-white/10"
      >
        <!-- Header -->
        <div
          class="grid grid-cols-12 px-4 py-2 bg-green-950/80 text-green-300/60 text-xs font-semibold uppercase tracking-wide"
        >
          <span class="col-span-1">#</span>
          <span class="col-span-8">Jugador</span>
          <span class="col-span-3 text-right">Pts</span>
        </div>

        <!-- Rows -->
        <div
          v-for="entry in standings.rankings"
          :key="entry.user_id"
          class="grid grid-cols-12 items-center px-4 py-3 border-t border-white/5 cursor-pointer hover:bg-white/5 transition-colors"
          @click="goToUserSquad(entry.user_id)"
        >
          <span
            class="col-span-1 text-sm font-semibold"
            :class="
              entry.rank === 1
                ? 'text-amber-400'
                : entry.rank <= 3
                  ? 'text-green-300'
                  : 'text-white/50'
            "
          >
            {{ entry.rank }}
          </span>
          <div class="col-span-8">
            <p class="text-sm font-medium text-white">
              {{ entry.display_name }}
            </p>
            <p class="text-xs text-green-300/50">@{{ entry.username }}</p>
          </div>
          <span
            class="col-span-3 text-right text-sm font-semibold text-amber-400"
          >
            {{ entry.total_points }}
          </span>
        </div>

        <!-- Empty rankings -->
        <div
          v-if="standings.rankings.length === 0"
          class="text-center text-green-300/60 py-8 text-sm border-t border-white/5"
        >
          No hay datos para esta jornada
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import AppLayout from "@/layouts/AppLayout.vue";
import api from "@/lib/api";

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

onMounted(async () => {
  try {
    leagues.value = await api.get<League[]>("/api/v1/leagues");
  } catch {
    error.value = "No se pudieron cargar las ligas";
    return;
  }

  const queryLeagueId = Number(route.query.leagueId);
  const queryMatchday = route.query.matchday;

  if (
    queryLeagueId &&
    leagues.value.some((l) => l.id === queryLeagueId)
  ) {
    selectedLeagueId.value = queryLeagueId;
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

watch([selectedLeagueId, selectedMatchdayNumber], ([leagueId, matchday]) => {
  const query: Record<string, string> = {};
  if (leagueId !== "") query.leagueId = String(leagueId);
  if (matchday !== null) query.matchday = String(matchday);
  router.replace({ query });
});

async function onLeagueChange() {
  if (!selectedLeagueId.value) return;
  selectedMatchdayNumber.value = null;
  matchdays.value = [];
  standings.value = null;
  error.value = "";

  await Promise.all([fetchMatchdays(), fetchStandings()]);
}

async function onMatchdayChange() {
  await fetchStandings();
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
