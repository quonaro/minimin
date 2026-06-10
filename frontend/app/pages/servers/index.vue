<template>
  <div class="p-8">
    <div class="mb-8 flex items-start justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">
          Servers
        </h1>
        <p class="text-gray-600 dark:text-neutral-400">
          Manage your Minecraft servers
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
      <p class="text-gray-500 dark:text-neutral-400">No servers found</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <ServerCard
        v-for="server in servers"
        :key="server.serverId"
        :server="server"
      />
    </div>

    <CreateServerModal
      v-if="showModal"
      @close="showModal = false"
      @created="onCreated"
    />
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

const { pending } = useServers();
const showModal = ref(false);

const { servers } = useServers();

function onCreated(serverId: string) {
  showModal.value = false;
  if (serverId) {
    navigateTo(`/servers/${serverId}`);
  }
}
</script>
