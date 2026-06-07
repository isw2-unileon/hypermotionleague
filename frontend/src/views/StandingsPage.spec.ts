import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ref } from "vue";

// --- Mocks ---
// Cliente API: cada test programa qué devuelve api.get y verifica qué 
// URLs se llaman.
const apiGetMock = vi.fn();
vi.mock("@/lib/api", () => ({
  default: {
    get: (path: string) => apiGetMock(path),
  },
}));

// El api importa @/router, que importa Supabase y peta con WebSocket.
// Stubbeamos ambos.
vi.mock("@/router", () => ({
  default: { push: vi.fn(), replace: vi.fn() },
}));
vi.mock("@/lib/supabase", () => ({
  supabase: {},
}));

// vue-router: route.query controlable, router.push/replace son spies.
const routePush = vi.fn();
const routeReplace = vi.fn();
const routeQuery = ref<Record<string, string>>({});
vi.mock("vue-router", () => ({
  useRoute: () => ({
    query: routeQuery.value,
    path: "/standings",
    params: {},
  }),
  useRouter: () => ({ push: routePush, replace: routeReplace }),
}));

// AppShell hace fetch al market status, no aporta nada al test → stub.
vi.mock("@/design-system/AppShell.vue", () => ({
  default: { template: "<div><slot /></div>" },
}));

// Podium y StandingsRow no son lo que probamos aquí; los stubeamos como
// componentes simples que emiten 'click' con el userId.
vi.mock("@/design-system/components/Podium.vue", () => ({
  default: {
    props: ["top3", "mobile"],
    template: "<div class='podium-stub' @click=\"$emit('click', top3?.[0]?.userId)\"></div>",
    emits: ["click"],
  },
}));
vi.mock("@/design-system/components/StandingsRow.vue", () => ({
  default: {
    props: ["row", "mobile"],
    template: "<div class='row-stub' @click=\"$emit('click', row.userId)\"></div>",
    emits: ["click"],
  },
}));

// auth: devolvemos un user id fijo.
vi.mock("@/lib/auth", () => ({
  currentUserId: () => 999,
}));

// Memoria de última liga visitada: ref simple para que el componente
// pueda leer/escribir sin tocar localStorage.
vi.mock("@/lib/standings-memory", () => ({
  lastStandingsLeagueId: ref<number | null>(null),
}));

// Importamos el componente DESPUÉS de los mocks.
import StandingsPage from "./StandingsPage.vue";

// --- Datos de prueba ---
const mockLeagues = [{ id: 1, name: "Liga test" }];
const mockMatchdays = {
  matchdays: [
    { id: 10, number: 1, name: "Jornada 1" },
    { id: 11, number: 2, name: "Jornada 2" },
  ],
};
const mockStandings = {
  league_id: 1,
  rankings: [
    { rank: 1, user_id: 100, username: "ana", display_name: "Ana", total_points: 50 },
    { rank: 2, user_id: 200, username: "luis", display_name: "Luis", total_points: 40 },
    { rank: 3, user_id: 300, username: "marc", display_name: "Marc", total_points: 30 },
    { rank: 4, user_id: 400, username: "eva", display_name: "Eva", total_points: 20 },
  ],
};

beforeEach(() => {
  apiGetMock.mockReset();
  routePush.mockReset();
  routeReplace.mockReset();
  routeQuery.value = {};
});

describe("StandingsPage", () => {
  // ---------------------------------------------------------------
  // TEST 1: cambiar de jornada (pill) dispara nueva petición a la API
  // con la URL específica de esa jornada. Esto verifica que la lógica
  // del filtro de jornada está cableada al fetchStandings.
  // ---------------------------------------------------------------
  it("al seleccionar una jornada llama a la API con la URL correcta", async () => {
    routeQuery.value = { leagueId: "1" };

    // onMounted hace 3 llamadas: /leagues, /matchdays, /standings.
    apiGetMock
      .mockResolvedValueOnce(mockLeagues)
      .mockResolvedValueOnce(mockMatchdays)
      .mockResolvedValueOnce(mockStandings)
      .mockResolvedValueOnce(mockStandings); // la que dispara selectMatchday

    const wrapper = mount(StandingsPage);
    await flushPromises();

    // Buscamos los pills de matchday. Tienen clase .matchday-pill o similar
    // según el template. Si tu template usa <button> dentro de un contenedor
    // de pills, los buscamos por texto "J·2".
    const buttons = wrapper.findAll("button");
    const j2 = buttons.find((b) => b.text().includes("2"));
    if (!j2) throw new Error("No se encontró el pill de Jornada 2");
    await j2.trigger("click");
    await flushPromises();

    // Aserción clave: la última llamada incluye el path con jornada 2.
    expect(apiGetMock).toHaveBeenLastCalledWith(
      "/api/v1/leagues/1/matchdays/2/standings",
    );
  });

  // ---------------------------------------------------------------
  // TEST 2: click en una fila navega a /squad/:leagueId/:userId.
  // Como StandingsRow está stubbeado, emite 'click' con el userId del row.
  // ---------------------------------------------------------------
  it("al hacer click en una fila navega con router.push a /squad", async () => {
    routeQuery.value = { leagueId: "1" };
    apiGetMock
      .mockResolvedValueOnce(mockLeagues)
      .mockResolvedValueOnce(mockMatchdays)
      .mockResolvedValueOnce(mockStandings);

    const wrapper = mount(StandingsPage);
    await flushPromises();

    // Las filas que NO están en el podio (slice(3)) se renderizan como 
    // StandingsRow. Cada stub emite click con su userId al pulsar.
    const rows = wrapper.findAll(".row-stub");
    expect(rows.length).toBeGreaterThanOrEqual(1);
    const firstRestRow = rows[0];
    if (!firstRestRow) throw new Error("No se encontró ninguna StandingsRow");

    await firstRestRow.trigger("click");

    // Aserción: navegó a /squad/1/<userId>. El primer row del 'rest' es
    // el de rank 4 (Eva, user_id 400).
    expect(routePush).toHaveBeenCalledWith("/squad/1/400");
  });

  // ---------------------------------------------------------------
  // TEST 3: con varias ligas y sin query param, debe cargar la primera
  // automáticamente. Esto cubre la lógica de "auto-load" del Sprint 3.
  // ---------------------------------------------------------------
  it("auto-selecciona la primera liga si no hay query ni memoria", async () => {
    routeQuery.value = {};
    apiGetMock
      .mockResolvedValueOnce(mockLeagues)        // /leagues
      .mockResolvedValueOnce(mockMatchdays)      // /matchdays de la 1
      .mockResolvedValueOnce(mockStandings);     // /standings de la 1

    mount(StandingsPage);
    await flushPromises();

    // Sin query y sin memoria, debe haber cargado las standings de la
    // primera liga (id=1).
    expect(apiGetMock).toHaveBeenCalledWith("/api/v1/leagues/1/standings");
  });
});