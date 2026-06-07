<template>
  <div class="bid-overlay" role="dialog" aria-modal="true" @click.self="emit('close')">
    <div class="bid-modal">
      <!-- header -->
      <div class="bm-header">
        <div class="mono kicker">
          ◆ {{ replaceBidId != null ? "SUBIR PUJA" : "NUEVA PUJA" }} · {{ leagueName }}
        </div>
        <div class="bm-header-right">
          <span class="mono activas-label">PUJAS ACTIVAS</span>
          <span class="display activas-count">
            {{ activeBidCount }}<span class="of">/{{ MAX_ACTIVE_BIDS }}</span>
          </span>
          <button class="btn btn-icon btn-ghost close-btn" aria-label="Cerrar" @click="emit('close')">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M3 3 L11 11 M11 3 L3 11" />
            </svg>
          </button>
        </div>
      </div>

      <!-- player hero -->
      <div class="bm-hero">
        <PlayerPhoto :team="short" :color="color" :size="88" />
        <div class="hero-id">
          <div class="hero-top">
            <span class="pos" :class="posMeta.cssClass">{{ posMeta.label }}</span>
            <TeamCrest :short="short" :color="color" :size="20" />
            <span class="mono club">{{ listing.player.team_name }}</span>
          </div>
          <h2 class="display hero-name">{{ name.toUpperCase() }}</h2>
          <!-- FormSpark omitted: no recent-form data on any market response (Phase 0 #4). -->
        </div>
        <div class="hero-close">
          <div class="mono col-label">CIERRA</div>
          <div class="display tnum hero-close-value" :class="{ closed }">
            {{ countdown ?? "CERRADO" }}
          </div>
        </div>
      </div>

      <!-- bid info row -->
      <div class="bm-info">
        <div class="info-cell">
          <div class="mono col-label">BASE</div>
          <div class="display tnum info-value base">{{ millions(min) }}</div>
        </div>
        <div class="info-cell">
          <div class="mono col-label">PUJA TOP ACTUAL</div>
          <div class="display tnum info-value top">{{ topBid != null ? millions(topBid) : "—" }}</div>
          <!-- "por <bidder>" omitted: top bidder name is never exposed (Phase 0 #3). -->
        </div>
        <div class="info-cell">
          <div class="mono col-label">TU SALDO</div>
          <div class="display tnum info-value">{{ millions(available) }}</div>
          <div class="mono info-sub">{{ millions(balance) }} total</div>
        </div>
      </div>

      <!-- bid input -->
      <div class="bm-input">
        <div class="mono col-label tu-puja">TU PUJA</div>
        <div class="input-row">
          <button
            class="btn btn-secondary btn-icon stepper"
            aria-label="Reducir"
            :disabled="!canBid || amount <= min"
            @click="step(-STEP)"
          >
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 7 H11" /></svg>
          </button>
          <div class="amount-box" :class="{ disabled: !canBid }">
            <span class="mono euro">€</span>
            <span class="display tnum amount-num">{{ amount.toLocaleString("es-ES") }}</span>
          </div>
          <button
            class="btn btn-secondary btn-icon stepper"
            aria-label="Aumentar"
            :disabled="!canBid || amount >= maxBid"
            @click="step(STEP)"
          >
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 7 H11 M7 3 V11" /></svg>
          </button>
        </div>

        <input
          v-if="canBid && maxBid > min"
          class="slider"
          type="range"
          :min="min"
          :max="maxBid"
          :step="STEP"
          :value="amount"
          @input="onSlider"
        />
        <div class="slider-bounds">
          <span class="mono">MÍN {{ millions(min) }}</span>
          <span class="mono">MÁX {{ millions(maxBid) }}</span>
        </div>

        <div v-if="quick.length" class="quick-row">
          <button
            v-for="v in quick"
            :key="v"
            class="quick-btn mono"
            :class="{ active: v === amount }"
            @click="setAmount(v)"
          >
            {{ millions(v) }}
          </button>
        </div>
      </div>

      <!-- blocking states / errors -->
      <div v-if="blockingMessage" class="bm-note warn">{{ blockingMessage }}</div>
      <div v-else-if="replaceBidId != null" class="bm-note">
        Al confirmar se cancela tu puja actual y se registra la nueva (no existe edición directa en el backend).
      </div>
      <div v-if="errorMsg" class="bm-note error">{{ errorMsg }}</div>

      <!-- CTAs -->
      <div class="bm-ctas">
        <button class="btn btn-ghost cancel" :disabled="submitting" @click="emit('close')">Cancelar</button>
        <button class="btn btn-primary confirm" :disabled="!canSubmit" @click="submit">
          {{ submitting ? "Enviando…" : `Confirmar puja · ${millions(amount)}` }}
        </button>
      </div>

      <!--
        Blind-auction footer copy from the design is intentionally NOT shown:
        the backend exposes the current top bid (highest_bid), so the bid is not
        hidden until close (Phase 0 #3). We state what is actually true instead.
      -->
      <div class="mono bm-footer">◆ GANA LA PUJA MÁS ALTA AL CIERRE · TU SALDO INCLUYE LO YA COMPROMETIDO</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import PlayerPhoto from "@/design-system/primitives/PlayerPhoto.vue";
import TeamCrest from "@/design-system/primitives/TeamCrest.vue";
import api from "@/lib/api";
import {
  type MarketListingWithDetails,
  MAX_ACTIVE_BIDS,
  bidErrorMessage,
  formatCountdown,
  fullName,
  millions,
  msUntil,
  positionMeta,
  quickAmounts,
  teamColor,
  teamShort,
} from "@/lib/market";

const STEP = 100_000; // 0.1M increments

interface Props {
  listing: MarketListingWithDetails;
  leagueId: number;
  leagueName: string;
  balance: number; // raw euros
  committed: number; // raw euros already committed in OTHER active bids (excludes the one being raised)
  activeBidCount: number;
  now: number;
  replaceBidId?: number | null; // set => raise mode: cancel this bid, then place the new one
}

const props = withDefaults(defineProps<Props>(), {
  replaceBidId: null,
});

const emit = defineEmits<{
  (e: "placed"): void;
  (e: "close"): void;
}>();

const submitting = ref(false);
const errorMsg = ref("");

const posMeta = computed(() => positionMeta(props.listing.player.position));
const name = computed(() => fullName(props.listing.player));
const short = computed(() => teamShort(props.listing.player.team_name));
const color = computed(() => teamColor(props.listing.player.team_name));
const topBid = computed(() => props.listing.highest_bid ?? null);

const min = computed(() => props.listing.base_price);
// Available = total balance minus funds already committed in other active bids.
const available = computed(() => props.balance - props.committed);
// Mirror the backend budget rule (PlaceBidTx: committed + amount <= budget).
const maxBid = computed(() => Math.max(min.value, props.balance - props.committed));
const affordable = computed(() => props.balance - props.committed >= min.value);

const countdown = computed(() =>
  formatCountdown(msUntil(props.listing.expires_at, props.now)),
);
const closed = computed(() => countdown.value === null);
const atCap = computed(
  () => props.replaceBidId == null && props.activeBidCount >= MAX_ACTIVE_BIDS,
);

const canBid = computed(() => !closed.value && !atCap.value && affordable.value);

const quick = computed(() =>
  canBid.value ? quickAmounts(min.value, topBid.value, maxBid.value) : [],
);

const amount = ref(min.value);

const canSubmit = computed(
  () =>
    canBid.value &&
    !submitting.value &&
    amount.value >= min.value &&
    amount.value <= maxBid.value,
);

const blockingMessage = computed(() => {
  if (closed.value) return "La subasta de este jugador ya ha cerrado.";
  if (atCap.value)
    return `Ya tienes ${MAX_ACTIVE_BIDS} pujas activas. Cancela una para pujar por otro jugador.`;
  if (!affordable.value)
    return "Tu saldo disponible no cubre el precio base de este jugador.";
  return "";
});

function clamp(v: number): number {
  if (Number.isNaN(v)) return min.value;
  const stepped = Math.round(v / STEP) * STEP;
  return Math.min(maxBid.value, Math.max(min.value, stepped));
}

function setAmount(v: number): void {
  amount.value = clamp(v);
}

function step(delta: number): void {
  amount.value = clamp(amount.value + delta);
}

function onSlider(e: Event): void {
  const target = e.target as HTMLInputElement;
  amount.value = clamp(Number(target.value));
}

async function submit(): Promise<void> {
  if (!canSubmit.value) return;
  submitting.value = true;
  errorMsg.value = "";
  try {
    // Raise mode: there is no PUT/raise endpoint and `bids` has no unique
    // (listing,user) constraint, so a second bid would double-commit budget and
    // burn another of the 5 slots. We therefore cancel the existing bid first,
    // then place the higher one. Non-atomic: if the POST fails the old bid is
    // already gone — the parent refetches so the user can re-bid.
    if (props.replaceBidId != null) {
      await api.delete(`/api/v1/leagues/${props.leagueId}/market/bids/${props.replaceBidId}`);
    }
    await api.post(`/api/v1/leagues/${props.leagueId}/market/bids`, {
      listing_id: props.listing.id,
      amount: amount.value, // raw euros — the unit the backend expects (Phase 0 #1)
    });
    emit("placed");
  } catch (e) {
    errorMsg.value = bidErrorMessage(e instanceof Error ? e.message : "Error al pujar");
  } finally {
    submitting.value = false;
  }
}

function onKeydown(e: KeyboardEvent): void {
  if (e.key === "Escape") emit("close");
}

onMounted(() => {
  // Default to the first sensible "beat the top" suggestion, else the base price.
  amount.value = clamp(quick.value[0] ?? min.value);
  window.addEventListener("keydown", onKeydown);
});

onUnmounted(() => {
  window.removeEventListener("keydown", onKeydown);
});
</script>

<style scoped>
.bid-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  background: rgba(8, 9, 11, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  backdrop-filter: blur(2px);
}

.bid-modal {
  width: 760px;
  max-width: 96vw;
  max-height: 94vh;
  overflow-y: auto;
  background: var(--ink-850);
  border: 1px solid var(--ink-600);
  border-radius: 16px;
  box-shadow: 0 40px 100px rgba(0, 0, 0, 0.6), 0 0 0 1px rgba(199, 255, 61, 0.15);
}

.bm-header {
  padding: 18px 24px;
  border-bottom: 1px solid var(--ink-700);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}
.kicker {
  font-size: 11px;
  color: var(--lime);
  letter-spacing: 0.2em;
}
.bm-header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}
.activas-label {
  font-size: 10px;
  color: var(--ink-400);
  letter-spacing: 0.15em;
}
.activas-count {
  font-size: 18px;
  color: var(--lime);
}
.activas-count .of {
  color: var(--ink-400);
}
.close-btn {
  margin-left: 4px;
}

.bm-hero {
  padding: 22px 24px;
  display: flex;
  align-items: center;
  gap: 18px;
  background: linear-gradient(180deg, rgba(199, 255, 61, 0.06), transparent 70%);
}
.hero-id {
  flex: 1;
  min-width: 0;
}
.hero-top {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.club {
  font-size: 10px;
  color: var(--ink-300);
  letter-spacing: 0.1em;
  text-transform: uppercase;
}
.hero-name {
  font-size: 40px;
  margin: 0;
  line-height: 1;
  letter-spacing: 0.02em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.hero-close {
  text-align: right;
}
.hero-close-value {
  font-size: 28px;
  color: var(--lime);
  line-height: 1;
  margin-top: 4px;
}
.hero-close-value.closed {
  color: var(--down);
  font-size: 20px;
}

.col-label {
  font-size: 10px;
  color: var(--ink-400);
  letter-spacing: 0.15em;
}

.bm-info {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  border-top: 1px solid var(--ink-700);
  border-bottom: 1px solid var(--ink-700);
}
.info-cell {
  padding: 14px 18px;
}
.info-cell + .info-cell {
  border-left: 1px solid var(--ink-700);
}
.info-value {
  font-size: 26px;
  line-height: 1;
  margin-top: 4px;
  color: var(--ink-200);
}
.info-value.top {
  color: var(--lime);
}
.info-sub {
  font-size: 10px;
  color: var(--ink-400);
  margin-top: 2px;
}

.bm-input {
  padding: 22px 24px 6px;
}
.tu-puja {
  margin-bottom: 12px;
}
.input-row {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 16px;
}
.stepper {
  width: 44px;
  height: 44px;
  flex-shrink: 0;
}
.amount-box {
  flex: 1;
  background: var(--ink-800);
  border: 1px solid var(--lime);
  border-radius: 10px;
  padding: 12px 18px;
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 6px;
  box-shadow: 0 0 0 4px rgba(199, 255, 61, 0.08);
}
.amount-box.disabled {
  border-color: var(--ink-600);
  box-shadow: none;
  opacity: 0.6;
}
.euro {
  font-size: 18px;
  color: var(--ink-300);
}
.amount-num {
  font-size: 52px;
  color: var(--lime);
  line-height: 1;
}

.slider {
  width: 100%;
  appearance: none;
  -webkit-appearance: none;
  height: 4px;
  border-radius: 2px;
  background: var(--ink-700);
  outline: none;
  margin: 6px 0;
}
.slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--lime);
  border: 3px solid var(--ink-900);
  box-shadow: 0 0 0 1px var(--lime), 0 0 14px var(--lime-glow);
  cursor: pointer;
}
.slider::-moz-range-thumb {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--lime);
  border: 3px solid var(--ink-900);
  cursor: pointer;
}
.slider-bounds {
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  color: var(--ink-400);
  letter-spacing: 0.1em;
}

.quick-row {
  display: flex;
  gap: 6px;
  margin-top: 14px;
}
.quick-btn {
  flex: 1;
  padding: 8px 0;
  background: transparent;
  border: 1px solid var(--ink-700);
  border-radius: 4px;
  color: var(--ink-200);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
}
.quick-btn.active {
  border-color: var(--lime);
  color: var(--lime);
}

.bm-note {
  margin: 4px 24px 0;
  padding: 10px 14px;
  border-radius: 6px;
  font-size: 12px;
  background: var(--ink-800);
  border: 1px solid var(--ink-700);
  color: var(--ink-300);
}
.bm-note.warn {
  border-color: rgba(255, 178, 59, 0.4);
  color: var(--warn);
}
.bm-note.error {
  border-color: rgba(255, 98, 98, 0.4);
  color: var(--down);
}

.bm-ctas {
  padding: 18px 24px 8px;
  display: flex;
  gap: 10px;
}
.cancel {
  flex: 1;
  height: 52px;
}
.confirm {
  flex: 2;
  height: 52px;
  font-size: 15px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
.confirm:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.bm-footer {
  font-size: 9px;
  color: var(--ink-500);
  letter-spacing: 0.12em;
  text-align: center;
  padding: 0 24px 18px;
}

@media (max-width: 600px) {
  .hero-name {
    font-size: 30px;
  }
  .amount-num {
    font-size: 40px;
  }
  .info-value {
    font-size: 20px;
  }
}
</style>
