<template>
  <AppLayout>
    <div class="px-4 pt-6 pb-4">
      <h1 class="text-xl font-semibold text-white mb-4">Mi Equipo</h1>

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

      <!-- Loading -->
      <div v-if="loading" class="text-center text-green-300/60 py-12 text-sm">
        Cargando equipo...
      </div>

      <!-- Error -->
      <div v-else-if="error" class="text-center text-red-400 py-12 text-sm">
        {{ error }}
      </div>

      <!-- Empty state: no league selected -->
      <div v-else-if="!selectedLeagueId" class="text-center text-green-300/60 py-12 text-sm">
        Selecciona una liga para ver tu equipo
      </div>

      <!-- Empty squad -->
      <div
        v-else-if="team && team.players.length === 0"
        class="text-center py-12"
      >
        <p class="text-4xl mb-3">👕</p>
        <p class="text-white font-semibold mb-1">Tu plantilla está vacía</p>
        <p class="text-green-300/60 text-sm">Ve al mercado para fichar jugadores</p>
      </div>

      <div v-else-if="team">
        <!-- Summary bar -->
        <div class="flex gap-3 mb-4">
          <div class="flex-1 bg-green-950/60 border border-white/10 rounded-xl p-3 text-center">
            <p class="text-green-300/60 text-xs mb-1">Valor plantilla</p>
            <p class="text-white font-bold text-sm">{{ formatMoney(team.total_value) }}</p>
          </div>
          <div class="flex-1 bg-green-950/60 border border-white/10 rounded-xl p-3 text-center">
            <p class="text-green-300/60 text-xs mb-1">Presupuesto</p>
            <p class="text-amber-400 font-bold text-sm">{{ formatMoney(team.budget) }}</p>
          </div>
          <div class="flex-1 bg-green-950/60 border border-white/10 rounded-xl p-3 text-center">
            <p class="text-green-300/60 text-xs mb-1">Jugadores</p>
            <p class="text-white font-bold text-sm">{{ team.players.length }}</p>
          </div>
        </div>

        <!-- Go to lineup button -->
        <button
          v-if="currentMatchday"
          @click="goToLineup"
          class="w-full py-3 bg-amber-500 hover:bg-amber-400 text-white font-semibold rounded-xl text-sm mb-5 transition-colors"
        >
          Editar alineación — Jornada {{ currentMatchday.number }}
        </button>

        <!-- Players grouped by position -->
        <div v-for="pos in POSITIONS" :key="pos">
          <div v-if="playersByPosition(pos).length > 0" class="mb-5">
            <h3 class="text-green-300/60 text-xs font-semibold uppercase tracking-wide mb-2 px-1">
              {{ positionLabel(pos) }} ({{ playersByPosition(pos).length }})
            </h3>
            <div class="rounded-xl overflow-hidden border border-white/10">
              <div
                v-for="(entry, i) in playersByPosition(pos)"
                :key="entry.player_id"
                class="flex items-center px-4 py-3"
                :class="i > 0 ? 'border-t border-white/5' : ''"
              >
                <!-- Position badge -->
                <span
                  class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold mr-3 flex-shrink-0"
                  :class="positionBadge(pos)"
                >
                  {{ pos }}
                </span>

                <!-- Player info -->
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium text-white truncate">
                    {{ entry.player.first_name }} {{ entry.player.last_name }}
                  </p>
                  <p class="text-xs text-green-300/50">{{ entry.player.team_name }}</p>
                </div>

                <!-- Values -->
                <div class="text-right ml-3">
                  <p class="text-xs font-semibold text-amber-400">
                    {{ formatMoney(entry.player.market_value) }}
                  </p>
                  <p class="text-xs text-green-300/40">
                    Compra: {{ formatMoney(entry.purchase_price) }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import AppLayout from '@/layouts/AppLayout.vue';
import api from '@/lib/api';

type PlayerPosition = 'GK' | 'DEF' | 'MID' | 'FWD';

const POSITIONS: PlayerPosition[] = ['GK', 'DEF', 'MID', 'FWD'];

interface League {
  id: number;
  name: string;
  budget_per_user: number;
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

interface Matchday {
  id: number;
  number: number;
  name: string;
  end_date: string;
  is_current: boolean;
}

const router = useRouter();
const leagues = ref<League[]>([]);
const selectedLeagueId = ref<number | ''>('');
const team = ref<UserTeam | null>(null);
const currentMatchday = ref<Matchday | null>(null);
const loading = ref(false);
const error = ref('');

function formatMoney(amount: number): string {
  return new Intl.NumberFormat('es-ES', {
    style: 'currency',
    currency: 'EUR',
    maximumFractionDigits: 0,
  }).format(amount);
}

function positionLabel(pos: PlayerPosition): string {
  return { GK: 'Porteros', DEF: 'Defensas', MID: 'Centrocampistas', FWD: 'Delanteros' }[pos];
}

function positionBadge(pos: PlayerPosition): string {
  return {
    GK: 'bg-amber-500/20 text-amber-400',
    DEF: 'bg-blue-500/20 text-blue-400',
    MID: 'bg-green-500/20 text-green-400',
    FWD: 'bg-red-500/20 text-red-400',
  }[pos];
}

function playersByPosition(pos: PlayerPosition): TeamPlayerWithDetails[] {
  return team.value?.players.filter(p => p.player.position === pos) ?? [];
}

async function onLeagueChange() {
  if (!selectedLeagueId.value) return;
  team.value = null;
  currentMatchday.value = null;
  error.value = '';
  loading.value = true;
  try {
    const [teamData, matchdaysData] = await Promise.all([
      api.get<UserTeam>(`/api/v1/leagues/${selectedLeagueId.value}/team`),
      api.get<{ matchdays: Matchday[] }>(`/api/v1/leagues/${selectedLeagueId.value}/matchdays`),
    ]);
    team.value = teamData;
    const matchdays = matchdaysData.matchdays ?? [];
    currentMatchday.value = matchdays.find(m => m.is_current) ?? matchdays[0] ?? null;
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Error al cargar el equipo';
  } finally {
    loading.value = false;
  }
}

function goToLineup() {
  if (!selectedLeagueId.value || !currentMatchday.value) return;
  router.push(`/lineup/${selectedLeagueId.value}/${currentMatchday.value.number}`);
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
