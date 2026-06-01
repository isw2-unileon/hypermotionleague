<template>
  <AppShell>
    <div class="create">
      <!-- Back -->
      <button type="button" class="back mono" @click="router.push('/leagues')">
        ← Mis Ligas
      </button>

      <!-- Header -->
      <header class="header">
        <div class="meta meta-lime mono">◆ NUEVA LIGA</div>
        <h1 class="title display">CREAR LIGA.</h1>
      </header>

      <!-- Error -->
      <div v-if="error" class="state state-error">{{ error }}</div>

      <!-- Form card -->
      <form class="card form-card" @submit.prevent="createLeague">
        <div class="field">
          <label class="label" for="league-name">Nombre de la liga</label>
          <input
            id="league-name"
            v-model="form.name"
            type="text"
            required
            maxlength="100"
            placeholder="Liga de amigos"
            class="input"
          />
        </div>

        <div class="field">
          <label class="label" for="league-max">Máximo de miembros</label>
          <input
            id="league-max"
            v-model.number="form.max_members"
            type="number"
            required
            min="2"
            max="20"
            class="input mono"
          />
          <p class="hint mono">Entre 2 y 20 participantes</p>
        </div>

        <div class="field">
          <label class="label" for="league-budget">Presupuesto por equipo (€)</label>
          <input
            id="league-budget"
            v-model.number="form.budget_per_user"
            type="number"
            required
            min="1000000"
            step="1000000"
            class="input mono"
          />
          <p class="hint mono">{{ budgetHint }}</p>
        </div>

        <button type="submit" class="btn btn-primary submit" :disabled="submitting">
          {{ submitting ? "Creando..." : "Crear liga" }}
        </button>
      </form>
    </div>
  </AppShell>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from "vue";
import { useRouter } from "vue-router";
import AppShell from "@/design-system/AppShell.vue";
import api from "@/lib/api";

const router = useRouter();
const error = ref("");
const submitting = ref(false);

// Data wiring unchanged from D1.2: same fields, same field names, same endpoint.
const form = reactive({
  name: "",
  max_members: 10,
  budget_per_user: 100000000,
});

// Friendly compact echo of the raw euro amount the form submits (e.g. "100M").
const budgetHint = computed(() => {
  const millions = Math.round(form.budget_per_user / 1_000_000);
  return `Mínimo 1.000.000 € · actual ≈ ${millions}M`;
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

<style scoped>
.create {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 520px;
  color: var(--ink-100);
}

.back {
  align-self: flex-start;
  background: transparent;
  border: none;
  color: var(--ink-400);
  font-size: 11px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  cursor: pointer;
  padding: 0;
  transition: color 0.15s;
}

.back:hover {
  color: var(--ink-100);
}

.header {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.meta {
  font-size: 11px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--ink-400);
}

.meta-lime {
  color: var(--lime);
  letter-spacing: 0.2em;
}

.title {
  font-size: 64px;
  line-height: 0.9;
  letter-spacing: 0.01em;
  margin: 0;
}

.form-card {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.field {
  display: flex;
  flex-direction: column;
}

/* .label and .input come from tokens.css */
.hint {
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--ink-400);
  margin: 6px 0 0;
}

.submit {
  width: 100%;
  height: 46px;
  margin-top: 4px;
}

.submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.state {
  text-align: center;
  padding: 14px;
  font-size: 13px;
  border-radius: var(--r-md);
}

.state-error {
  color: var(--down);
  background: rgba(255, 98, 98, 0.1);
  border: 1px solid rgba(255, 98, 98, 0.3);
}

@media (max-width: 767px) {
  .title {
    font-size: 40px;
  }
}
</style>
