<template>
  <div class="player-card card" :class="{ mobile, closed }">
    <PlayerPhoto :team="short" :color="color" :size="mobile ? 48 : 64" />

    <div class="identity">
      <div class="identity-top">
        <span class="pos" :class="posMeta.cssClass">{{ posMeta.label }}</span>
        <TeamCrest :short="short" :color="color" :size="18" />
      </div>
      <div class="name">{{ name }}</div>
      <!--
        FormSpark omitted here: no market endpoint exposes per-player recent form
        (Phase 0 #4). We show the real club + market value instead of faking a chart.
      -->
      <div class="subline mono">
        {{ listing.player.team_name }} · {{ millions(listing.player.market_value) }} valor
      </div>
    </div>

    <!-- BASE: desktop only, matching the design -->
    <div v-if="!mobile" class="col base-col">
      <div class="col-label mono">BASE</div>
      <div class="col-value display tnum base-value">{{ millions(listing.base_price) }}</div>
    </div>

    <div class="col">
      <div class="col-label mono">PUJA TOP</div>
      <div class="col-value display tnum top-value" :class="{ none: topBid == null }">
        {{ topBid != null ? millions(topBid) : "—" }}
      </div>
      <!-- Top-bidder name intentionally omitted: the backend never exposes it (Phase 0 #3). -->
      <div v-if="listing.bid_count > 0" class="col-sub mono">
        {{ listing.bid_count }} {{ listing.bid_count === 1 ? "puja" : "pujas" }}
      </div>
    </div>

    <div class="col close-col">
      <div class="col-label mono">CIERRA</div>
      <div class="col-value display tnum close-value" :class="{ closed }">
        {{ countdown ?? "CERRADO" }}
      </div>
    </div>

    <button
      class="btn pujar"
      :class="mine ? 'btn-secondary' : 'btn-primary'"
      :disabled="closed"
      @click="emit('bid', listing)"
    >
      {{ mine ? "Subir" : "Pujar" }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import PlayerPhoto from "@/design-system/primitives/PlayerPhoto.vue";
import TeamCrest from "@/design-system/primitives/TeamCrest.vue";
import {
  type MarketListingWithDetails,
  fullName,
  millions,
  msUntil,
  formatCountdown,
  positionMeta,
  teamColor,
  teamShort,
} from "@/lib/market";

interface Props {
  listing: MarketListingWithDetails;
  now: number;
  mobile?: boolean;
  mine?: boolean; // the signed-in user already holds an active bid on this listing
}

const props = withDefaults(defineProps<Props>(), {
  mobile: false,
  mine: false,
});

const emit = defineEmits<{ (e: "bid", listing: MarketListingWithDetails): void }>();

const posMeta = computed(() => positionMeta(props.listing.player.position));
const name = computed(() => fullName(props.listing.player));
const short = computed(() => teamShort(props.listing.player.team_name));
const color = computed(() => teamColor(props.listing.player.team_name));
const topBid = computed(() => props.listing.highest_bid ?? null);
const countdown = computed(() =>
  formatCountdown(msUntil(props.listing.expires_at, props.now)),
);
const closed = computed(() => countdown.value === null);
</script>

<style scoped>
.player-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
}
.player-card.mobile {
  gap: 12px;
  padding: 12px;
}
.player-card.closed {
  opacity: 0.6;
}

.identity {
  flex: 1;
  min-width: 0;
}
.identity-top {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}
.name {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.mobile .name {
  font-size: 14px;
}
.subline {
  font-size: 9px;
  color: var(--ink-400);
  letter-spacing: 0.08em;
  text-transform: uppercase;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.col {
  text-align: right;
}
.base-col {
  border-right: 1px solid var(--ink-700);
  padding-right: 16px;
}
.close-col {
  min-width: 90px;
}
.mobile .close-col {
  min-width: 60px;
}
.col-label {
  font-size: 9px;
  color: var(--ink-400);
  letter-spacing: 0.15em;
}
.col-sub {
  font-size: 9px;
  color: var(--ink-300);
  margin-top: 4px;
}
.col-value {
  line-height: 1;
  margin-top: 2px;
}
.base-value {
  font-size: 22px;
  color: var(--ink-200);
}
.top-value {
  font-size: 28px;
  color: var(--lime);
}
.mobile .top-value {
  font-size: 22px;
}
.top-value.none {
  color: var(--ink-400);
}
.close-value {
  font-size: 16px;
  color: var(--ink-100);
}
.mobile .close-value {
  font-size: 14px;
}
.close-value.closed {
  color: var(--down);
  font-size: 13px;
}

.pujar {
  height: 44px;
  padding: 0 22px;
  font-size: 13px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  flex-shrink: 0;
}
.mobile .pujar {
  height: 36px;
  padding: 0 14px;
  font-size: 12px;
}
.pujar:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
</style>
