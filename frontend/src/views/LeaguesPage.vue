<template>
  <AppShell>
    <div class="leagues">
      <!-- ===== Header ===== -->
      <header class="header">
        <div class="header-titles">
          <!-- Desktop shows the season meta line; mobile shows the product tag -->
          <div v-if="isMobile" class="meta mono">HM/LEAGUE</div>
          <div v-else class="meta meta-lime mono">◆ TEMPORADA 25/26 · J·32</div>
          <h1 class="title display">{{ isMobile ? "MIS LIGAS" : "MIS LIGAS." }}</h1>
        </div>

        <!-- Mobile: profile avatar. Desktop: action buttons. -->
        <!-- TODO Sprint 2: bind avatar initials to /api/v1/users/me profile. -->
        <Avatar v-if="isMobile" initials="AP" :size="36" />
        <div v-else class="header-actions">
          <button type="button" class="btn btn-secondary header-btn" @click="focusJoin">
            Unirse con código
          </button>
          <button type="button" class="btn btn-primary header-btn" @click="goCreate">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M7 2 V12 M2 7 H12" />
            </svg>
            Crear liga
          </button>
        </div>
      </header>

      <!-- ===== Mobile hero card: current matchday points ===== -->
      <div v-if="isMobile" class="hero">
        <PitchSVG stroke="rgba(8,9,11,0.18)" />
        <div class="hero-content">
          <div class="hero-label mono">◆ JORNADA EN CURSO · J·32</div>
          <div class="hero-row">
            <!-- TODO Sprint 2: bind to /api/v1/users/me/matchday-points endpoint. -->
            <div class="hero-points display tnum">+{{ matchdayPoints }}</div>
            <div class="hero-caption">
              <div class="hero-caption-main">puntos esta jornada</div>
              <div class="hero-caption-sub mono">+12 vs media · puesto ↑2</div>
            </div>
          </div>
        </div>
      </div>

      <!-- ===== Desktop stat strip ===== -->
      <div v-else class="stats">
        <div v-for="stat in stats" :key="stat.label" class="card stat-card">
          <div class="stat-label mono">{{ stat.label }}</div>
          <div class="stat-value display tnum" :class="{ lime: stat.lime }">{{ stat.value }}</div>
          <div class="stat-sub mono">{{ stat.sub }}</div>
        </div>
      </div>

      <!-- ===== Mobile action row: Crear liga + join code ===== -->
      <div v-if="isMobile" class="mobile-actions">
        <button type="button" class="btn btn-primary mobile-create" @click="goCreate">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M7 2 V12 M2 7 H12" />
          </svg>
          Crear liga
        </button>
        <form class="join-inline" @submit.prevent="joinLeague">
          <input
            ref="joinInputRef"
            v-model="inviteCode"
            class="input mono join-inline-input"
            placeholder="HML-CÓDIGO"
          />
          <button type="submit" class="btn btn-secondary join-inline-btn" :disabled="!inviteCode.trim() || joining">
            {{ joining ? "..." : "Unirse" }}
          </button>
        </form>
      </div>

      <!-- ===== Feedback ===== -->
      <div v-if="error" class="state state-error">{{ error }}</div>
      <div v-if="joinSuccess" class="state state-success">{{ joinSuccess }}</div>

      <!-- ===== Body ===== -->
      <div class="body">
        <!-- League list column -->
        <section class="list-col">
          <div class="list-head">
            <div class="meta mono">TUS LIGAS · {{ leagues.length }} ACTIVAS</div>
            <div class="meta meta-dim mono">{{ isMobile ? "FILTRAR ↓" : "POR ACTIVIDAD ↓" }}</div>
          </div>

          <!-- Loading -->
          <div v-if="loading" class="state state-muted">Cargando ligas...</div>

          <!-- Empty -->
          <div v-else-if="leagues.length === 0" class="state state-muted">
            <p class="empty-emoji">🏟️</p>
            <p class="empty-title">No estás en ninguna liga</p>
            <p class="empty-sub">Crea una liga o únete con un código de invitación</p>
          </div>

          <!-- League rows -->
          <div v-else class="rows">
            <div
              v-for="liga in leagueViews"
              :key="liga.id"
              class="card liga-card"
              :class="{ big: !isMobile, hot: liga.hot }"
              role="button"
              tabindex="0"
              @click="goLeague(liga.id)"
              @keydown.enter="goLeague(liga.id)"
            >
              <!-- "hot" treatment: lime border (via .hot) + corner ribbon.
                   TODO: gate on league.hasActiveBids once the backend exposes it;
                   first league is hardcoded hot for visual reference. -->
              <div v-if="liga.hot" class="ribbon mono">● PUJA EN VIVO</div>

              <LeagueAvatar :seed="liga.seed" :size="isMobile ? 44 : 56" :accent="liga.accent" />

              <div class="liga-info">
                <div class="liga-name">{{ liga.name }}</div>
                <div class="liga-meta mono">
                  <!-- TODO Sprint 2: members count not in /api/v1/leagues; using max_members. -->
                  <span>{{ liga.members }} MGRS</span>
                  <span class="liga-meta-dot">·</span>
                  <!-- TODO Sprint 2: current matchday not in /api/v1/leagues. -->
                  <span>JOR·{{ liga.matchday }}</span>
                  <span class="liga-meta-dot">·</span>
                  <span>€{{ liga.budget }}</span>
                </div>
              </div>

              <!-- TODO Sprint 2: position/total need a standings call per league. -->
              <div class="liga-pos">
                <span class="liga-pos-num display tnum" :class="{ lime: liga.position <= 3 }">
                  {{ liga.position.toString().padStart(2, "0") }}
                </span>
                <span class="liga-pos-total mono">/{{ liga.total }}</span>
              </div>

              <svg class="liga-chevron" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M6 4 L10 8 L6 12" />
              </svg>
            </div>
          </div>
        </section>

        <!-- Desktop side column: create CTA + recent activity -->
        <aside v-if="!isMobile" class="side-col">
          <div class="meta mono">¿NO TIENES LIGA?</div>

          <div class="card cta-card">
            <div class="cta-avatar">
              <LeagueAvatar :seed="3" :size="64" accent="var(--lime)" />
            </div>
            <h3 class="cta-title display">CREA TU LIGA</h3>
            <p class="cta-text">Invita a tus amigos con un código. Hasta 20 mánagers por liga.</p>
            <button type="button" class="btn btn-primary cta-create" @click="goCreate">+ Nueva liga</button>

            <div class="divider" />

            <div class="meta cta-join-label mono">O ÚNETE A UNA</div>
            <form class="cta-join" @submit.prevent="joinLeague">
              <input
                ref="joinInputRef"
                v-model="inviteCode"
                class="input mono"
                placeholder="HML-CÓDIGO"
              />
              <button type="submit" class="btn btn-secondary cta-join-btn" :disabled="!inviteCode.trim() || joining">
                {{ joining ? "..." : "Unirse" }}
              </button>
            </form>
          </div>

          <!-- Recent activity is mocked pending a Sprint 3 events endpoint. -->
          <div class="card activity-card">
            <div class="meta activity-label mono">● ACTIVIDAD RECIENTE</div>
            <div class="activity-list">
              <div v-for="(item, i) in recentActivityMock" :key="i" class="activity-row">
                <span>
                  <span class="activity-actor" :style="{ color: item.color }">{{ item.actor }}</span>
                  {{ item.action }}
                </span>
                <span class="activity-time mono">{{ item.time }}</span>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
  </AppShell>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import AppShell from "@/design-system/AppShell.vue";
import Avatar from "@/design-system/primitives/Avatar.vue";
import LeagueAvatar from "@/design-system/primitives/LeagueAvatar.vue";
import PitchSVG from "@/design-system/primitives/PitchSVG.vue";
import api from "@/lib/api";

// Backend shape returned by GET /api/v1/leagues (mirrors models.League).
// UI-only fields are layered on top as a computed view-model (LeagueView),
// never added to this fetched type.
interface League {
  id: number;
  name: string;
  invite_code: string;
  max_members: number;
  budget_per_user: number;
  market_close_time: string;
  created_by: number;
}

// UI view-model: derived presentation fields on top of the fetched League.
interface LeagueView {
  id: number;
  name: string;
  budget: string;
  seed: number;
  accent: string;
  members: number;
  matchday: number;
  position: number;
  total: number;
  hot: boolean;
}

interface StatCard {
  label: string;
  value: string;
  sub: string;
  lime: boolean;
}

interface ActivityItem {
  actor: string;
  action: string;
  time: string;
  color: string;
}

const router = useRouter();

const leagues = ref<League[]>([]);
const loading = ref(true);
const error = ref("");
const inviteCode = ref("");
const joining = ref(false);
const joinSuccess = ref("");
const isMobile = ref(false);
const joinInputRef = ref<HTMLInputElement | null>(null);

// Real market data for the stats cards
const userActiveBids = ref(0);
const maxBidsPerUser = ref(5);
const marketClosesAt = ref<string | null>(null);
const marketIsOpen = ref(false);

// TODO Sprint 2: bind to /api/v1/users/me/matchday-points endpoint.
const matchdayPoints = 87;

// Accent palette cycled per league so the same league always looks the same.
const ACCENTS = ["var(--lime)", "var(--pos-fwd)", "var(--pos-def)", "var(--pos-gk)"];

// Recent activity is mocked pending a Sprint 3 events/feed endpoint.
const recentActivityMock: ActivityItem[] = [
  { actor: "Lucía R.", action: "pujó por Etta Eyong", time: "2m", color: "var(--lime)" },
  { actor: "Carlos M.", action: "subió a 1°", time: "12m", color: "var(--pos-fwd)" },
  { actor: "Andrés P.", action: "vendió Iván Romero", time: "38m", color: "var(--pos-def)" },
];

// Format budget_per_user (euros) into the compact "100M" label the design uses.
function formatBudget(amount: number): string {
  return `${Math.round(amount / 1_000_000)}M`;
}

const leagueViews = computed<LeagueView[]>(() =>
  leagues.value.map((l, index) => ({
    id: l.id,
    name: l.name,
    budget: formatBudget(l.budget_per_user),
    // Stable seed derived from the league id so the avatar pattern is consistent.
    seed: l.id % 5,
    accent: ACCENTS[index % ACCENTS.length] ?? "var(--lime)",
    // TODO Sprint 2: real member count needs a dedicated field/endpoint; max_members for now.
    members: l.max_members,
    // TODO Sprint 2: current matchday number not in /api/v1/leagues.
    matchday: 32,
    // TODO Sprint 2: position/total require a standings call per league.
    position: 0,
    total: l.max_members,
    // TODO: gate on league.hasActiveBids once the backend exposes it. First
    // league hardcoded hot for visual reference of the "live bid" treatment.
    hot: index === 0,
  })),
);

const marketCloseLabel = computed(() => {
  if (!marketClosesAt.value) return marketIsOpen.value ? "Mercado abierto" : "Mercado cerrado";
  const ms = new Date(marketClosesAt.value).getTime() - Date.now();
  if (ms <= 0) return "Mercado cerrado";
  const totalSec = Math.floor(ms / 1000);
  const h = Math.floor(totalSec / 3600);
  const m = Math.floor((totalSec % 3600) / 60);
  const s = totalSec % 60;
  const pad = (n: number) => n.toString().padStart(2, "0");
  return `Cierra ${pad(h)}:${pad(m)}:${pad(s)}`;
});

const stats = computed<StatCard[]>(() => [
  { label: "LIGAS ACTIVAS", value: String(leagues.value.length), sub: "en juego", lime: false },
  // TODO Sprint 3: total points needs a per-user points endpoint.
  { label: "PUNTOS TOTALES", value: "—", sub: "pendiente", lime: false },
  // TODO Sprint 3: best position needs a standings call per league.
  { label: "MEJOR PUESTO", value: "—", sub: "pendiente", lime: true },
  { label: "PUJAS ACTIVAS", value: `${userActiveBids.value}/${maxBidsPerUser.value}`, sub: marketCloseLabel.value, lime: false },
]);

function updateIsMobile(): void {
  isMobile.value = window.innerWidth < 768;
}

interface MarketStatusResponse {
  status: string;
  data: {
    league_id: number;
    is_open: boolean;
    closes_at: string;
    active_listings: number;
    your_active_bids: number;
    max_bids_per_user: number;
  } | null;
}

async function fetchLeagues(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    leagues.value = await api.get<League[]>("/api/v1/leagues");
    // Fetch market status from the first league to populate the stats card
    if (leagues.value.length > 0) {
      fetchMarketStatus(leagues.value[0]!.id);
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Error al cargar ligas";
  } finally {
    loading.value = false;
  }
}

async function fetchMarketStatus(leagueId: number): Promise<void> {
  try {
    const res = await api.get<MarketStatusResponse>(`/api/v1/leagues/${leagueId}/market/status`);
    if (res.data) {
      userActiveBids.value = res.data.your_active_bids;
      maxBidsPerUser.value = res.data.max_bids_per_user;
      marketClosesAt.value = res.data.closes_at;
      marketIsOpen.value = res.data.is_open;
    }
  } catch {
    // non-blocking — stats card will show defaults
  }
}

// Existing join flow — unchanged: POST /api/v1/leagues/join then open the league.
async function joinLeague(): Promise<void> {
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

// Existing create-league flow.
function goCreate(): void {
  router.push("/leagues/new");
}

// Existing league detail route.
function goLeague(id: number): void {
  router.push(`/leagues/${id}`);
}

function focusJoin(): void {
  joinInputRef.value?.focus();
}

onMounted(() => {
  updateIsMobile();
  window.addEventListener("resize", updateIsMobile);
  fetchLeagues();
});

onUnmounted(() => {
  window.removeEventListener("resize", updateIsMobile);
});
</script>

<style scoped>
.leagues {
  display: flex;
  flex-direction: column;
  gap: 24px;
  color: var(--ink-100);
}

/* ===== Header ===== */
.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}

.header-titles {
  display: flex;
  flex-direction: column;
  gap: 10px;
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

.meta-dim {
  color: var(--ink-400);
  letter-spacing: 0.1em;
}

.title {
  font-size: 92px;
  line-height: 0.9;
  letter-spacing: 0.01em;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.header-btn {
  height: 44px;
}

/* ===== Mobile hero ===== */
.hero {
  position: relative;
  overflow: hidden;
  border-radius: var(--r-lg);
  padding: 18px;
  color: var(--ink-900);
  background: linear-gradient(135deg, var(--lime) 0%, var(--lime-deep) 100%);
}

.hero-content {
  position: relative;
}

.hero-label {
  font-size: 10px;
  letter-spacing: 0.18em;
  font-weight: 600;
}

.hero-row {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  margin-top: 8px;
}

.hero-points {
  font-size: 64px;
  line-height: 0.9;
  font-weight: 400;
}

.hero-caption {
  padding-bottom: 6px;
}

.hero-caption-main {
  font-size: 12px;
  font-weight: 600;
}

.hero-caption-sub {
  font-size: 10px;
  opacity: 0.7;
}

/* ===== Desktop stat strip ===== */
.stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.stat-card {
  padding: 18px;
}

.stat-label {
  font-size: 10px;
  letter-spacing: 0.15em;
  color: var(--ink-400);
}

.stat-value {
  font-size: 44px;
  line-height: 1;
  margin-top: 6px;
  color: var(--ink-100);
}

.stat-value.lime {
  color: var(--lime);
}

.stat-sub {
  font-size: 10px;
  color: var(--ink-300);
  margin-top: 4px;
}

/* ===== Mobile action row ===== */
.mobile-actions {
  display: flex;
  gap: 8px;
}

.mobile-create {
  flex: 1.4;
  height: 44px;
}

.join-inline {
  flex: 2;
  display: flex;
  background: var(--ink-800);
  border: 1px solid var(--ink-700);
  border-radius: var(--r-sm);
  overflow: hidden;
}

.join-inline-input {
  border: none;
  border-radius: 0;
  font-size: 12px;
  height: 42px;
  background: transparent;
}

.join-inline-input:focus {
  border: none;
}

.join-inline-btn {
  border-radius: 0;
  border: none;
  border-left: 1px solid var(--ink-700);
  height: 42px;
  padding: 0 14px;
  font-size: 12px;
}

/* ===== Body layout ===== */
.body {
  display: grid;
  grid-template-columns: 1.6fr 1fr;
  gap: 24px;
}

.list-col {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.list-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.rows {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

/* ===== League card ===== */
.liga-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: background 0.15s, border-color 0.15s;
}

.liga-card:hover {
  background: var(--ink-700);
}

.liga-card.big {
  padding: 20px;
  gap: 16px;
}

.liga-card.hot {
  border-color: var(--lime);
}

.ribbon {
  position: absolute;
  top: 8px;
  right: 8px;
  background: var(--lime);
  color: var(--ink-900);
  font-size: 8px;
  letter-spacing: 0.15em;
  padding: 2px 6px;
  border-radius: 2px;
  font-weight: 700;
}

.liga-info {
  flex: 1;
  min-width: 0;
}

.liga-name {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 4px;
}

.liga-card.big .liga-name {
  font-size: 16px;
}

.liga-meta {
  font-size: 10px;
  color: var(--ink-400);
  letter-spacing: 0.1em;
  display: flex;
  gap: 10px;
}

.liga-meta-dot {
  color: var(--ink-600);
}

.liga-pos {
  display: flex;
  align-items: baseline;
  gap: 6px;
  text-align: right;
}

.liga-pos-num {
  font-size: 32px;
  line-height: 1;
  color: var(--ink-100);
}

.liga-card.big .liga-pos-num {
  font-size: 38px;
}

.liga-pos-num.lime {
  color: var(--lime);
}

.liga-pos-total {
  font-size: 11px;
  color: var(--ink-400);
}

.liga-chevron {
  color: var(--ink-400);
  flex-shrink: 0;
}

/* ===== Side column ===== */
.side-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.cta-card {
  padding: 24px;
  text-align: center;
  border-style: dashed;
}

.cta-avatar {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
}

.cta-title {
  font-size: 26px;
  margin: 0 0 8px;
  letter-spacing: 0.02em;
}

.cta-text {
  font-size: 12px;
  color: var(--ink-300);
  margin: 0 0 16px;
  line-height: 1.5;
}

.cta-create {
  width: 100%;
  height: 42px;
}

.cta-card .divider {
  margin: 20px 0;
}

.cta-join-label {
  margin-bottom: 8px;
}

.cta-join {
  display: flex;
  gap: 8px;
}

.cta-join-btn {
  padding: 0 18px;
}

/* ===== Activity ===== */
.activity-card {
  padding: 14px;
}

.activity-label {
  color: var(--ink-400);
  margin-bottom: 10px;
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 12px;
}

.activity-row {
  display: flex;
  justify-content: space-between;
}

.activity-time {
  color: var(--ink-400);
  font-size: 10px;
}

/* ===== States ===== */
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

.state-success {
  color: var(--up);
}

.empty-emoji {
  font-size: 36px;
  margin: 0 0 12px;
}

.empty-title {
  color: var(--ink-100);
  font-weight: 600;
  margin: 0 0 4px;
}

.empty-sub {
  color: var(--ink-400);
  font-size: 13px;
  margin: 0;
}

/* ===== Mobile ===== */
@media (max-width: 767px) {
  .leagues {
    gap: 16px;
  }

  .title {
    font-size: 40px;
  }

  .body {
    display: block;
  }
}
</style>
