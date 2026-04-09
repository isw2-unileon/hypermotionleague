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
import { ref, reactive } from "vue";

const activeTab = ref<"login" | "register">("login");
const loading = ref(false);
const error = ref("");
const success = ref("");

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
