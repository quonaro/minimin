<template>
  <div class="p-8">
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">
        Servers
      </h1>
      <p class="text-gray-600 dark:text-gray-400">
        Servers for agent <span class="font-semibold">{{ agentName }}</span>
      </p>
    </div>

    <div v-if="pending" class="text-center py-12">
      <p class="text-gray-500 dark:text-gray-400">Loading...</p>
    </div>

    <div v-else-if="servers.length === 0" class="text-center py-12">
      <p class="text-gray-500 dark:text-gray-400">
        No servers found for this agent
      </p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="server in servers"
        :key="server.server_id"
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 hover:shadow-md transition-shadow"
      >
        <div class="flex items-start justify-between mb-4">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ server.server_id }}
            </h3>
          </div>
          <span
            :class="getStatusColor(server.status)"
            class="px-3 py-1 rounded-full text-xs font-medium"
          >
            {{ server.status }}
          </span>
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
                d="M5 12h14"
              />
            </svg>
            <span>Port: {{ server.game_port }}</span>
          </div>
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
                d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"
              />
            </svg>
            <span>{{ server.engine_type }}</span>
          </div>
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
                d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
              />
            </svg>
            <span>{{ server.game_version }}</span>
          </div>
        </div>

        <NuxtLink
          :to="`/agent/${agentId}/servers/${server.server_id}`"
          class="inline-flex items-center gap-2 text-primary hover:text-primary/80 font-medium text-sm"
        >
          View Details
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
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

interface Server {
  server_id: string;
  status: string;
  game_port: number;
  engine_type: string;
  game_version: string;
}

const { agentId, agent, pending } = useCurrentAgent();

const agentName = computed(() => agent.value?.name ?? "");

const servers = ref<Server[]>([]);

const { data } = await useApiFetch<Server[] | { body: Server[] }>(
  `/agent/${agentId.value}/servers`,
);

if (Array.isArray(data.value)) {
  servers.value = data.value;
} else if (
  data.value &&
  typeof data.value === "object" &&
  "body" in data.value
) {
  servers.value = (data.value as any).body as Server[];
}

function getStatusColor(status: string) {
  switch (status) {
    case "running":
      return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
    case "exited":
      return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
    default:
      return "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300";
  }
}
</script>
