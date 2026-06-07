import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { BASE_URL } from "@/lib/api";
import type { MarketStatus, ApiEnvelope } from "@/lib/market";
import { getToken } from "@/lib/tokenStore";

export function useMarketCountdown(leagueId: () => number | null) {
  const status = ref<MarketStatus | null>(null);
  const now = ref<number>(Date.now());

  let tickInterval: number | undefined;
  let refetchInterval: number | undefined;

  async function fetchStatus(): Promise<void> {
    const id = leagueId();
    if (id == null) return;
    try {
      const token = getToken();
      const res = await fetch(`${BASE_URL}/api/v1/leagues/${id}/market/status`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) return;
      const body: ApiEnvelope<MarketStatus> = await res.json();
      status.value = body.data;
    } catch (e){
        void e;
    }
  }

  const label = computed<string>(() => {
    if (!status.value) return "PRÓXIMA APERTURA";
    if (status.value.is_open) return "PRÓXIMO CIERRE";
    return "PRÓXIMA APERTURA";
  });

  const subtitle = computed<string>(() => {
    if (!status.value) return "Mercado abre a las 19:00";
    switch (status.value.reason) {
      case "open":
        return "Mercado abierto · cierra a las 00:00";
      case "outside_window":
        return "Mercado abre a las 19:00";
      case "active_matchday":
        return "Jornada en curso";
      default:
        return "Mercado abre a las 19:00";
    }
  });

  const timeText = computed<string>(() => {
    if (!status.value) return "--:--:--";
    const target = new Date(status.value.next_change_at).getTime();
    const diff = Math.max(0, target - now.value);
    const totalSeconds = Math.floor(diff / 1000);
    const h = Math.floor(totalSeconds / 3600);
    const m = Math.floor((totalSeconds % 3600) / 60);
    const s = totalSeconds % 60;
    return `${pad(h)}:${pad(m)}:${pad(s)}`;
  });

  function pad(n: number): string {
    return n.toString().padStart(2, "0");
  }

  // Re-fetch immediately whenever the active league ID changes (e.g. user
  // navigates to a page that writes a different league to activeLeagueId).
  const currentId = computed(leagueId);
  watch(currentId, (newId) => {
    if (newId != null) fetchStatus();
  });

  onMounted(() => {
    fetchStatus();
    tickInterval = window.setInterval(() => {
      now.value = Date.now();
    }, 1000);
    refetchInterval = window.setInterval(fetchStatus, 60000);
  });

  onUnmounted(() => {
    if (tickInterval !== undefined) window.clearInterval(tickInterval);
    if (refetchInterval !== undefined) window.clearInterval(refetchInterval);
  });

  return { timeText, label, subtitle, status };
}