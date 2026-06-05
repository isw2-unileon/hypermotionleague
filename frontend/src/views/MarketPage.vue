<template>
  <div class="market frame">
    <div class="halftone-bg"></div>

    <div class="shell" :class="{ mobile: isMobile }">
      <!--
        TODO: replace this inline sidebar with <AppShell> once D2.1 lands.
        AppShell.vue does not exist yet, so per the brief the desktop nav is a
        minimal inline sidebar matching Mercado.jsx, and mobile uses the TabBar
        primitive at the bottom.
      -->
      <aside v-if="!isMobile" class="sidebar">
        <Logo :size="26" />
        <div class="sidebar-gap"></div>
        <button
          v-for="item in navItems"
          :key="item.to"
          class="nav-item"
          :class="{ active: item.active }"
          @click="go(item.to)"
        >
          <span class="nav-dot" :class="{ active: item.active }"></span>
          <span class="nav-label">{{ item.label }}</span>
          <span v-if="item.active && activeBidCount > 0" class="tag tag-lime nav-badge">{{ activeBidCount }}</span>
        </button>
      </aside>

      <div class="main">
        <StatusBar v-if="isMobile" />

        <!-- header -->
        <header class="page-head" :class="{ mobile: isMobile }">
          <div class="head-left">
            <div class="mono kicker">
              <span v-if="!isMobile">◆ </span>{{ headerMeta }}
            </div>
            <h1 class="display title">MERCADO<span v-if="!isMobile">.</span></h1>
            <!-- League picker: there is no active-league concept in the app yet,
                 so the market resolves its league from ?league=<id> (else the
                 user's first league). This selector lets the user switch without
                 hardcoding an id. TODO: drive from a shared active-league store. -->
            <div v-if="leagues.length > 1" class="league-pick">
              <span class="mono pick-label">LIGA</span>
              <select class="pick-select mono" :value="leagueId ?? ''" @change="onLeagueChange">
                <option v-for="l in leagues" :key="l.id" :value="l.id">{{ l.name }}</option>
              </select>
            </div>
          </div>

          <div class="head-right">
            <div class="saldo">
              <div class="mono col-label">SALDO</div>
              <div class="display tnum saldo-value">{{ millions(balance) }}</div>
            </div>
            <div v-if="!isMobile && nextCloseLabel" class="card cierra-card">
              <span class="pulse"></span>
              <div>
                <div class="mono col-label">CIERRA EN</div>
                <div class="display tnum cierra-value">{{ nextCloseLabel }}</div>
              </div>
            </div>
          </div>
        </header>

        <!-- in-page tabs -->
        <div class="ptabs" :class="{ mobile: isMobile }">
          <button class="ptab mono" :class="{ active: activeTab === 'mercado' }" @click="activeTab = 'mercado'">
            MERCADO · {{ mercadoCount }}
          </button>
          <button class="ptab mono" :class="{ active: activeTab === 'pujas' }" @click="activeTab = 'pujas'">
            MIS PUJAS · {{ activeBidCount }}/{{ MAX_ACTIVE_BIDS }}
          </button>
          <div v-if="!isMobile && activeTab === 'mercado'" class="tab-tools">
            <input v-model="search" class="input search" placeholder="Buscar jugador…" />
          </div>
        </div>

        <!-- mobile countdown banner -->
        <div v-if="isMobile && nextCloseLabel" class="cierra-banner card">
          <span class="pulse"></span>
          <span class="mono banner-label">CIERRA EN</span>
          <span class="display tnum banner-value">{{ nextCloseLabel }}</span>
        </div>

        <!-- position filters (mercado tab) -->
        <div v-if="activeTab === 'mercado'" class="filters">
          <button
            v-for="f in filters"
            :key="f.label"
            class="filter-pill mono"
            :class="{ active: posFilter === f.value }"
            @click="posFilter = f.value"
          >
            {{ f.label }}
          </button>
        </div>

        <!-- market closed banner -->
        <div v-if="!loading && !marketOpen && leagueId != null" class="closed-banner card">
          <span class="closed-icon">🔒</span>
          <div>
            <div class="mono closed-label">MERCADO CERRADO</div>
            <div class="mono closed-sub">No se pueden realizar pujas en este momento.</div>
          </div>
        </div>

        <!-- content -->
        <div class="content">
          <div v-if="loading" class="state mono">CARGANDO MERCADO…</div>
          <div v-else-if="error" class="state error">{{ error }}</div>
          <div v-else-if="leagueId == null" class="state mono">
            No perteneces a ninguna liga todavía.
            <button class="btn btn-secondary join-btn" @click="go('/leagues')">Ver mis ligas</button>
          </div>

          <!-- MERCADO -->
          <template v-else-if="activeTab === 'mercado'">
            <div v-if="filteredListings.length === 0" class="state mono">
              {{ mercadoCount === 0 ? "No hay jugadores en el mercado ahora mismo." : "Ningún jugador coincide con el filtro." }}
            </div>
            <div v-else class="list">
              <PlayerCard
                v-for="l in filteredListings"
                :key="l.id"
                :listing="l"
                :now="now"
                :mobile="isMobile"
                :mine="myBidByListing.has(l.id)"
                @bid="openBid"
              />
            </div>
          </template>

          <!-- MIS PUJAS -->
          <template v-else>
            <!-- Status summary. PUJAS ACTIVAS + TOTAL COMPROMETIDO are directly
                 derivable. GANANDO/SUPERADA are derived from each listing's
                 visible highest_bid (the backend has no per-bid status field —
                 Phase 0 #7), so they reflect "is my bid the current top?". -->
            <div class="summary" :class="{ mobile: isMobile }">
              <div class="card summary-card">
                <div class="mono col-label">PUJAS ACTIVAS</div>
                <div class="display tnum summary-value">{{ activeBidCount }}</div>
                <div class="mono summary-sub">de {{ MAX_ACTIVE_BIDS }} disponibles</div>
              </div>
              <div class="card summary-card">
                <div class="mono col-label">GANANDO</div>
                <div class="display tnum summary-value lime">{{ leadingCount }}</div>
                <div class="mono summary-sub">según puja top visible</div>
              </div>
              <div class="card summary-card">
                <div class="mono col-label">SUPERADA</div>
                <div class="display tnum summary-value down">{{ outbidCount }}</div>
                <div class="mono summary-sub">por debajo del top</div>
              </div>
              <div class="card summary-card">
                <div class="mono col-label">TOTAL COMPROMETIDO</div>
                <div class="display tnum summary-value">{{ millions(committed) }}</div>
                <div class="mono summary-sub">de {{ millions(balance) }} saldo</div>
              </div>
            </div>

            <div v-if="bids.length === 0" class="state mono">Aún no has pujado por ningún jugador.</div>
            <div v-else class="list">
              <div
                v-for="b in bids"
                :key="b.id"
                class="bid-row card"
                :style="{ borderLeft: `3px solid ${rowAccent(b)}` }"
              >
                <PlayerPhoto :team="teamShort(b.player.team_name)" :color="teamColor(b.player.team_name)" :size="isMobile ? 44 : 56" />
                <div class="bid-id">
                  <div class="bid-id-top">
                    <span class="pos" :class="positionMeta(b.player.position).cssClass">{{ positionMeta(b.player.position).label }}</span>
                    <TeamCrest :short="teamShort(b.player.team_name)" :color="teamColor(b.player.team_name)" :size="18" />
                    <span class="bid-name">{{ fullName(b.player) }}</span>
                    <span
                      v-if="statusInfo(b)"
                      class="mono status-chip"
                      :style="{ color: statusInfo(b)!.color, borderColor: statusInfo(b)!.color }"
                    >● {{ statusInfo(b)!.label }}</span>
                  </div>
                  <div class="mono bid-sub">
                    <template v-if="bidCountdown(b)">Cierra en <span class="bid-sub-strong">{{ bidCountdown(b) }}</span></template>
                    <template v-else>Subasta cerrada</template>
                  </div>
                </div>
                <div class="col bid-amount-col">
                  <div class="mono col-label">TU PUJA</div>
                  <div class="display tnum bid-amount">{{ millions(b.amount) }}</div>
                </div>
                <div class="col">
                  <div class="mono col-label">PUJA TOP</div>
                  <div class="display tnum bid-top" :style="{ color: rowAccent(b) }">
                    {{ bidTop(b) != null ? millions(bidTop(b)!) : "—" }}
                  </div>
                </div>
                <button class="btn btn-secondary row-btn" :disabled="!isRaisable(b)" @click="openRaise(b)">Subir</button>
                <button class="btn btn-ghost row-btn cancel-btn" @click="cancelBid(b)">Cancelar</button>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>

    <div v-if="isMobile" class="mobile-tabbar">
      <TabBar active="mercado" @select="onBottomTab" />
    </div>

    <BidModal
      v-if="modalListing && leagueId != null && league"
      :key="`${modalListing.id}-${modalReplaceBidId ?? 0}`"
      :listing="modalListing"
      :league-id="leagueId"
      :league-name="league.name"
      :balance="balance"
      :committed="modalCommitted"
      :active-bid-count="activeBidCount"
      :now="now"
      :replace-bid-id="modalReplaceBidId"
      @placed="onPlaced"
      @close="closeModal"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import api from "@/lib/api";
import { currentUserId } from "@/lib/auth";
import Logo from "@/design-system/primitives/Logo.vue";
import StatusBar from "@/design-system/primitives/StatusBar.vue";
import TabBar from "@/design-system/primitives/TabBar.vue";
import PlayerPhoto from "@/design-system/primitives/PlayerPhoto.vue";
import TeamCrest from "@/design-system/primitives/TeamCrest.vue";
import PlayerCard from "@/design-system/components/PlayerCard.vue";
import BidModal from "@/design-system/components/BidModal.vue";
import {
  type ApiEnvelope,
  type BidWithDetails,
  type LeagueMember,
  type LeagueSummary,
  type Matchday,
  type MarketListingWithDetails,
  type MarketStatus,
  type PlayerPosition,
  MAX_ACTIVE_BIDS,
  formatCountdown,
  fullName,
  millions,
  msUntil,
  positionMeta,
  teamColor,
  teamShort,
} from "@/lib/market";

const route = useRoute();
const router = useRouter();
const myUserId = currentUserId();

const leagues = ref<LeagueSummary[]>([]);
const leagueId = ref<number | null>(null);
const league = ref<LeagueSummary | null>(null);
const jornada = ref<number | null>(null);
const balance = ref(0);
const listings = ref<MarketListingWithDetails[]>([]);
const bids = ref<BidWithDetails[]>([]);

const marketStatus = ref<MarketStatus | null>(null);
const marketOpen = computed(() => marketStatus.value?.is_open ?? true);

const loading = ref(true);
const error = ref("");

const activeTab = ref<"mercado" | "pujas">("mercado");
const posFilter = ref<PlayerPosition | null>(null);
const search = ref("");

const now = ref(Date.now());
const isMobile = ref(false);
let clockTimer: number | undefined;

const modalListing = ref<MarketListingWithDetails | null>(null);
const modalReplaceBidId = ref<number | null>(null);

const filters: { label: string; value: PlayerPosition | null }[] = [
  { label: "TODOS", value: null },
  { label: "POR", value: "GK" },
  { label: "DEF", value: "DEF" },
  { label: "MED", value: "MID" },
  { label: "DEL", value: "FWD" },
];

const navItems = computed(() => [
  { label: "Mis Ligas", to: "/leagues", active: false },
  { label: "Clasificación", to: "/standings", active: false },
  { label: "Mi Equipo", to: "/team", active: false },
  { label: "Mercado", to: "/market", active: true },
]);

const headerMeta = computed(() => {
  const parts: string[] = [];
  if (league.value) parts.push(league.value.name.toUpperCase());
  if (jornada.value != null) parts.push(`J·${jornada.value}`);
  if (!isMobile.value) parts.push(`${mercadoCount.value} JUGADORES DISPONIBLES`);
  return parts.join(" · ");
});

const listingById = computed(() => {
  const map = new Map<number, MarketListingWithDetails>();
  for (const l of listings.value) map.set(l.id, l);
  return map;
});

const myBidByListing = computed(() => {
  const map = new Map<number, BidWithDetails>();
  for (const b of bids.value) map.set(b.listing_id, b);
  return map;
});

const committed = computed(() => bids.value.reduce((sum, b) => sum + b.amount, 0));

const filteredListings = computed(() => {
  const q = search.value.trim().toLowerCase();
  return listings.value.filter((l) => {
    if (posFilter.value && l.player.position !== posFilter.value) return false;
    if (q && !fullName(l.player).toLowerCase().includes(q)) return false;
    return true;
  });
});

const mercadoCount = computed(() => listings.value.length);
const activeBidCount = computed(() => bids.value.length);

// Banner countdown: use the real closes_at from market status endpoint.
// Falls back to the soonest-closing listing if status is unavailable.
const nextCloseMs = computed(() => {
  if (marketStatus.value?.closes_at) {
    const ms = msUntil(marketStatus.value.closes_at, now.value);
    if (ms > 0) return ms;
  }
  let best = Number.POSITIVE_INFINITY;
  for (const l of listings.value) {
    const ms = msUntil(l.expires_at, now.value);
    if (ms > 0 && ms < best) best = ms;
  }
  return Number.isFinite(best) ? best : null;
});
const nextCloseLabel = computed(() =>
  nextCloseMs.value == null ? null : formatCountdown(nextCloseMs.value),
);

// Derived bid status from the listing's visible top bid (real data).
type DerivedStatus = "leading" | "outbid" | "unknown";
function bidStatus(b: BidWithDetails): DerivedStatus {
  const listing = listingById.value.get(b.listing_id);
  if (!listing || listing.highest_bid == null) return "unknown";
  return b.amount >= listing.highest_bid ? "leading" : "outbid";
}
function statusInfo(b: BidWithDetails): { label: string; color: string } | null {
  const s = bidStatus(b);
  if (s === "leading") return { label: "GANANDO", color: "var(--lime)" };
  if (s === "outbid") return { label: "SUPERADA", color: "var(--down)" };
  return null;
}
function rowAccent(b: BidWithDetails): string {
  return statusInfo(b)?.color ?? "var(--ink-600)";
}
function bidTop(b: BidWithDetails): number | null {
  return listingById.value.get(b.listing_id)?.highest_bid ?? null;
}
function bidCountdown(b: BidWithDetails): string | null {
  return formatCountdown(msUntil(b.listing.expires_at, now.value));
}
function isRaisable(b: BidWithDetails): boolean {
  const l = listingById.value.get(b.listing_id);
  return l != null && formatCountdown(msUntil(l.expires_at, now.value)) !== null;
}

const leadingCount = computed(() => bids.value.filter((b) => bidStatus(b) === "leading").length);
const outbidCount = computed(() => bids.value.filter((b) => bidStatus(b) === "outbid").length);

const modalCommitted = computed(() => {
  if (modalReplaceBidId.value == null) return committed.value;
  const raised = bids.value.find((b) => b.id === modalReplaceBidId.value);
  return committed.value - (raised?.amount ?? 0);
});

// ── data loading ─────────────────────────────────────────────────────────────

function resolveLeagueId(): number | null {
  const q = route.query.league;
  const raw = Array.isArray(q) ? q[0] : q;
  const parsed = raw != null ? Number(raw) : Number.NaN;
  if (!Number.isNaN(parsed) && leagues.value.some((l) => l.id === parsed)) {
    return parsed;
  }
  return leagues.value[0]?.id ?? null;
}

async function loadMarket(id: number): Promise<void> {
  const [leagueData, members, listingsEnv, bidsEnv, statusEnv] = await Promise.all([
    api.get<LeagueSummary>(`/api/v1/leagues/${id}`),
    api.get<LeagueMember[]>(`/api/v1/leagues/${id}/members`),
    api.get<ApiEnvelope<MarketListingWithDetails[]>>(`/api/v1/leagues/${id}/market/listings`),
    api.get<ApiEnvelope<BidWithDetails[]>>(`/api/v1/leagues/${id}/market/bids`),
    api.get<ApiEnvelope<MarketStatus>>(`/api/v1/leagues/${id}/market/status`),
  ]);
  league.value = leagueData;
  balance.value = members.find((m) => m.user_id === myUserId)?.budget ?? 0;
  listings.value = listingsEnv.data ?? [];
  bids.value = bidsEnv.data ?? [];
  marketStatus.value = statusEnv.data ?? null;

  // jornada is best-effort: leagues without a current matchday return 404.
  try {
    const md = await api.get<Matchday>(`/api/v1/leagues/${id}/matchdays/current`);
    jornada.value = md.number;
  } catch {
    jornada.value = null;
  }
}

async function refresh(): Promise<void> {
  if (leagueId.value == null) return;
  const id = leagueId.value;
  const [listingsEnv, bidsEnv, statusEnv] = await Promise.all([
    api.get<ApiEnvelope<MarketListingWithDetails[]>>(`/api/v1/leagues/${id}/market/listings`),
    api.get<ApiEnvelope<BidWithDetails[]>>(`/api/v1/leagues/${id}/market/bids`),
    api.get<ApiEnvelope<MarketStatus>>(`/api/v1/leagues/${id}/market/status`),
  ]);
  listings.value = listingsEnv.data ?? [];
  bids.value = bidsEnv.data ?? [];
  marketStatus.value = statusEnv.data ?? null;
}

// ── actions ──────────────────────────────────────────────────────────────────

function openBid(listing: MarketListingWithDetails): void {
  if (!marketOpen.value) return; // market is closed — ignore bid attempts
  // If the user already holds an active bid on this listing, open in raise mode
  // (cancel-then-place) instead of stacking a duplicate bid.
  const existing = myBidByListing.value.get(listing.id);
  modalReplaceBidId.value = existing ? existing.id : null;
  modalListing.value = listing;
}

function openRaise(b: BidWithDetails): void {
  const listing = listingById.value.get(b.listing_id);
  if (!listing) return; // button is disabled in this case
  modalReplaceBidId.value = b.id;
  modalListing.value = listing;
}

function closeModal(): void {
  modalListing.value = null;
  modalReplaceBidId.value = null;
}

async function onPlaced(): Promise<void> {
  closeModal();
  try {
    await refresh();
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Error al actualizar el mercado";
  }
}

async function cancelBid(b: BidWithDetails): Promise<void> {
  if (leagueId.value == null) return;
  const ok = window.confirm(`¿Cancelar tu puja de ${millions(b.amount)} por ${fullName(b.player)}?`);
  if (!ok) return;
  try {
    await api.delete(`/api/v1/leagues/${leagueId.value}/market/bids/${b.id}`);
    await refresh();
  } catch (e) {
    error.value = e instanceof Error ? e.message : "No se pudo cancelar la puja";
  }
}

function go(to: string): void {
  if (to !== route.path) router.push(to);
}

function onBottomTab(id: string): void {
  // id is the TabBar's TabId; typed loosely to avoid importing a type from an SFC.
  const map: Record<string, string> = {
    ligas: "/leagues",
    clasif: "/standings",
    equipo: "/team",
    mercado: "/market",
  };
  const target = map[id];
  if (target) go(target);
}

async function onLeagueChange(e: Event): Promise<void> {
  const value = Number((e.target as HTMLSelectElement).value);
  if (Number.isNaN(value) || value === leagueId.value) return;
  leagueId.value = value;
  await router.replace({ query: { ...route.query, league: String(value) } });
  loading.value = true;
  error.value = "";
  activeTab.value = "mercado";
  try {
    await loadMarket(value);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Error al cargar el mercado";
  } finally {
    loading.value = false;
  }
}

function updateIsMobile(): void {
  isMobile.value = window.innerWidth < 768;
}

onMounted(async () => {
  updateIsMobile();
  window.addEventListener("resize", updateIsMobile);
  clockTimer = window.setInterval(() => {
    now.value = Date.now();
  }, 1000);

  try {
    leagues.value = await api.get<LeagueSummary[]>("/api/v1/leagues");
    leagueId.value = resolveLeagueId();
    if (leagueId.value != null) {
      await loadMarket(leagueId.value);
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Error al cargar el mercado";
  } finally {
    loading.value = false;
  }
});

onUnmounted(() => {
  window.removeEventListener("resize", updateIsMobile);
  if (clockTimer !== undefined) window.clearInterval(clockTimer);
});
</script>

<style scoped>
.market {
  min-height: 100vh;
  position: relative;
}

.shell {
  position: relative;
  display: grid;
  grid-template-columns: 240px 1fr;
  height: 100vh; /* desktop: pin the frame, scroll only the list */
}
.shell.mobile {
  display: block;
  height: auto;
  min-height: 100vh;
  padding-bottom: 84px; /* room for the fixed TabBar */
}

.mobile-tabbar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 40;
}

/* sidebar */
.sidebar {
  background: var(--ink-850);
  border-right: 1px solid var(--ink-700);
  padding: 32px 20px;
}
.sidebar-gap {
  height: 32px;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 10px 12px;
  border-radius: 6px;
  background: transparent;
  border: none;
  color: var(--ink-200);
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
  cursor: pointer;
  text-align: left;
}
.nav-item.active {
  background: var(--ink-700);
  color: var(--lime);
}
.nav-dot {
  width: 16px;
  height: 16px;
  border-radius: 2px;
  background: var(--ink-500);
  opacity: 0.5;
}
.nav-dot.active {
  background: var(--lime);
  opacity: 1;
}
.nav-label {
  flex: 1;
}
.nav-badge {
  font-size: 9px;
  padding: 1px 5px;
}

/* main */
.main {
  padding: 32px 48px;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden; /* desktop: contain the scrolling list */
}
.shell.mobile .main {
  padding: 0 20px;
  overflow: visible;
}

.page-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 16px;
  gap: 16px;
}
.page-head.mobile {
  padding-top: 8px;
}
.kicker {
  font-size: 11px;
  color: var(--lime);
  letter-spacing: 0.2em;
  margin-bottom: 8px;
}
.page-head.mobile .kicker {
  font-size: 9px;
  letter-spacing: 0.18em;
  margin-bottom: 2px;
}
.title {
  font-size: 76px;
  margin: 0;
  line-height: 0.9;
}
.page-head.mobile .title {
  font-size: 36px;
}
.league-pick {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
}
.pick-label {
  font-size: 10px;
  color: var(--ink-400);
  letter-spacing: 0.15em;
}
.pick-select {
  background: var(--ink-850);
  border: 1px solid var(--ink-700);
  color: var(--ink-100);
  font-size: 11px;
  padding: 6px 10px;
  border-radius: var(--r-sm);
  cursor: pointer;
}

.head-right {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}
.saldo {
  text-align: right;
}
.saldo-value {
  font-size: 20px;
  color: var(--lime);
  margin-top: 2px;
}
.cierra-card {
  padding: 12px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
  border-color: var(--lime);
}
.cierra-value {
  font-size: 30px;
  color: var(--lime);
  line-height: 1;
}
.pulse {
  width: 10px;
  height: 10px;
  background: var(--pos-fwd);
  border-radius: 50%;
  box-shadow: 0 0 16px var(--pos-fwd);
  flex-shrink: 0;
}

.col-label {
  font-size: 9px;
  color: var(--ink-400);
  letter-spacing: 0.15em;
}

/* in-page tabs */
.ptabs {
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--ink-700);
  margin-bottom: 16px;
}
.ptab {
  padding: 12px 24px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--ink-400);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.15em;
  cursor: pointer;
}
.ptabs.mobile .ptab {
  flex: 1;
  padding: 10px 0;
  font-size: 11px;
}
.ptab.active {
  color: var(--lime);
  border-bottom-color: var(--lime);
}
.tab-tools {
  margin-left: auto;
  padding-bottom: 8px;
}
.search {
  width: 200px;
  font-size: 12px;
  padding: 8px 12px;
}

/* mobile countdown banner */
.cierra-banner {
  padding: 10px 14px;
  background: var(--ink-850);
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}
.banner-label {
  font-size: 10px;
  letter-spacing: 0.15em;
  color: var(--ink-200);
}
.banner-value {
  font-size: 18px;
  color: var(--lime);
  margin-left: auto;
}

/* market closed banner */
.closed-banner {
  padding: 14px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 16px;
  border-color: var(--down);
  background: rgba(255, 98, 98, 0.06);
}
.closed-icon {
  font-size: 20px;
  flex-shrink: 0;
}
.closed-label {
  font-size: 12px;
  letter-spacing: 0.15em;
  color: var(--down);
  font-weight: 600;
}
.closed-sub {
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--ink-300);
  margin-top: 2px;
}

/* filters */
.filters {
  display: flex;
  gap: 6px;
  margin-bottom: 12px;
  overflow-x: auto;
}
.filter-pill {
  padding: 8px 14px;
  border-radius: 4px;
  background: transparent;
  color: var(--ink-200);
  border: 1px solid var(--ink-700);
  font-size: 10px;
  letter-spacing: 0.12em;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
}
.filter-pill.active {
  background: var(--lime);
  color: var(--ink-900);
  border-color: var(--lime);
}

/* content */
.content {
  flex: 1;
  min-height: 0;
  overflow-y: auto; /* desktop: only the list scrolls */
}
.shell.mobile .content {
  overflow-y: visible;
}
.list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-bottom: 24px;
}
.state {
  padding: 48px 16px;
  text-align: center;
  color: var(--ink-300);
  font-size: 12px;
  letter-spacing: 0.1em;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}
.state.error {
  color: var(--down);
}
.join-btn {
  align-self: center;
}

/* mis pujas summary */
.summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  margin-bottom: 22px;
}
.summary.mobile {
  grid-template-columns: repeat(2, 1fr);
}
.summary-card {
  padding: 18px;
}
.summary-value {
  font-size: 36px;
  line-height: 1;
  margin-top: 6px;
  color: var(--ink-100);
}
.summary-value.lime {
  color: var(--lime);
}
.summary-value.down {
  color: var(--down);
}
.summary-sub {
  font-size: 10px;
  color: var(--ink-300);
  margin-top: 4px;
}

/* mis pujas rows */
.bid-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
}
.bid-id {
  flex: 1;
  min-width: 0;
}
.bid-id-top {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
  flex-wrap: wrap;
}
.bid-name {
  font-size: 15px;
  font-weight: 600;
}
.status-chip {
  font-size: 9px;
  letter-spacing: 0.15em;
  padding: 2px 6px;
  border: 1px solid;
  border-radius: 3px;
}
.bid-sub {
  font-size: 10px;
  color: var(--ink-400);
  letter-spacing: 0.1em;
}
.bid-sub-strong {
  color: var(--ink-200);
}
.bid-amount-col {
  border-right: 1px solid var(--ink-700);
  padding-right: 18px;
}
.bid-amount {
  font-size: 24px;
  line-height: 1;
  margin-top: 2px;
}
.bid-top {
  font-size: 24px;
  line-height: 1;
  margin-top: 2px;
}
.row-btn {
  height: 40px;
  flex-shrink: 0;
}
.row-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.cancel-btn {
  color: var(--down);
  border-color: rgba(255, 98, 98, 0.3);
}

@media (max-width: 767px) {
  .bid-amount-col {
    border-right: none;
    padding-right: 0;
  }
  .bid-amount,
  .bid-top {
    font-size: 18px;
  }
  .bid-row {
    gap: 10px;
    flex-wrap: wrap;
  }
}
</style>
