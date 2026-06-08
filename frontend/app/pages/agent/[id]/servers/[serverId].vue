<template>
  <div class="p-6">
    <div class="mb-6">
      <NuxtLink
        :to="`/agent/${agentId}/servers`"
        class="inline-flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors mb-4"
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
        <span>Back to servers</span>
      </NuxtLink>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
        Server Details
      </h1>
    </div>

    <div
      v-if="server"
      class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl p-6"
    >
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <span class="font-semibold text-gray-700 dark:text-gray-300"
            >Server ID:</span
          >
          <span class="ml-2 text-gray-900 dark:text-white">{{
            server.server_id
          }}</span>
        </div>
        <div>
          <span class="font-semibold text-gray-700 dark:text-gray-300"
            >Agent ID:</span
          >
          <span class="ml-2 text-gray-900 dark:text-white">{{ agentId }}</span>
        </div>
        <div>
          <span class="font-semibold text-gray-700 dark:text-gray-300"
            >Status:</span
          >
          <span
            :class="getStatusColor(server.status)"
            class="px-2 py-1 rounded text-sm ml-2"
          >
            {{ server.status }}
          </span>
        </div>
        <div>
          <span class="font-semibold text-gray-700 dark:text-gray-300"
            >Game Port:</span
          >
          <span class="ml-2 text-gray-900 dark:text-white">{{
            server.game_port
          }}</span>
        </div>
        <div>
          <span class="font-semibold text-gray-700 dark:text-gray-300"
            >Engine:</span
          >
          <span class="ml-2 text-gray-900 dark:text-white">{{
            server.engine_type
          }}</span>
        </div>
        <div>
          <span class="font-semibold text-gray-700 dark:text-gray-300"
            >Version:</span
          >
          <span class="ml-2 text-gray-900 dark:text-white">{{
            server.game_version
          }}</span>
        </div>
      </div>
      <div class="mt-6 flex gap-2">
        <button
          class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors font-medium"
        >
          Start
        </button>
        <button
          class="bg-red-500 text-white px-4 py-2 rounded-lg hover:bg-red-600 transition-colors font-medium"
        >
          Stop
        </button>
        <button
          class="bg-blue-500 text-white px-4 py-2 rounded-lg hover:bg-blue-600 transition-colors font-medium"
        >
          Restart
        </button>
      </div>
    </div>
    <div v-else class="text-gray-500 dark:text-gray-400">Loading...</div>
  </div>
</template>

<script setup lang="ts">
interface Server {
  server_id: string;
  status: string;
  game_port: number;
  engine_type: string;
  game_version: string;
}

const route = useRoute();
const { agentId } = useCurrentAgent();

const serverId = route.params.serverId as string;
const server = ref<Server | null>(null);

const { data } = await useApiFetch<{ body: Server }>(
  `/agent/${agentId.value}/servers/${serverId}`,
);

if (data.value && typeof data.value === "object" && "body" in data.value) {
  server.value = (data.value as any).body as Server;
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
