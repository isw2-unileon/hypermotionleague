<template>
  <AppLayout>
    <div class="px-4 pt-6 pb-4">
      <!-- Header -->
      <div class="flex items-center gap-3 mb-4">
        <button
          @click="router.back()"
          class="text-green-300/60 hover:text-white transition-colors text-sm"
        >
          ← Volver
        </button>
        <h1 class="text-xl font-semibold text-white">Alineación</h1>
      </div>

      <!-- Matchday info -->
      <div
        v-if="matchday"
        class="bg-green-950/60 border border-white/10 rounded-xl p-3 mb-4"
      >
        <p class="text-white font-semibold text-sm">
          Jornada {{ matchday.number }} — {{ matchday.name }}
        </p>
        <p class="text-green-300/50 text-xs">Cierre: {{ formatDate(matchday.end_date) }}</p>
      </div>

      <!-- Formation selector -->
      <div class="mb-4">
        <p class="text-green-300/60 text-xs font-semibold uppercase tracking-wide mb-2">
          Formación
        </p>
        <div class="flex gap-2 flex-wrap">
          <button
            v-for="f in formationKeys"
            :key="f"
            @click="selectFormation(f)"
            class="px-3 py-1.5 rounded-lg text-sm font-semibold border transition-colors"
            :class="
              formation === f
                ? 'bg-amber-500 border-amber-500 text-white'
                : 'bg-white/10 border-white/20 text-green-300/80 hover:bg-white/20'
            "
          >
            {{ f }}
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="text-center text-green-300/60 py-12 text-sm">
        Cargando...
      </div>

      <!-- Error -->
      <div v-else-if="error" class="text-center text-red-400 py-6 text-sm">
        {{ error }}
      </div>

      <!-- No players -->
      <div v-else-if="players.length === 0" class="text-center text-green-300/60 py-12 text-sm">
        <p class="text-4xl mb-3">👕</p>
        <p class="text-white font-semibold mb-1">No tienes jugadores en esta liga</p>
        <RouterLink to="/market" class="text-amber-400 underline text-sm">Ir al mercado</RouterLink>
      </div>

      <div v-else>
        <!-- Position slot progress -->
        <div class="flex gap-2 mb-4">
          <div
            v-for="pos in POSITIONS"
            :key="pos"
            class="flex-1 rounded-lg p-2 text-center border transition-colors"
            :class="
              slotsFilled(pos) === formationSlots(pos)
                ? 'border-green-400/40 bg-green-500/10'
                : 'border-white/10 bg-white/5'
            "
          >
            <p class="text-xs font-bold mb-0.5" :class="positionColor(pos)">{{ pos }}</p>
            <p class="text-white text-xs font-semibold">
              {{ slotsFilled(pos) }}/{{ formationSlots(pos) }}
            </p>
          </div>
        </div>

        <!-- Save button -->
        <button
          @click="saveLineup"
          :disabled="!isLineupValid || saving"
          class="w-full py-3 rounded-xl text-sm font-semibold mb-3 transition-colors"
          :class="
            isLineupValid
              ? 'bg-amber-500 hover:bg-amber-400 text-white'
              : 'bg-white/10 text-white/30 cursor-not-allowed'
          "
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
        <div
          v-if="saveError"
          class="mb-4 p-3 bg-red-500/20 border border-red-400/30 rounded-lg text-red-200 text-sm"
        >
          {{ saveError }}
        </div>
        <div
          v-if="saveSuccess"
          class="mb-4 p-3 bg-green-500/20 border border-green-400/30 rounded-lg text-green-200 text-sm"
        >
          ¡Alineación guardada correctamente!
        </div>

        <!-- Players grouped by position -->
        <div v-for="pos in POSITIONS" :key="pos" class="mb-5">
          <h3 class="text-green-300/60 text-xs font-semibold uppercase tracking-wide mb-2 px-1">
            {{ positionLabel(pos) }}
          </h3>
          <div
            v-if="playersByPosition(pos).length > 0"
            class="rounded-xl overflow-hidden border border-white/10"
          >
            <div
              v-for="(player, i) in playersByPosition(pos)"
              :key="player.player_id"
              class="flex items-center px-4 py-3 cursor-pointer transition-colors"
              :class="[
                i > 0 ? 'border-t border-white/5' : '',
                starterMap[player.player_id]
                  ? 'bg-amber-500/10'
                  : canAddStarter(pos)
                  ? 'hover:bg-white/5'
                  : 'opacity-50',
              ]"
              @click="toggleStarter(player)"
            >
              <!-- Checkbox -->
              <div
                class="w-5 h-5 rounded-full border-2 flex items-center justify-center mr-3 flex-shrink-0 transition-colors"
                :class="
                  starterMap[player.player_id]
                    ? 'border-amber-400 bg-amber-400'
                    : 'border-white/20'
                "
              >
                <span v-if="starterMap[player.player_id]" class="text-white text-xs leading-none">
                  ✓
                </span>
              </div>

              <!-- Player info -->
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-white truncate">
                  {{ player.player.first_name }} {{ player.player.last_name }}
                </p>
                <p class="text-xs text-green-300/50">{{ player.player.team_name }}</p>
              </div>

              <!-- Starter / Bench badge -->
              <span
                class="ml-3 px-2 py-0.5 rounded text-xs font-semibold flex-shrink-0"
                :class="
                  starterMap[player.player_id]
                    ? 'bg-amber-500/20 text-amber-400'
                    : 'bg-white/10 text-white/40'
                "
              >
                {{ starterMap[player.player_id] ? 'Titular' : 'Suplente' }}
              </span>
            </div>
          </div>
          <p v-else class="text-green-300/40 text-xs px-1">
            No tienes jugadores en esta posición
          </p>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter, RouterLink } from 'vue-router';
import AppLayout from '@/layouts/AppLayout.vue';
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

function positionColor(pos: PlayerPosition): string {
  return {
    GK: 'text-amber-400',
    DEF: 'text-blue-400',
    MID: 'text-green-400',
    FWD: 'text-red-400',
  }[pos];
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
