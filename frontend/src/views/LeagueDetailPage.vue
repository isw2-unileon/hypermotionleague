<template>
  <AppLayout>
    <div class="p-4 max-w-lg mx-auto">
      <!-- Back -->
      <button @click="router.push('/leagues')" class="text-green-300/60 hover:text-white transition-colors mb-4 text-sm">
        ← Mis Ligas
      </button>

      <!-- Loading -->
      <div v-if="loading" class="text-center py-12">
        <p class="text-green-300/60">Cargando...</p>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="p-3 bg-red-500/20 border border-red-400/30 rounded-lg text-red-200 text-sm">
        {{ error }}
      </div>

      <!-- Content -->
      <template v-else-if="league">
        <!-- League header -->
        <div class="bg-white/10 border border-white/10 rounded-xl p-5 mb-4">
          <h1 class="text-2xl font-bold text-white mb-3">{{ league.name }}</h1>
          <div class="space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="text-green-300/60">Presupuesto</span>
              <span class="text-white font-medium">{{ formatBudget(league.budget_per_user) }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-green-300/60">Miembros</span>
              <span class="text-white font-medium">{{ members.length }} / {{ league.max_members }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-green-300/60">Cierre mercado</span>
              <span class="text-white font-medium">{{ league.market_close_time }}</span>
            </div>
          </div>

          <!-- Invite code -->
          <div class="mt-4 pt-4 border-t border-white/10">
            <p class="text-green-300/60 text-xs mb-2">Código de invitación</p>
            <div class="flex items-center gap-2">
              <code class="flex-1 bg-white/5 px-3 py-2 rounded-lg text-amber-400 font-mono text-sm tracking-wider">
                {{ league.invite_code }}
              </code>
              <button
                @click="copyCode"
                class="px-3 py-2 bg-white/10 hover:bg-white/20 text-white text-xs rounded-lg transition-colors"
              >
                {{ copied ? "Copiado!" : "Copiar" }}
              </button>
            </div>
          </div>
        </div>

        <!-- Members list -->
        <div class="bg-white/10 border border-white/10 rounded-xl p-5 mb-4">
          <h2 class="text-white font-semibold mb-3">Miembros</h2>
          <div class="space-y-3">
            <div
              v-for="member in members"
              :key="member.id"
              class="flex items-center justify-between"
            >
              <div class="flex items-center gap-3">
                <img
                  v-if="member.avatar_url"
                  :src="member.avatar_url"
                  :alt="member.display_name || member.username"
                  class="w-8 h-8 rounded-full object-cover"
                />
                <div
                  v-else
                  class="w-8 h-8 bg-amber-500/20 rounded-full flex items-center justify-center text-amber-400 text-xs font-bold"
                >
                  {{ (member.display_name || member.username).charAt(0).toUpperCase() }}
                </div>
                <div>
                  <p class="text-white text-sm font-medium">
                    {{ member.display_name || member.username }}
                    <span v-if="member.role === 'owner'" class="text-amber-400 text-xs ml-1">👑</span>
                  </p>
                  <p class="text-green-300/40 text-xs">{{ formatBudget(member.budget) }}</p>
                </div>
              </div>
              <span class="text-green-300/40 text-xs capitalize">{{ member.role }}</span>
            </div>
          </div>
        </div>

        <!-- Delete button (owner only) -->
        <div v-if="isOwner">
          <button
            v-if="!confirmDelete"
            @click="confirmDelete = true"
            class="w-full py-3 bg-red-500/20 hover:bg-red-500/30 text-red-300 font-semibold rounded-lg border border-red-500/30 transition-colors text-sm"
          >
            Eliminar liga
          </button>
          <div v-else class="flex gap-2">
            <button
              @click="deleteLeague"
              :disabled="deleting"
              class="flex-1 py-3 bg-red-600 hover:bg-red-500 disabled:opacity-50 text-white font-semibold rounded-lg transition-colors text-sm"
            >
              {{ deleting ? "Eliminando..." : "Confirmar eliminación" }}
            </button>
            <button
              @click="confirmDelete = false"
              class="px-4 py-3 bg-white/10 text-white rounded-lg text-sm"
            >
              Cancelar
            </button>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
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

interface Member {
  id: number;
  league_id: number;
  user_id: number;
  role: string;
  budget: number;
  joined_at: string;
  username: string;
  display_name: string;
  avatar_url?: string | null;
}

const route = useRoute();
const router = useRouter();

const league = ref<League | null>(null);
const members = ref<Member[]>([]);
const loading = ref(true);
const error = ref("");
const copied = ref(false);
const confirmDelete = ref(false);
const deleting = ref(false);

// Check if current user is the owner — we decode the user ID from the JWT payload
const currentUserID = computed(() => {
  const token = localStorage.getItem("token");
  if (!token) return 0;
  try {
    const segment = token.split(".")[1];
    if (!segment) return 0;
    const payload = JSON.parse(atob(segment));
    return payload.user_id || 0;
  } catch {
    return 0;
  }
});

const isOwner = computed(() => league.value?.created_by === currentUserID.value);

function formatBudget(amount: number): string {
  return new Intl.NumberFormat("es-ES", { style: "currency", currency: "EUR", maximumFractionDigits: 0 }).format(amount);
}

async function copyCode() {
  if (!league.value) return;
  await navigator.clipboard.writeText(league.value.invite_code);
  copied.value = true;
  setTimeout(() => (copied.value = false), 2000);
}

async function deleteLeague() {
  deleting.value = true;
  try {
    await api.delete(`/api/v1/leagues/${route.params.id}`);
    router.push("/leagues");
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Error al eliminar la liga";
    confirmDelete.value = false;
  } finally {
    deleting.value = false;
  }
}

onMounted(async () => {
  try {
    const id = route.params.id;
    const [leagueData, membersData] = await Promise.all([
      api.get<League>(`/api/v1/leagues/${id}`),
      api.get<Member[]>(`/api/v1/leagues/${id}/members`),
    ]);
    league.value = leagueData;
    members.value = membersData;
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Error al cargar la liga";
  } finally {
    loading.value = false;
  }
});
</script>
