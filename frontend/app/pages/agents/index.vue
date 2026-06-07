<template>
  <div class="p-8">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">
          Agents
        </h1>
        <p class="text-gray-600 dark:text-gray-400">
          Manage your agent connections
        </p>
      </div>
      <button
        class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors font-medium"
      >
        Add Agent
      </button>
    </div>

    <div v-if="agents.length === 0" class="text-center py-12">
      <p class="text-gray-500 dark:text-gray-400">No agents found</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="agent in agents"
        :key="agent.id"
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 hover:shadow-md transition-shadow"
      >
        <div class="mb-4">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-1">
            {{ agent.name }}
          </h3>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ agent.host }}
          </p>
        </div>

        <div class="space-y-2 mb-4">
          <div
            class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400"
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
                d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
            <span
              >Created:
              {{ new Date(agent.created_at).toLocaleDateString() }}</span
            >
          </div>
        </div>

        <div class="flex gap-2">
          <button
            class="flex-1 px-3 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white border border-gray-200 dark:border-gray-700 rounded-lg transition-colors"
          >
            Edit
          </button>
          <button
            class="flex-1 px-3 py-2 text-sm text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 border border-red-200 dark:border-red-900/30 rounded-lg transition-colors"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

interface Agent {
  id: string;
  name: string;
  host: string;
  created_at: string;
}

const agents = ref<Agent[]>([]);

const { data } = await useFetch("/agents", {
  baseURL: useApiBase(),
  credentials: "include",
});

if (data.value && typeof data.value === "object" && "body" in data.value) {
  agents.value = (data.value as any).body;
}
</script>
