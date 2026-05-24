<template>
  <AppLayout>
    <div class="px-4 pt-6 pb-4">
      <div class="flex items-center justify-between mb-4">
        <h1 class="text-xl font-semibold text-white">
          Equipo de {{ userName }}
        </h1>
        <button
          @click="router.back()"
          class="text-sm text-green-300 hover:text-white transition-colors"
        >
          ← Volver
        </button>
      </div>

      <!-- Matchday selector -->
      <select
        v-model="selectedMatchdayNumber"
        class="w-full bg-green-950/60 border border-white/10 text-white rounded-lg px-3 py-2 mb-4 text-sm"
        @change="onMatchdayChange"
      >
        <option :value="null">General (Total)</option>
        <option v-for="md in matchdays" :key="md.id" :value="md.number">
          Jornada {{ md.number }} — {{ md.name }}
        </option>
      </select>

      <!-- Loading -->
      <div v-if="loading" class="text-center text-green-300/60 py-12 text-sm">
        Cargando equipo...
      </div>

      <!-- Error -->
      <div v-else-if="error" class="text-center text-red-400 py-12 text-sm">
        {{ error }}
      </div>

      <!-- User info -->
      <div
        v-else-if="userStanding && selectedMatchdayNumber"
        class="rounded-xl bg-green-950/40 border border-white/10 p-4 mb-4"
      >
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-green-300/60">Posición</p>
            <p class="text-2xl font-bold" :class="rankClass">
              #{{ userStanding.rank }}
            </p>
          </div>
          <div class="text-right">
            <p class="text-sm text-green-300/60">Puntos</p>
            <p class="text-2xl font-bold text-amber-400">
              {{ userStanding.total_points }}
            </p>
          </div>
        </div>
      </div>

      <!-- Empty states -->
      <div
        v-if="!loading && matchdays.length === 0"
        class="text-center text-green-300/60 py-12 text-sm"
      >
        No hay jornadas creadas en esta liga todavía.
      </div>

      <div
        v-else-if="!selectedMatchdayNumber"
        class="text-center text-green-300/60 py-12 text-sm"
      >
        Selecciona una jornada para ver los jugadores
      </div>

      <div
        v-else-if="!userStanding && !loading"
        class="text-center text-green-300/60 py-8 text-sm"
      >
        No hay datos para esta jornada
      </div>

      <!-- Players table -->
      <div
        v-if="
          userStanding &&
          userStanding.players &&
          userStanding.players.length > 0
        "
        class="rounded-xl overflow-hidden border border-white/10"
      >
        <!-- Header -->
        <div
          class="grid grid-cols-12 px-4 py-2 bg-green-950/80 text-green-300/60 text-xs font-semibold uppercase tracking-wide"
        >
          <span class="col-span-1">#</span>
          <span class="col-span-5">Jugador</span>
          <span class="col-span-3">Posición</span>
          <span class="col-span-3 text-right">Pts</span>
        </div>

        <!-- Rows -->
        <div
          v-for="(player, index) in userStanding.players"
          :key="player.player_id"
          class="grid grid-cols-12 items-center px-4 py-3 border-t border-white/5"
        >
          <span class="col-span-1 text-sm text-white/50">
            {{ index + 1 }}
          </span>
          <div class="col-span-5">
            <p class="text-sm font-medium text-white">
              {{ player.first_name }} {{ player.last_name }}
            </p>
            <p class="text-xs text-green-300/50">{{ player.team_name }}</p>
          </div>
          <span
            class="col-span-3 text-sm"
            :class="positionClass(player.position)"
          >
            {{ player.position }}
          </span>
          <span
            class="col-span-3 text-right text-sm font-semibold text-amber-400"
          >
            {{ player.points }}
          </span>
        </div>

        <!-- Empty players -->
        <div
          v-if="userStanding.players.length === 0"
          class="text-center text-green-300/60 py-8 text-sm border-t border-white/5"
        >
          Este usuario no tiene jugadores en su alineación para esta jornada.
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import AppLayout from "@/layouts/AppLayout.vue";
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

const rankClass = computed(() => {
  if (!userStanding.value) return "text-white";
  const rank = userStanding.value.rank;
  if (rank === 1) return "text-amber-400";
  if (rank <= 3) return "text-green-300";
  return "text-white/70";
});

function positionClass(position: string): string {
  switch (position) {
    case "GK":
      return "text-orange-400";
    case "DEF":
      return "text-blue-400";
    case "MID":
      return "text-green-400";
    case "FWD":
      return "text-red-400";
    default:
      return "text-white/70";
  }
}

onMounted(async () => {
  await fetchMatchdays();
  if (matchdays.value.length > 0) {
    selectedMatchdayNumber.value = matchdays.value[0].number;
    await fetchUserStanding();
  }
});

watch(leagueId, async () => {
  await fetchMatchdays();
  if (matchdays.value.length > 0) {
    selectedMatchdayNumber.value = matchdays.value[0].number;
    await fetchUserStanding();
  }
});

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
      userName.value = userRank.display_name || userRank.username || "Usuario";
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
</script>
