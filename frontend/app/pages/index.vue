<script setup lang="ts">
import type { Server } from "~/composables/useServers";

definePageMeta({
  middleware: "auth",
});

const showModal = ref(false);

const { data: rawServers } = await useFetch<Server[]>("/servers", {
  baseURL: useApiBase(),
  credentials: "include",
});

const servers = computed<Server[]>(() => {
  if (!rawServers.value) return [];
  return Array.isArray(rawServers.value)
    ? rawServers.value
    : (rawServers.value as any).body || [];
});

const firstServer = servers.value[0];
if (firstServer) {
  await navigateTo(`/servers/${firstServer.serverId}`, { replace: true });
}

function onCreated(serverId: string) {
  showModal.value = false;
  if (serverId) {
    navigateTo(`/servers/${serverId}`);
  }
}
</script>

<template>
  <div
    v-if="servers.length === 0"
    class="flex flex-col items-center justify-center h-full p-8"
  >
    <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">
      No Servers
    </h1>
    <p class="text-gray-600 dark:text-neutral-400 mb-6">
      You don't have any servers yet.
    </p>
    <button
      class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors font-medium"
      @click="showModal = true"
    >
      Create Server
    </button>
    <LazyCreateServerModal
      v-if="showModal"
      @close="showModal = false"
      @created="onCreated"
    />
  </div>
</template>
