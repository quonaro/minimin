<template>
  <div class="p-6">
    <h1 class="text-2xl font-bold mb-4">Server Details</h1>
    <div v-if="server" class="bg-white border border-gray-300 rounded p-6">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <span class="font-semibold">Server ID:</span> {{ server.server_id }}
        </div>
        <div>
          <span class="font-semibold">Agent ID:</span> {{ server.agent_id }}
        </div>
        <div>
          <span class="font-semibold">Status:</span>
          <span
            :class="getStatusColor(server.status)"
            class="px-2 py-1 rounded text-sm ml-2"
          >
            {{ server.status }}
          </span>
        </div>
        <div>
          <span class="font-semibold">Game Port:</span> {{ server.game_port }}
        </div>
        <div>
          <span class="font-semibold">Engine:</span> {{ server.engine }}
        </div>
        <div>
          <span class="font-semibold">Version:</span> {{ server.version }}
        </div>
      </div>
      <div class="mt-6 flex gap-2">
        <button
          class="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
        >
          Start
        </button>
        <button
          class="bg-yellow-500 text-white px-4 py-2 rounded hover:bg-yellow-600"
        >
          Stop
        </button>
        <button
          class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          Restart
        </button>
      </div>
    </div>
    <div v-else class="text-gray-500">Loading...</div>
  </div>
</template>

<script setup lang="ts">
interface Server {
  id: string;
  server_id: string;
  agent_id: string;
  status: string;
  game_port: number;
  engine: string;
  version: string;
}

const route = useRoute();
const server = ref<Server | null>(null);

const { data } = await useFetch(`/servers`, {
  baseURL: useApiBase(),
  credentials: "include",
});

if (data.value && typeof data.value === "object" && "body" in data.value) {
  const servers = (data.value as any).body as Server[];
  server.value = servers.find((s: Server) => s.id === route.params.id) || null;
}

function getStatusColor(status: string) {
  switch (status) {
    case "running":
      return "bg-green-100 text-green-800";
    case "exited":
      return "bg-red-100 text-red-800";
    default:
      return "bg-gray-100 text-gray-800";
  }
}
</script>
