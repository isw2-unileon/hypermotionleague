<template>
  <AppShell>
    <div class="detail">
      <!-- Back -->
      <button type="button" class="back mono" @click="router.push('/leagues')">
        ← Mis Ligas
      </button>

      <!-- Loading -->
      <div v-if="loading" class="state state-muted">Cargando...</div>

      <!-- Error -->
      <div v-else-if="error" class="state state-error">{{ error }}</div>

      <!-- Content -->
      <template v-else-if="league">
        <!-- Header -->
        <header class="header">
          <!-- nº miembros is real; TEMP 25/26 · J·32 is static pending a
               matchday/season endpoint — TODO Sprint 2: bind J·xx and season. -->
          <div class="meta meta-lime mono">
            ◆ {{ members.length }} MÁNAGERS · TEMP 25/26 · J·32
          </div>
          <h1 class="title display">{{ league.name }}</h1>
        </header>

        <!-- Info strip -->
        <div class="info">
          <div class="card info-cell">
            <div class="info-label mono">PRESUPUESTO</div>
            <div class="info-value display tnum">{{ budgetCompact(league.budget_per_user) }}</div>
            <div class="info-sub mono">{{ formatBudget(league.budget_per_user) }}</div>
          </div>
          <div class="card info-cell">
            <div class="info-label mono">USUARIOS</div>
            <div class="info-value display tnum">{{ members.length }}<span class="info-slash">/{{ league.max_members }}</span></div>
            <div class="info-sub mono">plazas ocupadas</div>
          </div>
          <div class="card info-cell">
            <div class="info-label mono">CIERRE MERCADO</div>
            <div class="info-value display tnum lime">{{ formatCloseTime(league.market_close_time) }}</div>
            <div class="info-sub mono">hora diaria</div>
          </div>
        </div>

        <!-- Invite code -->
        <div class="card invite">
          <label class="label" for="invite-code">Código de invitación</label>
          <div class="invite-row">
            <input
              id="invite-code"
              class="input mono"
              :value="league.invite_code"
              readonly
            />
            <button type="button" class="btn btn-secondary invite-btn" @click="copyCode">
              {{ copied ? "Copiado!" : "Copiar" }}
            </button>
          </div>
        </div>

        <!-- Members -->
        <section class="members">
          <div class="meta mono members-head">MIEMBROS · {{ members.length }}</div>
          <div class="rows">
            <div
              v-for="(member, index) in members"
              :key="member.id"
              class="member-row card"
              :class="{ me: isMe(member) }"
            >
              <img
                v-if="member.avatar_url"
                :src="member.avatar_url"
                :alt="memberName(member)"
                class="member-photo"
                :style="{ borderColor: rowColor(index) }"
              />
              <Avatar
                v-else
                :initials="initialsFor(member)"
                :size="40"
                :color="rowColor(index)"
              />

              <div class="member-info">
                <div class="member-name">
                  {{ memberName(member) }}
                  <span v-if="member.role === 'owner'" class="crown" title="Propietario">👑</span>
                  <span v-if="isMe(member)" class="tag tag-lime me-tag">TÚ</span>
                </div>
                <div class="member-budget mono">{{ formatBudget(member.budget) }}</div>
              </div>

              <span class="member-role mono">{{ member.role }}</span>

              <template v-if="isOwner && member.role !== 'owner'">
                <button
                  v-if="confirmKickId !== member.user_id"
                  class="btn btn-ghost kick-btn mono"
                  :disabled="kickingId === member.user_id"
                  @click="confirmKickId = member.user_id"
                >
                  Expulsar
                </button>
                <div v-else class="kick-confirm">
                  <button
                    class="btn kick-confirm-btn mono"
                    :disabled="kickingId === member.user_id"
                    @click="kickMember(member.user_id)"
                  >
                    {{ kickingId === member.user_id ? '…' : 'Confirmar' }}
                  </button>
                  <button class="btn btn-secondary mono" @click="confirmKickId = null">
                    Cancelar
                  </button>
                </div>
              </template>
            </div>
          </div>
        </section>

        <!-- Delete (owner only) -->
        <div v-if="isOwner" class="danger">
          <button
            v-if="!confirmDelete"
            type="button"
            class="btn btn-ghost danger-btn"
            @click="confirmDelete = true"
          >
            Eliminar liga
          </button>
          <div v-else class="danger-confirm">
            <button
              type="button"
              class="btn danger-confirm-btn"
              :disabled="deleting"
              @click="deleteLeague"
            >
              {{ deleting ? "Eliminando..." : "Confirmar eliminación" }}
            </button>
            <button type="button" class="btn btn-secondary" @click="confirmDelete = false">
              Cancelar
            </button>
          </div>
        </div>
      </template>
    </div>
    <Toast
      :visible="showCreatedToast"
      message="¡Liga creada! Se te han asignado 15 jugadores. Mira tu plantilla en Mi Equipo."
      variant="success"
      @close="showCreatedToast = false"
    />
  </AppShell>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import AppShell from "@/design-system/AppShell.vue";
import Avatar from "@/design-system/primitives/Avatar.vue";
import Toast from "@/design-system/Toast.vue";
import { currentUserId } from "@/lib/auth";
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


const showCreatedToast = ref(false);


const league = ref<League | null>(null);
const members = ref<Member[]>([]);
const loading = ref(true);
const error = ref("");
const copied = ref(false);
const confirmDelete = ref(false);
const deleting = ref(false);
const kickingId = ref<number | null>(null);
const confirmKickId = ref<number | null>(null);

// Current user comes from the existing auth helper (decodes the JWT).
const meId = currentUserId();

const isOwner = computed(() => league.value?.created_by === meId);

// Row accent cycled by index, per the design system position palette.
const ROW_COLORS = ["var(--pos-mid)", "var(--pos-def)", "var(--pos-fwd)", "var(--pos-gk)"];
function rowColor(index: number): string {
  return ROW_COLORS[index % ROW_COLORS.length] ?? "var(--pos-mid)";
}

function isMe(member: Member): boolean {
  return meId !== 0 && member.user_id === meId;
}

// Always prefer real identity — never a "Jugador #N" placeholder.
function memberName(member: Member): string {
  return member.display_name || member.username;
}

// Initials = first letter of the first two words of display_name,
// falling back to the first two chars of username.
function initialsFor(member: Member): string {
  const name = member.display_name?.trim() ?? "";
  if (name) {
    const parts = name.split(/\s+/);
    const initials = parts
      .slice(0, 2)
      .map((p) => p.charAt(0))
      .join("");
    return initials.toUpperCase();
  }
  return member.username.slice(0, 2).toUpperCase();
}

function formatBudget(amount: number): string {
  return new Intl.NumberFormat("es-ES", {
    style: "currency",
    currency: "EUR",
    maximumFractionDigits: 0,
  }).format(amount);
}

function budgetCompact(amount: number): string {
  return `${Math.round(amount / 1_000_000)}M`;
}

function formatCloseTime(time: string): string {
  return time.split(".")[0] ?? time;
}

async function copyCode() {
  if (!league.value) return;
  await navigator.clipboard.writeText(league.value.invite_code);
  copied.value = true;
  setTimeout(() => (copied.value = false), 2000);
}

async function kickMember(userId: number) {
  kickingId.value = userId;
  try {
    await api.delete(`/api/v1/leagues/${route.params.id}/members/${userId}`);
    members.value = members.value.filter(m => m.user_id !== userId);
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Error al expulsar al miembro";
  } finally {
    kickingId.value = null;
    confirmKickId.value = null;
  }
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

// Data wiring unchanged: same two GETs, same routes.
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

  if (route.query.created === "1") {
    showCreatedToast.value = true;
    router.replace({ query: {} });
    setTimeout(() => {
      showCreatedToast.value = false;
    }, 6000);
  }
});
</script>

<style scoped>
.detail {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 760px;
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

/* Header */
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

/* Info strip */
.info {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}

.info-cell {
  padding: 16px;
}

.info-label {
  font-size: 10px;
  letter-spacing: 0.15em;
  color: var(--ink-400);
}

.info-value {
  font-size: 30px;
  line-height: 1;
  margin-top: 6px;
  color: var(--ink-100);
}

.info-value.lime {
  color: var(--lime);
}

.info-slash {
  font-size: 16px;
  color: var(--ink-400);
}

.info-sub {
  font-size: 10px;
  color: var(--ink-300);
  margin-top: 4px;
}

/* Invite */
.invite {
  padding: 18px;
}

.invite-row {
  display: flex;
  gap: 8px;
}

.invite-row .input {
  letter-spacing: 0.18em;
  color: var(--lime);
}

.invite-btn {
  padding: 0 18px;
  white-space: nowrap;
}

/* Members */
.members {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.members-head {
  letter-spacing: 0.15em;
}

.rows {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.member-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-left: 3px solid transparent;
}

.member-row.me {
  border-left: 3px solid var(--lime);
  background: var(--lime-glow);
}

.member-photo {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  border: 1.5px solid var(--lime);
  flex-shrink: 0;
}

.member-info {
  flex: 1;
  min-width: 0;
}

.member-name {
  font-size: 14px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;
}

.crown {
  font-size: 12px;
}

.me-tag {
  font-size: 9px;
  padding: 1px 6px;
}

.member-budget {
  font-size: 11px;
  color: var(--ink-400);
  margin-top: 2px;
}

.member-role {
  font-size: 10px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--ink-400);
}

.kick-btn {
  margin-left: auto;
  font-size: 11px;
  color: var(--pos-fwd);
  border-color: var(--pos-fwd);
  padding: 2px 10px;
}

.kick-confirm {
  margin-left: auto;
  display: flex;
  gap: 6px;
}

.kick-confirm-btn {
  font-size: 11px;
  background: var(--pos-fwd);
  color: var(--ink-900);
  padding: 2px 10px;
}

/* Danger */
.danger {
  margin-top: 4px;
}

.danger-btn {
  width: 100%;
  color: var(--down);
  border-color: rgba(255, 98, 98, 0.3);
}

.danger-btn:hover {
  background: rgba(255, 98, 98, 0.1);
}

.danger-confirm {
  display: flex;
  gap: 8px;
}

.danger-confirm-btn {
  flex: 1;
  background: var(--down);
  color: #fff;
}

.danger-confirm-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

@media (max-width: 767px) {
  .title {
    font-size: 40px;
  }

  .info {
    grid-template-columns: 1fr;
  }
}
</style>
