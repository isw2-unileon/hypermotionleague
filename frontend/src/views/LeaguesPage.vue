<template>
  <AppLayout>
    <div class="p-4 max-w-lg mx-auto">
      <!-- Header -->
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-bold text-white">Mis Ligas</h1>
        <div class="flex items-center gap-2">
          <RouterLink
            to="/leagues/new"
            class="px-4 py-2 bg-amber-500 hover:bg-amber-400 text-white text-sm font-semibold rounded-lg transition-colors"
          >
            + Crear liga
          </RouterLink>
          <button
            @click="logout"
            class="px-3 py-2 bg-white/10 hover:bg-white/20 text-green-300/80 hover:text-white text-xs font-medium rounded-lg border border-white/10 transition-colors"
          >
            Cerrar sesión
          </button>
        </div>
      </div>

      <!-- Join by code -->
      <form @submit.prevent="joinLeague" class="flex gap-2 mb-6">
        <input
          v-model="inviteCode"
          type="text"
          placeholder="Código de invitación"
          class="flex-1 px-4 py-2 bg-white/10 border border-white/20 rounded-lg text-white placeholder-green-300/50 focus:outline-none focus:ring-2 focus:ring-amber-400 text-sm"
        />
        <button
          type="submit"
          :disabled="!inviteCode.trim() || joining"
          class="px-4 py-2 bg-white/10 hover:bg-white/20 disabled:opacity-40 text-white text-sm font-semibold rounded-lg border border-white/20 transition-colors"
        >
          {{ joining ? "..." : "Unirse" }}
        </button>
      </form>

      <!-- Error / Success -->
      <div v-if="error" class="mb-4 p-3 bg-red-500/20 border border-red-400/30 rounded-lg text-red-200 text-sm">
        {{ error }}
      </div>
      <div v-if="joinSuccess" class="mb-4 p-3 bg-green-500/20 border border-green-400/30 rounded-lg text-green-200 text-sm">
        {{ joinSuccess }}
      </div>

      <!-- Loading -->
      <div v-if="loading" class="text-center py-12">
        <p class="text-green-300/60">Cargando ligas...</p>
      </div>

      <!-- Empty state -->
      <div v-else-if="leagues.length === 0" class="text-center py-12">
        <p class="text-4xl mb-3">🏟️</p>
        <p class="text-white font-semibold mb-1">No estás en ninguna liga</p>
        <p class="text-green-300/60 text-sm">Crea una liga o únete con un código de invitación</p>
      </div>

      <!-- League list -->
      <div v-else class="space-y-3">
        <RouterLink
          v-for="league in leagues"
          :key="league.id"
          :to="`/leagues/${league.id}`"
          class="block bg-white/10 border border-white/10 rounded-xl p-4 hover:bg-white/15 transition-colors"
        >
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-white font-semibold">{{ league.name }}</h3>
              <p class="text-green-300/60 text-xs mt-1">
                {{ formatBudget(league.budget_per_user) }} por equipo
              </p>
            </div>
            <span class="text-green-300/40 text-xl">›</span>
          </div>
        </RouterLink>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter, RouterLink } from "vue-router";
import AppLayout from "@/layouts/AppLayout.vue";
import api from "@/lib/api";

interface League {
  id: number;
  name: string;
  invite_code: string;
  max_members: number;
  budget_per_user: number;
  market_close_time: string;
  created_by: number;
}

const router = useRouter();
const leagues = ref<League[]>([]);
const loading = ref(true);
const error = ref("");
const inviteCode = ref("");
const joining = ref(false);
const joinSuccess = ref("");

function formatBudget(amount: number): string {
  return new Intl.NumberFormat("es-ES", { style: "currency", currency: "EUR", maximumFractionDigits: 0 }).format(amount);
}

async function fetchLeagues() {
  loading.value = true;
  error.value = "";
  try {
    leagues.value = await api.get<League[]>("/api/v1/leagues");
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Error al cargar ligas";
  } finally {
    loading.value = false;
  }
}

async function joinLeague() {
  joining.value = true;
  error.value = "";
  joinSuccess.value = "";
  try {
    const league = await api.post<League>("/api/v1/leagues/join", {
      invite_code: inviteCode.value.trim(),
    });
    inviteCode.value = "";
    joinSuccess.value = `Te has unido a "${league.name}"`;
    router.push(`/leagues/${league.id}`);
  } catch (e) {
    error.value = e instanceof Error ? e.message : "No se pudo unir a la liga";
  } finally {
    joining.value = false;
  }
}

function logout() {
  localStorage.removeItem("token");
  router.push("/auth");
}

onMounted(fetchLeagues);
</script>
