<template>
  <div class="p-8">
    <div class="mb-8 flex items-start justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">
          Servers
        </h1>
        <p class="text-gray-600 dark:text-neutral-400">
          Servers for agent <span class="font-semibold">{{ agentName }}</span>
        </p>
      </div>
      <button
        class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors font-medium"
        @click="showModal = true"
      >
        Create Server
      </button>
    </div>

    <div v-if="pending" class="text-center py-12">
      <p class="text-gray-500 dark:text-neutral-400">Loading...</p>
    </div>

    <div v-else-if="servers.length === 0" class="text-center py-12">
      <p class="text-gray-500 dark:text-neutral-400">
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

    <CreateServerModal
      v-if="showModal"
      :agent-id="agentId"
      @close="showModal = false"
      @created="onCreated"
    />
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

const { agentId, agent, pending } = useCurrentAgent();

const agentName = computed(() => agent.value?.name ?? "");
const showModal = ref(false);

const { servers } = useServers(agentId);

function onCreated(serverId: string) {
  showModal.value = false;
  if (serverId) {
    navigateTo(`/agent/${agentId.value}/servers/${serverId}`);
  }
}
</script>
