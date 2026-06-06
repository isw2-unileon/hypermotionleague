import { ref, computed, onMounted, onUnmounted } from "vue";
import { BASE_URL } from "@/lib/api";

interface MarketStatus {
  league_id: number;
  is_open: boolean;
  next_change_at: string;
  reason: "open" | "outside_window" | "active_matchday";
  active_listings: number;
  your_active_bids: number;
  max_bids_per_user: number;
}

interface ApiEnvelope<T> {
  status: string;
  data: T;
}

export function useMarketCountdown(leagueId: () => number | null) {
  const status = ref<MarketStatus | null>(null);
  const now = ref<number>(Date.now());

  let tickInterval: number | undefined;
  let refetchInterval: number | undefined;

  async function fetchStatus(): Promise<void> {
    const id = leagueId();
    if (id == null) return;
    try {
      const token = localStorage.getItem("token");
      const res = await fetch(`${BASE_URL}/api/v1/leagues/${id}/market/status`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) return;
      const body: ApiEnvelope<MarketStatus> = await res.json();
      status.value = body.data;
    } catch {
    }
  }

  const label = computed<string>(() => {
    if (!status.value) return "PRÓXIMO CIERRE";
    if (status.value.is_open) return "PRÓXIMO CIERRE";
    return "PRÓXIMA APERTURA";
  });

  const subtitle = computed<string>(() => {
    if (!status.value) return "Mercado cierra hoy";
    switch (status.value.reason) {
      case "open":
        return "Mercado abierto";
      case "outside_window":
        return "Fuera de horario";
      case "active_matchday":
        return "Jornada en curso";
      default:
        return "Mercado cierra hoy";
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

  onMounted(() => {
    fetchStatus();
    tickInterval = window.setInterval(() => {
      now.value = Date.now();
    }, 1000);
    //60 seg change the status(open→closed)
    refetchInterval = window.setInterval(fetchStatus, 60000);
  });

  onUnmounted(() => {
    if (tickInterval !== undefined) window.clearInterval(tickInterval);
    if (refetchInterval !== undefined) window.clearInterval(refetchInterval);
  });

  return { timeText, label, subtitle, status };
}