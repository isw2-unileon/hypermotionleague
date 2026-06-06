import { ref } from "vue";

// In-memory singleton: last league the user viewed on StandingsPage.
// Persists across in-app navigation for the lifetime of the browser session.
// Not localStorage — the planning explicitly ruled that out.
export const lastStandingsLeagueId = ref<number | null>(null);
