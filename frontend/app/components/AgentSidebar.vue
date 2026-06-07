<template>
  <aside
    class="w-64 bg-white dark:bg-gray-900 text-gray-900 dark:text-white flex flex-col h-screen fixed left-0 top-0 transition-colors duration-200 border-r border-gray-200 dark:border-gray-700"
  >
    <div class="p-4 border-b border-gray-200 dark:border-gray-700">
      <NuxtLink
        to="/agents"
        class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors mb-3"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="w-4 h-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 19l-7-7m0 0l7-7m-7 7h18"
          />
        </svg>
        <span>Back to agents</span>
      </NuxtLink>
      <h1 class="text-lg font-bold text-primary truncate" :title="agentName">
        {{ agentName }}
      </h1>
      <p class="text-xs text-gray-500 dark:text-gray-400 truncate">
        {{ agentHost }}
      </p>
    </div>

    <nav class="flex-1 p-4 space-y-2">
      <NuxtLink
        :to="`/agent/${agentId}/servers`"
        class="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors hover:bg-primary/10 dark:hover:bg-primary/20"
        :class="isActive('servers') ? 'bg-primary/20 text-primary' : ''"
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

    <div class="p-4 border-t border-gray-200 dark:border-gray-700 space-y-2">
      <button
        @click="toggleColorMode"
        class="flex items-center gap-3 px-4 py-3 w-full rounded-lg transition-colors hover:bg-primary/10 dark:hover:bg-primary/20"
      >
        <svg
          v-if="colorMode.value === 'dark'"
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
            d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
          />
        </svg>
        <svg
          v-else
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
            d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
          />
        </svg>
        <span>{{ colorMode.value === "dark" ? "Light Mode" : "Dark Mode" }}</span>
      </button>
      <button
        @click="logout"
        class="flex items-center gap-3 px-4 py-3 w-full rounded-lg transition-colors hover:bg-red-500/10 dark:hover:bg-red-500/20 text-red-600 dark:text-red-400"
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
            d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
          />
        </svg>
        <span>Logout</span>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
const colorMode = useColorMode();
const route = useRoute();
const { logout } = useAuth();
const { agent } = useCurrentAgent();

const agentId = computed(() => route.params.id as string);
const agentName = computed(() => agent.value?.name ?? "Agent");
const agentHost = computed(() => agent.value?.host ?? "");

function toggleColorMode() {
  colorMode.preference = colorMode.value === "dark" ? "light" : "dark";
}

function isActive(segment: string) {
  return route.path.includes(`/agent/${agentId.value}/${segment}`);
}
</script>
