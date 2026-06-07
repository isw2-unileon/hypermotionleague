import { ref } from "vue";

// Shared reactive ref: any page that resolves a league writes here.
// AppShell reads it as fallback when no league ID appears in the route.
export const activeLeagueId = ref<number | null>(null);
