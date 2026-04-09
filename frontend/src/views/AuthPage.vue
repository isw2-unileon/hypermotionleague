<template>
  <div class="min-h-screen bg-gradient-to-br from-green-900 via-green-800 to-emerald-900 flex items-center justify-center p-4">
    <!-- Football field lines background -->
    <div class="absolute inset-0 overflow-hidden opacity-10">
      <div class="absolute top-1/2 left-0 right-0 h-px bg-white"></div>
      <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-48 h-48 border-2 border-white rounded-full"></div>
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-72 h-32 border-2 border-white border-t-0"></div>
      <div class="absolute bottom-0 left-1/2 -translate-x-1/2 w-72 h-32 border-2 border-white border-b-0"></div>
    </div>

    <div class="relative w-full max-w-md">
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-20 h-20 bg-white/10 backdrop-blur-sm rounded-full mb-4 border-2 border-amber-400">
          <span class="text-4xl">⚽</span>
        </div>
        <h1 class="text-3xl font-bold text-white tracking-tight">
          HyperMotion League
        </h1>
        <p class="text-green-200 mt-1 text-sm">La Liga Hypermotion — Fantasy Manager</p>
      </div>

      <!-- Card -->
      <div class="bg-white/10 backdrop-blur-md rounded-2xl shadow-2xl border border-white/20 p-8">
        <!-- Tab switcher -->
        <div class="flex bg-white/10 rounded-lg p-1 mb-6">
          <button
            @click="activeTab = 'login'"
            :class="[
              'flex-1 py-2 text-sm font-semibold rounded-md transition-all',
              activeTab === 'login'
                ? 'bg-amber-500 text-white shadow-md'
                : 'text-green-200 hover:text-white',
            ]"
          >
            Iniciar Sesión
          </button>
          <button
            @click="activeTab = 'register'"
            :class="[
              'flex-1 py-2 text-sm font-semibold rounded-md transition-all',
              activeTab === 'register'
                ? 'bg-amber-500 text-white shadow-md'
                : 'text-green-200 hover:text-white',
            ]"
          >
            Registrarse
          </button>
        </div>

        <!-- Error message -->
        <div
          v-if="error"
          class="mb-4 p-3 bg-red-500/20 border border-red-400/30 rounded-lg text-red-200 text-sm"
        >
          {{ error }}
        </div>

        <!-- Success message -->
        <div
          v-if="success"
          class="mb-4 p-3 bg-green-500/20 border border-green-400/30 rounded-lg text-green-200 text-sm"
        >
          {{ success }}
        </div>

        <!-- OAuth buttons -->
        <div class="space-y-3 mb-6">
          <button
            @click="handleOAuth('google')"
            :disabled="loading"
            class="w-full flex items-center justify-center gap-3 py-3 bg-white hover:bg-gray-100 disabled:opacity-50 text-gray-800 font-semibold rounded-lg transition-all shadow-md"
          >
            <svg class="w-5 h-5" viewBox="0 0 24 24">
              <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"/>
              <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
              <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
            </svg>
            Continuar con Google
          </button>
          <button
            @click="handleOAuth('apple')"
            :disabled="loading"
            class="w-full flex items-center justify-center gap-3 py-3 bg-black hover:bg-gray-900 disabled:opacity-50 text-white font-semibold rounded-lg transition-all shadow-md"
          >
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M17.05 20.28c-.98.95-2.05.88-3.08.4-1.09-.5-2.08-.48-3.24 0-1.44.62-2.2.44-3.06-.4C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z"/>
            </svg>
            Continuar con Apple
          </button>
        </div>

        <!-- Divider -->
        <div class="flex items-center gap-3 mb-6">
          <div class="flex-1 h-px bg-white/20"></div>
          <span class="text-green-300/60 text-xs uppercase tracking-wider">o con email</span>
          <div class="flex-1 h-px bg-white/20"></div>
        </div>

        <!-- Login Form -->
        <form v-if="activeTab === 'login'" @submit.prevent="handleLogin" class="space-y-4">
          <div>
            <label class="block text-green-200 text-sm font-medium mb-1">Email</label>
            <input
              v-model="loginForm.email"
              type="email"
              required
              placeholder="tu@email.com"
              class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-green-300/50 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent transition"
            />
          </div>
          <div>
            <label class="block text-green-200 text-sm font-medium mb-1">Contraseña</label>
            <input
              v-model="loginForm.password"
              type="password"
              required
              placeholder="••••••••"
              class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-green-300/50 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent transition"
            />
          </div>
          <button
            type="submit"
            :disabled="loading"
            class="w-full py-3 bg-amber-500 hover:bg-amber-400 disabled:bg-amber-500/50 text-white font-bold rounded-lg transition-all shadow-lg hover:shadow-amber-500/25"
          >
            {{ loading ? "Entrando..." : "Entrar al campo" }}
          </button>
        </form>

        <!-- Register Form -->
        <form v-if="activeTab === 'register'" @submit.prevent="handleRegister" class="space-y-4">
          <div>
            <label class="block text-green-200 text-sm font-medium mb-1">Nombre de usuario</label>
            <input
              v-model="registerForm.username"
              type="text"
              required
              minlength="3"
              maxlength="50"
              placeholder="tu_nombre"
              class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-green-300/50 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent transition"
            />
          </div>
          <div>
            <label class="block text-green-200 text-sm font-medium mb-1">Nombre visible</label>
            <input
              v-model="registerForm.display_name"
              type="text"
              required
              maxlength="100"
              placeholder="Nombre que verán los demás"
              class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-green-300/50 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent transition"
            />
          </div>
          <div>
            <label class="block text-green-200 text-sm font-medium mb-1">Email</label>
            <input
              v-model="registerForm.email"
              type="email"
              required
              placeholder="tu@email.com"
              class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-green-300/50 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent transition"
            />
          </div>
          <div>
            <label class="block text-green-200 text-sm font-medium mb-1">Contraseña</label>
            <input
              v-model="registerForm.password"
              type="password"
              required
              minlength="8"
              placeholder="Mínimo 8 caracteres"
              class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-green-300/50 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent transition"
            />
          </div>
          <button
            type="submit"
            :disabled="loading"
            class="w-full py-3 bg-amber-500 hover:bg-amber-400 disabled:bg-amber-500/50 text-white font-bold rounded-lg transition-all shadow-lg hover:shadow-amber-500/25"
          >
            {{ loading ? "Registrando..." : "Fichar como mánager" }}
          </button>
        </form>
      </div>

      <!-- Footer -->
      <p class="text-center text-green-300/50 text-xs mt-6">
        LaLiga Hypermotion Fantasy Manager &copy; 2026
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from "vue";
import type { Subscription } from "@supabase/supabase-js";
import { supabase } from "@/lib/supabase";

const activeTab = ref<"login" | "register">("login");
const loading = ref(false);
const error = ref("");
const success = ref("");
let authSubscription: Subscription | null = null;

const loginForm = reactive({
  email: "",
  password: "",
});

const registerForm = reactive({
  username: "",
  display_name: "",
  email: "",
  password: "",
});

function clearMessages() {
  error.value = "";
  success.value = "";
}

// Send the Supabase access token to our backend to get our own JWT
async function handleOAuthCallback(accessToken: string) {
  loading.value = true;
  clearMessages();

  try {
    const res = await fetch("/api/v1/auth/oauth", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ access_token: accessToken }),
    });

    const result = await res.json();

    if (!res.ok) {
      error.value = result.error || "Error en autenticación OAuth";
      return;
    }

    localStorage.setItem("token", result.token);
    success.value = `¡Bienvenido, ${result.user.display_name}!`;

    // Clean up the Supabase session since we use our own JWT
    await supabase.auth.signOut();
  } catch {
    error.value = "Error de conexión con el servidor";
  } finally {
    loading.value = false;
  }
}

// Listen for OAuth redirect — fires when Supabase processes the token from the URL hash
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

async function handleOAuth(provider: "google" | "apple") {
  clearMessages();
  loading.value = true;

  const { error: oauthError } = await supabase.auth.signInWithOAuth({
    provider,
    options: {
      redirectTo: window.location.origin + "/auth",
    },
  });

  if (oauthError) {
    error.value = oauthError.message;
    loading.value = false;
  }
  // If no error, the browser will redirect to the provider
}

async function handleLogin() {
  clearMessages();
  loading.value = true;

  try {
    const res = await fetch("/api/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(loginForm),
    });

    const data = await res.json();

    if (!res.ok) {
      error.value = data.error || "Credenciales incorrectas";
      return;
    }

    localStorage.setItem("token", data.token);
    success.value = `¡Bienvenido, ${data.user.display_name}!`;
  } catch {
    error.value = "Error de conexión con el servidor";
  } finally {
    loading.value = false;
  }
}

async function handleRegister() {
  clearMessages();
  loading.value = true;

  try {
    const res = await fetch("/api/v1/auth/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(registerForm),
    });

    const data = await res.json();

    if (!res.ok) {
      error.value = data.error || "Error al registrarse";
      return;
    }

    localStorage.setItem("token", data.token);
    success.value = `¡Cuenta creada! Bienvenido, ${data.user.display_name}`;
  } catch {
    error.value = "Error de conexión con el servidor";
  } finally {
    loading.value = false;
  }
}
</script>
