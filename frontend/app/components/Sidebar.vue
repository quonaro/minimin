<template>
  <aside
    class="w-64 bg-white dark:bg-background-dark text-gray-900 dark:text-white flex flex-col h-screen fixed left-0 top-0 transition-colors duration-200 border-r border-gray-200 dark:border-border-dark"
  >
    <div class="p-4 border-b border-gray-200 dark:border-border-dark">
      <div class="flex items-center justify-between mb-4">
        <img src="/img/MiniMin_L.avif" alt="MiniMin" class="h-10 w-auto" />
        <img
          :src="
            colorMode.value === 'dark'
              ? '/img/MiniMin_T_light.avif'
              : '/img/MiniMin_T.avif'
          "
          alt="MiniMin"
          class="h-9 w-auto"
        />
      </div>
      <select
        v-model="selectedAgentId"
        @change="onAgentChange"
        :disabled="agents.length === 0"
        class="w-full px-3 py-2 rounded-lg bg-gray-100 dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <option value="">
          {{ agents.length === 0 ? "Agents not found" : "Select Agent" }}
        </option>
        <option v-for="agent in agents" :key="agent.id" :value="agent.id">
          {{ agent.name }}
        </option>
      </select>
    </div>

    <!-- Agent Nav -->
    <nav v-if="selectedAgentId" class="flex-1 p-4 space-y-2">
      <NuxtLink
        :to="`/agent/${selectedAgentId}/servers`"
        class="flex items-center gap-3 px-3 py-2 rounded-xl transition-colors hover:bg-gray-100 dark:hover:bg-gray-800"
        :class="
          isAgentNavActive('servers')
            ? 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white'
            : ''
        "
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="w-5 h-5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"
          />
        </svg>
        <span>Servers</span>
      </NuxtLink>
    </nav>

    <!-- Empty State -->
    <div v-else class="flex-1 flex items-center justify-center p-4">
      <p class="text-sm text-gray-500 dark:text-gray-400 text-center">
        Select an agent to get started
      </p>
    </div>

    <div
      class="p-4 border-t border-gray-200 dark:border-border-dark flex items-center justify-center gap-2"
    >
      <button
        @click="toggleColorMode"
        class="group flex items-center gap-2 px-2 py-2 rounded-xl transition-all hover:bg-gray-100 dark:hover:bg-gray-800"
      >
        <svg
          v-if="colorMode.value === 'dark'"
          xmlns="http://www.w3.org/2000/svg"
          class="w-5 h-5 shrink-0"
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
          xmlns="http://www.w3.org/2000/svg"
          class="w-5 h-5 shrink-0"
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
        <span
          class="max-w-0 overflow-hidden opacity-0 transition-all duration-300 whitespace-nowrap group-hover:max-w-40 group-hover:opacity-100 group-hover:ml-1"
        >
          {{ colorMode.value === "dark" ? "Light Mode" : "Dark Mode" }}
        </span>
      </button>
      <button
        @click="logout"
        class="group flex items-center gap-2 px-2 py-2 rounded-xl transition-all hover:bg-red-500/10 dark:hover:bg-red-500/20 text-red-600 dark:text-red-400"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="w-5 h-5 shrink-0"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
          />
        </svg>
        <span
          class="max-w-0 overflow-hidden opacity-0 transition-all duration-300 whitespace-nowrap group-hover:max-w-40 group-hover:opacity-100 group-hover:ml-1"
        >
          Logout
        </span>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
const colorMode = useColorMode();
const route = useRoute();
const { logout } = useAuth();

function toggleColorMode() {
  colorMode.preference = colorMode.value === "dark" ? "light" : "dark";
}

const { data: agentsData } = await useFetch("/agents", {
  baseURL: useApiBase(),
  credentials: "include",
  server: false,
  key: "agents",
});

const agents = computed(() => {
  if (
    agentsData.value &&
    typeof agentsData.value === "object" &&
    "body" in agentsData.value
  ) {
    return (agentsData.value as any).body || [];
  }
  return [];
});

const selectedAgentId = ref("");

// Sync selected agent with route
watch(
  () => route.params.id as string,
  (newId) => {
    if (newId) {
      selectedAgentId.value = newId;
    }
  },
  { immediate: true },
);

function onAgentChange() {
  if (selectedAgentId.value) {
    navigateTo(`/agent/${selectedAgentId.value}/servers`);
  }
}

function isAgentNavActive(segment: string) {
  return route.path.includes(`/agent/${selectedAgentId.value}/${segment}`);
}
</script>
