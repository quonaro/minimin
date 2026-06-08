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
      <ServerCard
        v-for="server in servers"
        :key="server.serverId"
        :server="server"
        :agent-id="agentId"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

interface Server {
  serverId: string;
  status: string;
  gamePort: number;
  engineType: string;
  gameVersion: string;
  startedAt?: string;
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
</script>
