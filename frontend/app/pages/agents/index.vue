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
        @click="showForm = true"
        class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors font-medium"
      >
        Add Agent
      </button>
    </div>

    <div v-if="agents.length === 0" class="text-center py-12">
      <p class="text-gray-500 dark:text-gray-400">No agents found</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <NuxtLink
        v-for="agent in agents"
        :key="agent.id"
        :to="`/agent/${agent.id}/`"
        class="block bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 hover:shadow-md transition-shadow cursor-pointer"
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
              {{ new Date(agent.createdAt).toLocaleDateString() }}</span
            >
          </div>
        </div>

        <div
          class="flex items-center gap-2 text-primary font-medium text-sm mt-4"
        >
          <span>Open agent</span>
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
              d="M9 5l7 7-7 7"
            />
          </svg>
        </div>
      </NuxtLink>
    </div>
    <AgentFormModal v-model="showForm" />
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
  createdAt: string;
}

const showForm = ref(false);

const { data: agentsData } = await useFetch("/agents", {
  baseURL: useApiBase(),
  credentials: "include",
  key: "agents",
});

const agents = computed<Agent[]>(() => {
  const val = agentsData.value;
  if (Array.isArray(val)) return val as Agent[];
  if (val && typeof val === "object" && "body" in val) {
    return (val as any).body || [];
  }
  return [];
});
</script>
