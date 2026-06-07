// Types + helpers for the league market UI.
//
// The interfaces mirror the backend DTOs verbatim:
//   backend/internal/models/{market,player,league,matchday}.go
// Every market endpoint wraps its payload in { status, data }, and `data` is
// `null` (not []) when there is nothing to return (Go nil slices -> JSON null).
//
// MONEY UNIT: the backend stores and expects RAW EUROS (seed data: base_price
// 10000000 = €10M, budgets 100000000 = €100M). The "€X.XM" the design shows is
// purely a display transform; everything sent to the API is in raw euros.

export type PlayerPosition = "GK" | "DEF" | "MID" | "FWD";

export interface Player {
  id: number;
  first_name: string;
  last_name: string;
  position: PlayerPosition;
  team_name: string;
  market_value: number; // raw euros
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface MarketListing {
  id: number;
  league_id: number;
  player_id: number;
  base_price: number; // raw euros
  seller_id?: number | null; // null/absent when the listing comes from the global pool
  status: string;
  listed_at: string;
  expires_at: string; // ISO timestamp — the real, per-listing countdown source
}

export interface MarketListingWithDetails extends MarketListing {
  player: Player;
  // seller_name is the SELLER's display name (null for global-pool listings).
  // It is NOT the top bidder — the backend never exposes who holds the top bid.
  seller_name?: string | null;
  highest_bid?: number | null; // raw euros; absent/null when there are no active bids
  bid_count: number;
}

export type BidStatus = "active" | "won" | "lost" | "cancelled";

export interface Bid {
  id: number;
  listing_id: number;
  user_id: number;
  amount: number; // raw euros
  status: BidStatus;
  placed_at: string;
}

export interface BidWithDetails extends Bid {
  // NOTE: this nested listing is a plain MarketListing — it does NOT carry
  // highest_bid / bid_count. We cross-reference the listings feed for the top bid.
  listing: MarketListing;
  player: Player;
}

// Why the market window is in its current state. Mirrors models.MarketStatus
// Reason values produced by the market window rule (market/window.go).
export type MarketReason = "open" | "outside_window" | "active_matchday";

// Mirrors backend models.MarketStatus (see market_handler.go GetMarketStatus).
// `next_change_at` is the ISO timestamp when the trading window next flips: the
// upcoming CLOSE time while the market is open, the next OPEN time while closed.
// (There is no `closes_at` field — the backend only ever sends next_change_at.)
export interface MarketStatus {
  league_id: number;
  is_open: boolean;
  next_change_at: string;
  reason: MarketReason;
  active_listings: number;
  your_active_bids: number;
  max_bids_per_user: number;
}

// Every market endpoint returns this envelope.
export interface ApiEnvelope<T> {
  status: string;
  data: T | null;
}

export interface Matchday {
  id: number;
  league_id: number;
  number: number;
  name: string;
  start_date: string;
  end_date: string;
  is_current: boolean;
  created_at: string;
}

export interface LeagueSummary {
  id: number;
  name: string;
  invite_code: string;
  max_members: number;
  budget_per_user: number;
  market_close_time: string;
  created_by: number;
}

export interface LeagueMember {
  id: number;
  league_id: number;
  user_id: number;
  role: string;
  budget: number; // raw euros — the user's remaining budget in this league
  joined_at: string;
  username: string;
  display_name: string;
  avatar_url?: string | null;
}

// The backend caps active bids per user per league at 5: PlaceBidTx rejects a
// 6th with "MAX_BIDS_REACHED", and GetMarketStatus reports max_bids_per_user=5.
export const MAX_ACTIVE_BIDS = 5;

export interface PositionMeta {
  label: string; // Spanish abbreviation in the badge (design: POR/DEF/MED/DEL)
  cssClass: string; // position token class from tokens.css
}

// Maps the backend position enum (GK/DEF/MID/FWD) to the design's Spanish badge.
export function positionMeta(pos: PlayerPosition): PositionMeta {
  switch (pos) {
    case "GK":
      return { label: "POR", cssClass: "pos-gk" };
    case "DEF":
      return { label: "DEF", cssClass: "pos-def" };
    case "MID":
      return { label: "MED", cssClass: "pos-mid" };
    case "FWD":
      return { label: "DEL", cssClass: "pos-fwd" };
  }
}

export function fullName(p: Player): string {
  return `${p.first_name} ${p.last_name}`.trim();
}

// Raw euros -> "€18.500.000" (es-ES grouping). For exact amounts.
const eurosFmt = new Intl.NumberFormat("es-ES", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 0,
});
export function euros(amount: number): string {
  return eurosFmt.format(amount);
}

// Raw euros -> "€18.5M" (one decimal, trailing .0 trimmed). Display only.
export function millions(amount: number): string {
  const text = (amount / 1_000_000).toFixed(1).replace(/\.0$/, "");
  return `€${text}M`;
}

// ── Decorative team identity ────────────────────────────────────────────────
// The backend exposes only `team_name` (e.g. "Real Madrid"); there is NO team
// short-code or color field anywhere in the API. Both helpers DERIVE a stable
// value from the name for the crest/photo visuals only — never sent back.

export function teamShort(teamName: string): string {
  const letters = teamName.replace(/[^A-Za-z]/g, "").toUpperCase();
  return letters ? letters.slice(0, 3) : "—";
}

export function teamColor(teamName: string): string {
  let hue = 0;
  for (let i = 0; i < teamName.length; i++) {
    hue = (hue * 31 + teamName.charCodeAt(i)) % 360;
  }
  return `hsl(${hue}, 60%, 45%)`;
}

// ── Countdown ────────────────────────────────────────────────────────────────

export function msUntil(iso: string, now: number): number {
  const t = new Date(iso).getTime();
  if (Number.isNaN(t)) return 0;
  return t - now;
}

// ms remaining -> "HH:MM:SS" (hours may exceed 24). Returns null once elapsed,
// which callers render as "CERRADO".
export function formatCountdown(ms: number): string | null {
  if (ms <= 0) return null;
  const totalSec = Math.floor(ms / 1000);
  const h = Math.floor(totalSec / 3600);
  const m = Math.floor((totalSec % 3600) / 60);
  const s = totalSec % 60;
  const pad = (n: number): string => n.toString().padStart(2, "0");
  return `${pad(h)}:${pad(m)}:${pad(s)}`;
}

// ── Bid suggestions + budget ─────────────────────────────────────────────────

// Quick-amount buttons for the modal: a few sensible round amounts that beat the
// current top bid, in raw euros, capped at the affordable maximum. Derived UX
// suggestions — the backend does not mandate any minimum beyond amount > 0.
export function quickAmounts(
  base: number,
  top: number | null | undefined,
  max: number,
): number[] {
  const start = Math.max(base, (top ?? 0) + 100_000);
  const rounded = Math.ceil(start / 100_000) * 100_000;
  const out: number[] = [];
  for (const inc of [0, 500_000, 1_000_000, 2_000_000, 3_500_000]) {
    const v = rounded + inc;
    if (v <= max && !out.includes(v)) out.push(v);
  }
  return out;
}

// Maps the exact error strings returned by market_handler.go to clear Spanish
// copy, falling back to the backend's own message for anything unrecognised.
export function bidErrorMessage(raw: string): string {
  const m = raw.toLowerCase();
  if (m.includes("maximum limit") || m.includes("max_bids"))
    return "Has alcanzado el máximo de 5 pujas activas.";
  if (m.includes("insufficient"))
    return "Saldo insuficiente: recuerda que incluye el dinero ya comprometido en otras pujas.";
  if (m.includes("expired")) return "La subasta de este jugador ya ha cerrado.";
  if (m.includes("listing not found"))
    return "Este jugador ya no está disponible en el mercado.";
  if (m.includes("idor") || m.includes("miembro"))
    return "No tienes acceso al mercado de esta liga.";
  if (m.includes("invalid request") || m.includes("invalid bid"))
    return "Revisa el importe de la puja e inténtalo de nuevo.";
  return raw;
}
