<template>
  <AppLayout>
    <div class="p-4 max-w-lg mx-auto">
      <!-- Header -->
      <div class="flex items-center gap-3 mb-6">
        <button @click="router.back()" class="text-green-300/60 hover:text-white transition-colors">
          ← Volver
        </button>
        <h1 class="text-2xl font-bold text-white">Crear Liga</h1>
      </div>

      <div v-if="error" class="mb-4 p-3 bg-red-500/20 border border-red-400/30 rounded-lg text-red-200 text-sm">
        {{ error }}
      </div>

      <form @submit.prevent="createLeague" class="space-y-5">
        <div>
          <label class="block text-green-200 text-sm font-medium mb-1">Nombre de la liga</label>
          <input
            v-model="form.name"
            type="text"
            required
            maxlength="100"
            placeholder="Liga de amigos"
            class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-green-300/50 focus:outline-none focus:ring-2 focus:ring-amber-400 transition"
          />
        </div>

        <div>
          <label class="block text-green-200 text-sm font-medium mb-1">Máximo de miembros</label>
          <input
            v-model.number="form.max_members"
            type="number"
            required
            min="2"
            max="20"
            class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-green-300/50 focus:outline-none focus:ring-2 focus:ring-amber-400 transition"
          />
          <p class="text-green-300/40 text-xs mt-1">Entre 2 y 20 participantes</p>
        </div>

        <div>
          <label class="block text-green-200 text-sm font-medium mb-1">Presupuesto por equipo (€)</label>
          <input
            v-model.number="form.budget_per_user"
            type="number"
            required
            min="1000000"
            step="1000000"
            class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-green-300/50 focus:outline-none focus:ring-2 focus:ring-amber-400 transition"
          />
          <p class="text-green-300/40 text-xs mt-1">Mínimo 1.000.000 €</p>
        </div>

        <button
          type="submit"
          :disabled="submitting"
          class="w-full py-3 bg-amber-500 hover:bg-amber-400 disabled:bg-amber-500/50 text-white font-bold rounded-lg transition-all shadow-lg"
        >
          {{ submitting ? "Creando..." : "Crear liga" }}
        </button>
      </form>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive } from "vue";
import { useRouter } from "vue-router";
import AppLayout from "@/layouts/AppLayout.vue";
import api from "@/lib/api";

const router = useRouter();
const error = ref("");
const submitting = ref(false);

const form = reactive({
  name: "",
  max_members: 10,
  budget_per_user: 100000000,
});

async function createLeague() {
  error.value = "";
  submitting.value = true;
  try {
    const league = await api.post<{ id: number }>("/api/v1/leagues", form);
    router.push(`/leagues/${league.id}`);
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Error al crear la liga";
  } finally {
    submitting.value = false;
  }
}
</script>
