import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ref } from "vue";

const routePush = vi.fn();
const routeReplace = vi.fn();
const routePath = ref<string>("/leagues");
vi.mock("vue-router", () => ({
  useRoute: () => ({
    get path() { return routePath.value; },
    query: {},
    params: {},
  }),
  useRouter: () => ({ push: routePush, replace: routeReplace }),
}));

vi.mock("@/composables/useMarketCountdown", () => ({
  useMarketCountdown: () => ({
    timeText: ref("00:00:00"),
    label: ref("PRÓXIMO CIERRE"),
    subtitle: ref("Mercado cierra hoy"),
    status: ref(null),
  }),
}));

vi.mock("./primitives/Logo.vue", () => ({
  default: { template: "<div class='logo-stub'></div>" },
}));
vi.mock("./primitives/TabBar.vue", () => ({
  default: {
    props: ["active"],
    template: "<div class='tabbar-stub' :data-active='active'></div>",
    emits: ["select"],
  },
}));
vi.mock("./MarketCountdown.vue", () => ({
  default: { template: "<div class='countdown-stub'></div>" },
}));

import AppShell from "./AppShell.vue";

beforeEach(() => {
  routePush.mockReset();
  routeReplace.mockReset();
  routePath.value = "/leagues";
});

describe("AppShell", () => {
 
  it("renderiza siempre sidebar, mobile-topbar y tab-bar (responsive vía CSS)", () => {
    const wrapper = mount(AppShell);

    expect(wrapper.find("aside.sidebar").exists()).toBe(true);
    expect(wrapper.find("header.mobile-topbar").exists()).toBe(true);
    expect(wrapper.find(".tabbar-stub").exists()).toBe(true);
  });

  
  it("marca el nav-item activo según la ruta actual", async () => {
    routePath.value = "/standings";
    const wrapper = mount(AppShell);
    await flushPromises();

   
    const navItems = wrapper.findAll("nav.nav button.nav-item");
    expect(navItems.length).toBe(4); // Ligas, Clasificación, Equipo, Mercado

    const active = navItems.find((b) => b.classes().includes("active"));
    if (!active) throw new Error("No hay ningún nav-item activo");

    expect(active.text()).toBe("Clasificación");
  });

  
  it("al hacer click en un nav-item navega con router.push", async () => {
    routePath.value = "/leagues"; // estamos en otra ruta
    const wrapper = mount(AppShell);
    await flushPromises();

    const navItems = wrapper.findAll("nav.nav button.nav-item");
    const mercadoItem = navItems.find((b) => b.text().includes("Mercado"));
    if (!mercadoItem) throw new Error("No se encontró el nav-item de Mercado");

    await mercadoItem.trigger("click");

    expect(routePush).toHaveBeenCalledWith("/market");
  });

  it("al hacer click en logout limpia el token y redirige a /auth", async () => {
    localStorage.setItem("token", "fake-jwt-token");

    const wrapper = mount(AppShell);
    await flushPromises();

    const logoutBtn = wrapper.find("aside.sidebar button.logout-btn");
    expect(logoutBtn.exists()).toBe(true);

    await logoutBtn.trigger("click");

    expect(localStorage.getItem("token")).toBeNull();
    expect(routePush).toHaveBeenCalledWith("/auth");
  });
});