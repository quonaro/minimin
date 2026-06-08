<template>
  <div class="min-h-screen flex bg-gray-50 dark:bg-background-dark">
    <!-- Left Side -->
    <div
      class="hidden md:flex md:w-1/2 relative overflow-hidden bg-gradient-to-br from-gray-100 via-gray-200 to-primary-100 dark:from-gray-900 dark:via-gray-800 dark:to-primary-900"
    >
      <!-- Abstract background shapes -->
      <div class="absolute inset-0 opacity-20">
        <div
          class="absolute -top-24 -left-24 w-96 h-96 rounded-full bg-primary blur-3xl animate-blob"
        />
        <div
          class="absolute top-1/2 right-0 w-80 h-80 rounded-full bg-primary-300 blur-3xl animate-blob animation-delay-2000"
        />
        <div
          class="absolute -bottom-32 left-1/3 w-96 h-96 rounded-full bg-primary-700 blur-3xl animate-blob animation-delay-4000"
        />
      </div>

      <!-- Dot pattern overlay -->
      <div
        class="absolute inset-0 opacity-10 dark:hidden"
        style="
          background-image: radial-gradient(
            circle,
            rgba(0, 0, 0, 0.5) 1px,
            transparent 1px
          );
          background-size: 24px 24px;
        "
      />
      <div
        class="absolute inset-0 opacity-10 hidden dark:block"
        style="
          background-image: radial-gradient(
            circle,
            rgba(255, 255, 255, 0.5) 1px,
            transparent 1px
          );
          background-size: 24px 24px;
        "
      />

      <!-- Content -->
      <div class="relative z-10 flex flex-col justify-between p-12 w-full">
        <div class="flex items-center justify-center gap-4 pt-4">
          <img src="/img/MiniMin_L.avif" alt="MiniMin" class="h-20 w-auto" />
          <img
            src="/img/MiniMin_T.avif"
            alt="MiniMin"
            class="h-14 w-auto dark:hidden"
          />
          <img
            src="/img/MiniMin_T_light.avif"
            alt="MiniMin"
            class="h-14 w-auto hidden dark:block"
          />
        </div>

        <div class="max-w-lg">
          <h2
            class="text-4xl xl:text-5xl font-bold text-gray-900 dark:text-white leading-tight mb-6"
          >
            Manage your Minecraft servers with
            <span class="text-primary">ease</span>
          </h2>
          <p
            class="text-lg text-gray-600 dark:text-gray-300 leading-relaxed mb-8"
          >
            A lightweight, web-based control panel for your Minecraft server
            agents. Deploy, monitor, and manage multiple servers from a single
            dashboard.
          </p>
          <a
            href="https://github.com"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-2 px-5 py-3 rounded-lg bg-gray-900/10 hover:bg-gray-900/20 dark:bg-white/10 dark:hover:bg-white/20 backdrop-blur-sm border border-gray-900/10 dark:border-white/10 text-gray-900 dark:text-white transition-all group"
          >
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
              <path
                d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"
              />
            </svg>
            <span class="font-medium">View on GitHub</span>
            <svg
              class="w-4 h-4 transition-transform group-hover:translate-x-0.5"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3"
              />
            </svg>
          </a>
        </div>

        <div class="text-sm text-gray-500 dark:text-gray-400">
          &copy; {{ new Date().getFullYear() }} MiniMin. Open source.
        </div>
      </div>
    </div>

    <!-- Right Side -->
    <div
      class="w-full md:w-1/2 flex items-center justify-center p-6 lg:p-12 relative"
    >
      <button
        @click="toggleColorMode"
        class="absolute top-4 right-4 p-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        aria-label="Toggle theme"
      >
        <svg
          v-if="colorMode.value === 'dark'"
          class="w-5 h-5"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
          />
        </svg>
        <svg
          v-else
          class="w-5 h-5"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
          />
        </svg>
      </button>
      <div class="w-full max-w-sm">
        <div class="md:hidden text-center mb-8">
          <div class="mb-4 flex items-center justify-center gap-2">
            <img src="/img/MiniMin_L.avif" alt="MiniMin" class="h-10 w-auto" />
            <img
              src="/img/MiniMin_T.avif"
              alt="MiniMin"
              class="h-8 w-auto dark:hidden"
            />
            <img
              src="/img/MiniMin_T_light.avif"
              alt="MiniMin"
              class="h-8 w-auto hidden dark:block"
            />
          </div>
        </div>

        <div
          class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl border border-gray-200 dark:border-gray-700 p-8"
        >
          <div class="mb-8">
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
              Welcome back
            </h1>
            <p class="text-gray-600 dark:text-gray-400 text-sm">
              Sign in to your dashboard
            </p>
          </div>

          <form @submit.prevent="handleLogin" class="space-y-5">
            <div>
              <label
                for="apiKey"
                class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2"
              >
                Password
              </label>
              <div class="relative">
                <div
                  class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"
                >
                  <svg
                    class="w-5 h-5 text-gray-400"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z"
                    />
                  </svg>
                </div>
                <input
                  id="apiKey"
                  v-model="apiKeyInput"
                  type="password"
                  placeholder="Enter your password"
                  autocomplete="new-password"
                  class="w-full pl-10 pr-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-all"
                  required
                />
              </div>
            </div>

            <div
              v-if="error"
              class="flex items-center gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/50 rounded-lg"
            >
              <svg
                class="w-5 h-5 text-red-600 dark:text-red-400 shrink-0"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
                />
              </svg>
              <p class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
            </div>

            <button
              type="submit"
              :disabled="loading"
              class="w-full bg-primary text-white py-3 rounded-lg hover:bg-primary/90 transition-colors font-semibold disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-primary/25"
            >
              {{ loading ? "Signing in..." : "Sign In" }}
            </button>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const colorMode = useColorMode();

function toggleColorMode() {
  colorMode.preference = colorMode.value === "dark" ? "light" : "dark";
}
const { login, isAuthenticated } = useAuth();
const { show: showToast } = useToast();
const apiKeyInput = ref("");
const loading = ref(false);
const error = ref("");

// Redirect if already authenticated
if (isAuthenticated.value) {
  await navigateTo("/agents");
}

async function handleLogin() {
  loading.value = true;
  error.value = "";

  try {
    const success = await login(apiKeyInput.value);
    if (success) {
      showToast("success", "Signed in successfully");
      await navigateTo("/agents");
    } else {
      error.value = "Invalid password or connection failed";
      showToast("error", "Sign in failed", { description: error.value });
    }
  } catch (e) {
    error.value = "Invalid password or connection failed";
    showToast("error", "Sign in failed", { description: error.value });
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
@keyframes blob {
  0% {
    transform: translate(0px, 0px) scale(1);
  }
  33% {
    transform: translate(30px, -50px) scale(1.1);
  }
  66% {
    transform: translate(-20px, 20px) scale(0.9);
  }
  100% {
    transform: translate(0px, 0px) scale(1);
  }
}

.animate-blob {
  animation: blob 7s infinite;
}

.animation-delay-2000 {
  animation-delay: 2s;
}

.animation-delay-4000 {
  animation-delay: 4s;
}
</style>
