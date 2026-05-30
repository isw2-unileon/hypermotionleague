<template>
  <div class="frame auth-frame">
    <div class="halftone-bg" />

    <!-- Mobile-only system status row at the very top, per Login.jsx LoginMobile -->
    <StatusBar class="mobile-only status-bar-slot" />

    <div class="auth-grid">
      <!-- ===== HERO (desktop: left panel; mobile: top stack) ===== -->
      <section class="hero-panel">
        <PitchSVG :stroke="pitchStroke" />
        <div class="lime-glow" aria-hidden="true" />

        <header class="hero-header">
          <Logo :size="logoSize" />
          <span class="mono season-label desktop-only">TEMP. 25/26 · J·32 EN CURSO</span>
        </header>

        <div class="hero-copy">
          <p class="mono lime-tag">◆ FANTASY MANAGER · LIGA HYPERMOTION</p>
          <h1 class="display hero-title">
            <span class="hero-line">ENTRA</span>
            <span class="hero-line">AL <span class="lime">BANQUILLO<span class="desktop-only">.</span></span></span>
          </h1>
          <p class="hero-sub">
            <span class="desktop-only">La liga privada con tus panas. Fichajes, pujas en directo y la tabla que decide quién paga la cena. Segunda división, primera obsesión.</span>
            <span class="mobile-only">Tu liga privada con los panas. Fichajes, pujas y tabla, jornada a jornada.</span>
          </p>
        </div>

        <!-- Desktop-only stat strip. Static values per spec. -->
        <!-- TODO Sprint 2: replace static stats with /api/v1/platform/metrics -->
        <div class="stats desktop-only">
          <div class="mono stats-eyebrow">◆ TEMPORADA EN VIVO</div>
          <div class="stats-row">
            <div class="stat">
              <div class="mono stat-label">MÁNAGERS</div>
              <div class="display tnum stat-value">12.847</div>
            </div>
            <div class="stat">
              <div class="mono stat-label">LIGAS</div>
              <div class="display tnum stat-value">1.203</div>
            </div>
            <div class="stat">
              <div class="mono stat-label">PUJA RÉCORD</div>
              <div class="display tnum stat-value">€42M</div>
            </div>
          </div>
        </div>
      </section>

      <!-- ===== FORM PANEL ===== -->
      <section class="form-panel">
        <div class="form-header desktop-only">
          <span class="mono form-eyebrow">◆ ACCESO</span>
          <h2 class="display form-title">{{ activeTab === "login" ? "INICIAR SESIÓN" : "REGISTRO" }}</h2>
        </div>

        <!-- Tabs preserved from prior AuthPage (login | register) -->
        <div class="auth-tabs" role="tablist">
          <button
            type="button"
            role="tab"
            class="tab-pill mono"
            :class="{ active: activeTab === 'login' }"
            :aria-selected="activeTab === 'login'"
            @click="switchTab('login')"
          >
            Iniciar sesión
          </button>
          <button
            type="button"
            role="tab"
            class="tab-pill mono"
            :class="{ active: activeTab === 'register' }"
            :aria-selected="activeTab === 'register'"
            @click="switchTab('register')"
          >
            Registrarse
          </button>
        </div>

        <div v-if="error" class="error-banner" role="alert">{{ error }}</div>

        <div class="oauth-buttons">
          <button
            type="button"
            class="btn btn-secondary oauth-btn"
            :disabled="loading"
            @click="handleOAuth('google')"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" aria-hidden="true">
              <path fill="#fff" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
              <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path fill="#FBBC05" d="M5.84 14.1c-.22-.66-.35-1.36-.35-2.1s.13-1.44.35-2.1V7.07H2.18A10.97 10.97 0 0 0 1 12c0 1.77.42 3.45 1.18 4.93l3.66-2.83z"/>
              <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.83C6.71 7.31 9.14 5.38 12 5.38z"/>
            </svg>
            Continuar con Google
          </button>
          <button
            type="button"
            class="btn btn-secondary oauth-btn"
            :disabled="loading"
            @click="handleOAuth('apple')"
          >
            <svg width="16" height="18" viewBox="0 0 24 24" fill="#fff" aria-hidden="true">
              <path d="M17.05 12.04c-.03-3.04 2.49-4.5 2.6-4.57-1.42-2.07-3.62-2.36-4.4-2.39-1.87-.19-3.66 1.1-4.61 1.1-.96 0-2.42-1.08-3.99-1.05-2.05.03-3.96 1.2-5.02 3.04-2.14 3.71-.55 9.21 1.54 12.22 1.02 1.48 2.24 3.13 3.83 3.07 1.54-.06 2.12-1 3.99-1 1.86 0 2.39 1 4.02.97 1.66-.03 2.71-1.5 3.72-2.99 1.18-1.71 1.66-3.38 1.69-3.46-.04-.02-3.24-1.24-3.27-4.94zM14.07 4.18c.85-1.03 1.42-2.45 1.27-3.88-1.22.05-2.71.81-3.59 1.84-.79.91-1.49 2.37-1.3 3.76 1.36.11 2.76-.69 3.62-1.72z"/>
            </svg>
            Continuar con Apple
          </button>
        </div>

        <div class="divider-mono">O CON EMAIL</div>

        <!-- ============== LOGIN ============== -->
        <form v-if="activeTab === 'login'" class="auth-form" @submit.prevent="handleLogin">
          <label class="label" for="login-email">EMAIL</label>
          <input
            id="login-email"
            v-model="loginForm.email"
            class="input"
            type="email"
            required
            autocomplete="email"
            placeholder="tu@email.com"
          />
          <div class="field-gap" />

          <label class="label" for="login-password">CONTRASEÑA</label>
          <input
            id="login-password"
            v-model="loginForm.password"
            class="input"
            type="password"
            required
            autocomplete="current-password"
            placeholder="••••••••"
          />

          <div class="form-meta">
            <label class="remember">
              <!-- TODO Sprint 2: persist 'remember me' preference -->
              <input v-model="rememberMe" type="checkbox" />
              Mantener sesión
            </label>
            <!-- TODO Sprint 3: wire actual password reset endpoint -->
            <a class="forgot-link" href="#" @click.prevent>¿Olvidaste tu contraseña?</a>
          </div>

          <button type="submit" class="btn btn-primary submit-btn" :disabled="loading">
            {{ loading ? "Cargando..." : "Entrar al campo →" }}
          </button>
        </form>

        <!-- ============== REGISTER ============== -->
        <form v-else class="auth-form" @submit.prevent="handleRegister">
          <label class="label" for="reg-username">NOMBRE DE USUARIO</label>
          <input
            id="reg-username"
            v-model="registerForm.username"
            class="input"
            type="text"
            required
            minlength="3"
            maxlength="50"
            autocomplete="username"
            placeholder="tu_nombre"
          />
          <div class="field-gap" />

          <label class="label" for="reg-display">NOMBRE VISIBLE</label>
          <input
            id="reg-display"
            v-model="registerForm.display_name"
            class="input"
            type="text"
            required
            maxlength="100"
            placeholder="Como quieres que te vean"
          />
          <div class="field-gap" />

          <label class="label" for="reg-email">EMAIL</label>
          <input
            id="reg-email"
            v-model="registerForm.email"
            class="input"
            type="email"
            required
            autocomplete="email"
            placeholder="tu@email.com"
          />
          <div class="field-gap" />

          <label class="label" for="reg-password">CONTRASEÑA</label>
          <input
            id="reg-password"
            v-model="registerForm.password"
            class="input"
            type="password"
            required
            minlength="8"
            autocomplete="new-password"
            placeholder="Mínimo 8 caracteres"
          />

          <button type="submit" class="btn btn-primary submit-btn" :disabled="loading">
            {{ loading ? "Cargando..." : "Crear cuenta" }}
          </button>
        </form>

        <p class="alt-tab-cta">
          <template v-if="activeTab === 'login'">
            ¿Aún no juegas?
            <a class="lime-link" href="#" @click.prevent="switchTab('register')">Crea tu cuenta gratis</a>
          </template>
          <template v-else>
            ¿Ya tienes cuenta?
            <a class="lime-link" href="#" @click.prevent="switchTab('login')">Inicia sesión</a>
          </template>
        </p>
      </section>
    </div>

    <div class="mono footer-meta mobile-only">HML/26 — UNOFFICIAL FANTASY MANAGER</div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import type { Subscription } from "@supabase/supabase-js";
import { supabase } from "@/lib/supabase";
import Logo from "@/design-system/primitives/Logo.vue";
import PitchSVG from "@/design-system/primitives/PitchSVG.vue";
import StatusBar from "@/design-system/primitives/StatusBar.vue";

type AuthTab = "login" | "register";
type OAuthProvider = "google" | "apple";

interface LoginForm {
  email: string;
  password: string;
}

interface RegisterForm {
  username: string;
  display_name: string;
  email: string;
  password: string;
}

interface AuthResponse {
  token: string;
  error?: string;
}

const router = useRouter();
const activeTab = ref<AuthTab>("login");
const loading = ref(false);
const error = ref("");
const rememberMe = ref(true);

const loginForm = reactive<LoginForm>({ email: "", password: "" });
const registerForm = reactive<RegisterForm>({
  username: "",
  display_name: "",
  email: "",
  password: "",
});

let authSubscription: Subscription | null = null;

// Pitch SVG stroke is slightly more opaque on desktop than mobile, per Login.jsx.
// CSS handles the visual context (mobile/desktop) but the stroke value itself
// is a fixed string passed to PitchSVG; the desktop tone is a hair stronger.
const pitchStroke = computed<string>(() => "rgba(199,255,61,0.07)");
const logoSize = 28;

function clearError(): void {
  error.value = "";
}

function switchTab(tab: AuthTab): void {
  if (activeTab.value === tab) return;
  activeTab.value = tab;
  clearError();
}

async function handleLogin(): Promise<void> {
  clearError();
  loading.value = true;
  try {
    const res = await fetch("/api/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(loginForm),
    });
    const data: AuthResponse = await res.json();
    if (!res.ok) {
      error.value = data.error ?? "Credenciales incorrectas";
      return;
    }
    localStorage.setItem("token", data.token);
    router.push("/leagues");
  } catch {
    error.value = "Error de conexión con el servidor";
  } finally {
    loading.value = false;
  }
}

async function handleRegister(): Promise<void> {
  clearError();
  loading.value = true;
  try {
    const res = await fetch("/api/v1/auth/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(registerForm),
    });
    const data: AuthResponse = await res.json();
    if (!res.ok) {
      error.value = data.error ?? "Error al registrarse";
      return;
    }
    localStorage.setItem("token", data.token);
    router.push("/leagues");
  } catch {
    error.value = "Error de conexión con el servidor";
  } finally {
    loading.value = false;
  }
}

// OAuth flow: Supabase JS SDK initiates the redirect to the provider. Once the
// provider sends the user back to /auth, onAuthStateChange fires with the
// resulting session, and we trade Supabase's access_token for our own JWT via
// POST /api/v1/auth/oauth. This is exactly the Sprint 1.5 hardened flow — do
// not refactor.
async function handleOAuthCallback(accessToken: string): Promise<void> {
  loading.value = true;
  clearError();
  try {
    const res = await fetch("/api/v1/auth/oauth", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ access_token: accessToken }),
    });
    const result: AuthResponse = await res.json();
    if (!res.ok) {
      error.value = result.error ?? "Error en autenticación OAuth";
      return;
    }
    localStorage.setItem("token", result.token);
    await supabase.auth.signOut();
    router.push("/leagues");
  } catch {
    error.value = "Error de conexión con el servidor";
  } finally {
    loading.value = false;
  }
}

async function handleOAuth(provider: OAuthProvider): Promise<void> {
  clearError();
  loading.value = true;
  const { error: oauthError } = await supabase.auth.signInWithOAuth({
    provider,
    options: { redirectTo: window.location.origin + "/auth" },
  });
  if (oauthError) {
    error.value = oauthError.message;
    loading.value = false;
  }
  // No error: browser is now navigating to the provider; loading stays true.
}

onMounted(() => {
  const { data } = supabase.auth.onAuthStateChange(async (event, session) => {
    if (event === "SIGNED_IN" && session?.access_token) {
      await handleOAuthCallback(session.access_token);
    }
  });
  authSubscription = data.subscription;
});

onUnmounted(() => {
  authSubscription?.unsubscribe();
});
</script>

<style scoped>
/* ============================================================
   Layout primitives
   ============================================================ */
.auth-frame {
  min-height: 100vh;
  position: relative;
  background: var(--ink-900);
  color: var(--ink-100);
  font-family: var(--f-ui);
  overflow-x: hidden;
}

.status-bar-slot {
  position: relative;
  z-index: 1;
}

.auth-grid {
  position: relative;
  display: grid;
  grid-template-columns: 1fr;
  min-height: 100vh;
}

@media (min-width: 768px) {
  .auth-grid {
    grid-template-columns: 1fr 1fr;
  }
}

/* Mobile/desktop helpers — used pervasively */
.mobile-only { display: initial; }
.desktop-only { display: none; }

@media (min-width: 768px) {
  .mobile-only { display: none; }
  .desktop-only { display: initial; }
}

/* ============================================================
   Hero panel
   ============================================================ */
.hero-panel {
  position: relative;
  padding: 24px 24px 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

@media (min-width: 768px) {
  .hero-panel {
    padding: 56px 80px;
    justify-content: space-between;
    border-right: 1px solid var(--ink-700);
    min-height: 100vh;
  }
}

.lime-glow {
  position: absolute;
  top: -100px;
  right: -100px;
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, var(--lime-glow) 0%, transparent 60%);
  pointer-events: none;
}

@media (min-width: 768px) {
  .lime-glow {
    top: 30%;
    left: -20%;
    right: auto;
    width: 700px;
    height: 700px;
    background: radial-gradient(circle, var(--lime-glow) 0%, transparent 55%);
  }
}

.hero-header {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  z-index: 1;
}

.season-label {
  font-size: 11px;
  color: var(--ink-300);
  letter-spacing: 0.2em;
}

.hero-copy {
  position: relative;
  z-index: 1;
  margin-top: 36px;
}

@media (min-width: 768px) {
  .hero-copy {
    margin-top: 0;
  }
}

.lime-tag {
  font-size: 10px;
  letter-spacing: 0.2em;
  color: var(--lime);
  margin: 0;
}

@media (min-width: 768px) {
  .lime-tag {
    font-size: 11px;
    margin-bottom: 16px;
  }
}

.hero-title {
  font-size: 76px;
  line-height: 0.86;
  margin: 10px 0 0;
  letter-spacing: 0.005em;
  display: flex;
  flex-direction: column;
}

@media (min-width: 768px) {
  .hero-title {
    font-size: 124px;
    margin: 0;
    letter-spacing: 0.01em;
  }
}

.hero-line {
  display: block;
}

.lime { color: var(--lime); }

.hero-sub {
  font-size: 13px;
  color: var(--ink-300);
  line-height: 1.5;
  margin-top: 14px;
  max-width: 300px;
}

@media (min-width: 768px) {
  .hero-sub {
    font-size: 16px;
    color: var(--ink-200);
    line-height: 1.55;
    margin-top: 24px;
    max-width: 460px;
  }
}

/* ============================================================
   Stats (desktop only)
   ============================================================ */
.stats {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.stats-eyebrow {
  font-size: 10px;
  color: var(--ink-400);
  letter-spacing: 0.18em;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(3, auto);
  gap: 0 28px;
  align-items: baseline;
}

.stat-label {
  font-size: 10px;
  color: var(--ink-400);
  letter-spacing: 0.15em;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 26px;
  color: var(--ink-100);
  line-height: 1;
}

/* ============================================================
   Form panel
   ============================================================ */
.form-panel {
  position: relative;
  z-index: 1;
  padding: 16px 24px 24px;
  display: flex;
  flex-direction: column;
}

@media (min-width: 768px) {
  .form-panel {
    background: var(--ink-850);
    padding: 72px 100px;
    justify-content: center;
    min-height: 100vh;
  }
}

.form-header {
  margin-bottom: 24px;
}

.form-eyebrow {
  font-size: 11px;
  letter-spacing: 0.2em;
  color: var(--ink-400);
}

.form-title {
  font-size: 56px;
  margin: 8px 0 0;
  letter-spacing: 0.02em;
}

/* ============================================================
   Tabs — mono pill switcher (lime active, transparent inactive)
   ============================================================ */
.auth-tabs {
  display: inline-flex;
  gap: 6px;
  margin-bottom: 16px;
  padding: 4px;
  background: var(--ink-800);
  border: 1px solid var(--ink-700);
  border-radius: var(--r-sm);
  align-self: flex-start;
}

.tab-pill {
  font-size: 11px;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  padding: 8px 14px;
  border-radius: var(--r-xs);
  background: transparent;
  border: none;
  color: var(--ink-300);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.tab-pill:hover:not(.active) {
  color: var(--ink-100);
}

.tab-pill.active {
  background: var(--lime);
  color: var(--ink-900);
  font-weight: 600;
}

/* ============================================================
   Error banner
   ============================================================ */
.error-banner {
  background: rgba(255, 98, 98, 0.12);
  border: 1px solid rgba(255, 98, 98, 0.35);
  color: var(--down);
  font-size: 13px;
  padding: 10px 12px;
  border-radius: var(--r-sm);
  margin-bottom: 14px;
}

/* ============================================================
   OAuth buttons + divider
   ============================================================ */
.oauth-buttons {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 18px;
}

.oauth-btn {
  height: 48px;
  justify-content: flex-start;
  padding: 0 18px;
  font-weight: 500;
}

@media (min-width: 768px) {
  .oauth-btn {
    height: 52px;
    padding: 0 20px;
  }
}

/* ============================================================
   Form fields
   ============================================================ */
.auth-form {
  display: flex;
  flex-direction: column;
}

.field-gap {
  height: 14px;
}

.form-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  flex-wrap: wrap;
  gap: 8px;
}

.remember {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--ink-300);
  cursor: pointer;
}

.remember input[type="checkbox"] {
  accent-color: var(--lime);
}

.forgot-link {
  font-size: 12px;
  color: var(--lime);
  text-decoration: none;
}

.submit-btn {
  margin-top: 18px;
  height: 52px;
  font-size: 15px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

@media (min-width: 768px) {
  .submit-btn {
    margin-top: 26px;
    height: 56px;
    letter-spacing: 0.06em;
  }
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ============================================================
   Alt-tab CTA + footer
   ============================================================ */
.alt-tab-cta {
  margin-top: 18px;
  font-size: 12px;
  color: var(--ink-300);
  text-align: center;
}

@media (min-width: 768px) {
  .alt-tab-cta {
    margin-top: 24px;
    font-size: 13px;
    text-align: left;
  }
}

.lime-link {
  color: var(--lime);
  text-decoration: none;
  font-weight: 600;
}

.footer-meta {
  position: relative;
  z-index: 1;
  font-size: 9px;
  color: var(--ink-500);
  text-align: center;
  padding: 16px;
  letter-spacing: 0.15em;
}
</style>
