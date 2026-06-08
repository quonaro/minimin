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
            server.serverId
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
          <span
            v-if="
              server.desiredStatus && server.desiredStatus !== server.status
            "
            class="ml-2 text-xs text-gray-500 dark:text-gray-400 italic"
          >
            ({{ server.desiredStatus }}…)
          </span>
        </div>
        <div>
          <span class="font-semibold text-gray-700 dark:text-gray-300"
            >Game Port:</span
          >
          <span class="ml-2 text-gray-900 dark:text-white">{{
            server.gamePort
          }}</span>
        </div>
        <div>
          <span class="font-semibold text-gray-700 dark:text-gray-300"
            >Engine:</span
          >
          <span class="ml-2 text-gray-900 dark:text-white">{{
            server.engineType
          }}</span>
        </div>
        <div>
          <span class="font-semibold text-gray-700 dark:text-gray-300"
            >Version:</span
          >
          <span class="ml-2 text-gray-900 dark:text-white">{{
            server.gameVersion
          }}</span>
        </div>
      </div>
      <div class="mt-6 flex gap-2">
        <button
          :disabled="actionLoading || server?.status === 'running' || isPending"
          class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          @click="doAction('start')"
        >
          Start
        </button>
        <button
          :disabled="actionLoading || server?.status !== 'running' || isPending"
          class="bg-red-500 text-white px-4 py-2 rounded-lg hover:bg-red-600 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          @click="doAction('stop')"
        >
          Stop
        </button>
        <button
          :disabled="actionLoading || isPending"
          class="bg-blue-500 text-white px-4 py-2 rounded-lg hover:bg-blue-600 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          @click="doAction('restart')"
        >
          Restart
        </button>
      </div>
    </div>
    <div v-else class="text-gray-500 dark:text-gray-400">Loading...</div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

interface Server {
  serverId: string;
  status: string;
  desiredStatus?: string;
  gamePort: number;
  engineType: string;
  gameVersion: string;
}

const route = useRoute();
const { agentId } = useCurrentAgent();
const { show: showToast } = useToast();
const { lastEvent } = useEventSource();

const serverId = route.params.serverId as string;
const server = ref<Server | null>(null);
const actionLoading = ref(false);

const { data, refresh } = await useApiFetch<Server | { body: Server }>(
  `/agent/${agentId.value}/servers/${serverId}`,
);

if (data.value && typeof data.value === "object") {
  if ("body" in data.value) {
    server.value = (data.value as any).body as Server;
  } else if ("serverId" in data.value) {
    server.value = data.value as Server;
  }
}

watch(data, (val) => {
  if (val && typeof val === "object") {
    if ("body" in val) {
      server.value = (val as any).body as Server;
    } else if ("serverId" in val) {
      server.value = val as Server;
    }
  }
});

watch(lastEvent, (evt) => {
  if (!evt || evt.type !== "server.status") return;
  if (evt.agentId !== agentId.value || evt.serverId !== serverId) return;
  if (!server.value) return;
  server.value.status = evt.newStatus || server.value.status;
  server.value.desiredStatus = evt.desiredStatus;
});

const isPending = computed(() => {
  if (!server.value) return false;
  const d = server.value.desiredStatus;
  return !!d && d !== server.value.status;
});

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

async function doAction(action: "start" | "stop" | "restart") {
  actionLoading.value = true;
  try {
    await $fetch(`/agent/${agentId.value}/servers/${serverId}/${action}`, {
      baseURL: useApiBase(),
      method: "POST",
      credentials: "include",
    });
    showToast("info", `Server ${action} requested`, {
      description: `${serverId} — operation in progress.`,
    });
    await refresh();
  } catch (err: any) {
    const status = err?.status || err?.statusCode;
    const msg =
      err?.data?.detail || err?.message || `Failed to ${action} server`;
    if (status === 409) {
      showToast("error", "Operation in progress", { description: msg });
    } else {
      showToast("error", `Server ${action} failed`, { description: msg });
    }
  } finally {
    actionLoading.value = false;
  }
}
</script>
