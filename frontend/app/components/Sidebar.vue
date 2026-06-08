<template>
  <aside
    class="w-64 bg-white dark:bg-background-dark text-gray-900 dark:text-white flex flex-col h-screen fixed left-0 top-0 transition-colors duration-200 border-r border-gray-200 dark:border-border-dark"
  >
    <div class="p-4 border-b border-gray-200 dark:border-border-dark">
      <div class="flex items-center justify-between">
        <img src="/img/MiniMin_L.avif" alt="MiniMin" class="h-10 w-auto" />
        <img
          :src="
            colorMode.value === 'dark'
              ? '/img/MiniMin_T_light.avif'
              : '/img/MiniMin_T.avif'
          "
          alt="MiniMin"
          class="h-9 w-auto"
        />
      </div>
    </div>

    <nav class="flex-1 p-4 space-y-2 flex flex-col">
      <div class="space-y-1 flex-1">
        <div class="flex items-stretch gap-1">
          <NuxtLink
            to="/agents"
            class="group flex-1 flex items-center gap-3 px-3 py-2 rounded-xl transition-colors hover:bg-gray-100 dark:hover:bg-neutral-800"
            :class="
              isAgentsActive()
                ? 'bg-gray-200 dark:bg-neutral-700 text-gray-900 dark:text-white'
                : ''
            "
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="w-5 h-5 transition-transform duration-200 group-hover:scale-110"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
              />
            </svg>
            <span>Agents</span>
          </NuxtLink>
          <button
            @click="showForm = true"
            class="group relative px-2 rounded-lg hover:bg-gray-200 dark:hover:bg-neutral-600 transition-colors text-gray-500 dark:text-neutral-400 shrink-0 flex items-center justify-center w-10"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="w-4 h-4 transition-transform duration-200 group-hover:scale-110 group-hover:rotate-90"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            <span
              class="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 px-2 py-1 text-xs text-white bg-gray-500 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none"
            >
              Register Agent
            </span>
          </button>
        </div>

        <div v-if="agents.length > 0" class="ml-4 space-y-0.5">
          <template v-for="agent in agents" :key="agent.id">
            <NuxtLink
              :to="`/agent/${agent.id}/servers`"
              class="flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm transition-colors hover:bg-gray-100 dark:hover:bg-neutral-800"
              :class="
                route.params.id === agent.id
                  ? 'bg-gray-200 dark:bg-neutral-700 text-gray-900 dark:text-white'
                  : 'text-gray-600 dark:text-neutral-400'
              "
            >
              <span
                class="w-2 h-2 rounded-full flex-shrink-0"
                :class="getAgentStatusColor(agent.id)"
              ></span>
              <span class="truncate">{{ agent.name }}</span>
            </NuxtLink>
            <div
              v-if="agent.id === agentId && agentServers.length > 0"
              class="ml-4 space-y-0.5"
            >
              <NuxtLink
                v-for="server in agentServers"
                :key="server.serverId"
                :to="`/agent/${agent.id}/servers/${server.serverId}`"
                class="flex items-center gap-2 px-3 py-1 rounded-lg text-xs transition-colors hover:bg-gray-100 dark:hover:bg-neutral-800"
                :class="
                  route.params.serverId === server.serverId
                    ? 'bg-gray-200 dark:bg-neutral-700 text-gray-900 dark:text-white'
                    : 'text-gray-500 dark:text-neutral-500'
                "
              >
                <span
                  class="w-1.5 h-1.5 rounded-full flex-shrink-0"
                  :class="getServerStatusColor(server.status)"
                ></span>
                <span class="truncate">{{ server.serverId }}</span>
              </NuxtLink>
            </div>
          </template>
        </div>
      </div>

      <div v-if="serverId && agentId" class="space-y-1 mt-4">
        <div
          class="px-3 py-1 text-xs font-semibold text-gray-500 dark:text-neutral-400 uppercase tracking-wider"
        >
          Server ({{ currentServerName }})
        </div>
        <template v-for="item in serverNav" :key="item.to">
          <NuxtLink
            v-if="!isServerNavItemDisabled(item)"
            :to="item.to"
            class="flex items-center gap-3 px-3 py-2 rounded-xl text-sm transition-colors hover:bg-gray-100 dark:hover:bg-neutral-800"
            :class="
              route.path === item.to
                ? 'bg-gray-200 dark:bg-neutral-700 text-gray-900 dark:text-white'
                : 'text-gray-600 dark:text-neutral-400'
            "
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="w-5 h-5 flex-shrink-0"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                :d="item.icon"
              />
            </svg>
            <span>{{ item.label }}</span>
            <span
              v-if="item.label === 'Overview'"
              class="w-2 h-2 rounded-full flex-shrink-0 ml-auto"
              :class="getServerStatusColor(currentServerStatus)"
            ></span>
          </NuxtLink>
          <div
            v-else
            class="flex items-center gap-3 px-3 py-2 rounded-xl text-sm text-gray-400 dark:text-neutral-600 cursor-not-allowed opacity-60 select-none"
            :title="getDisabledReason(item)"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="w-5 h-5 flex-shrink-0"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                :d="item.icon"
              />
            </svg>
            <span>{{ item.label }}</span>
          </div>
        </template>
      </div>
    </nav>

    <div
      class="p-4 border-t border-gray-200 dark:border-border-dark flex items-center justify-center gap-2"
    >
      <button
        @click="toggleColorMode"
        class="group flex items-center gap-2 px-2 py-2 rounded-xl transition-all hover:bg-gray-100 dark:hover:bg-neutral-800"
      >
        <svg
          v-if="colorMode.value === 'dark'"
          xmlns="http://www.w3.org/2000/svg"
          class="w-5 h-5 shrink-0 transition-transform duration-200 group-hover:scale-110 group-hover:rotate-12"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
          />
        </svg>
        <svg
          v-else
          xmlns="http://www.w3.org/2000/svg"
          class="w-5 h-5 shrink-0 transition-transform duration-200 group-hover:scale-110 group-hover:rotate-12"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
          />
        </svg>
        <span
          class="max-w-0 overflow-hidden opacity-0 transition-all duration-300 whitespace-nowrap group-hover:max-w-40 group-hover:opacity-100 group-hover:ml-1"
        >
          {{ colorMode.value === "dark" ? "Light Mode" : "Dark Mode" }}
        </span>
      </button>
      <button
        @click="logout"
        class="group flex items-center gap-2 px-2 py-2 rounded-xl transition-all hover:bg-red-500/10 dark:hover:bg-red-500/20 text-red-600 dark:text-red-400"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="w-5 h-5 shrink-0 transition-transform duration-200 group-hover:scale-110 group-hover:translate-x-0.5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
          />
        </svg>
        <span
          class="max-w-0 overflow-hidden opacity-0 transition-all duration-300 whitespace-nowrap group-hover:max-w-40 group-hover:opacity-100 group-hover:ml-1"
        >
          Logout
        </span>
      </button>
    </div>
    <AgentFormModal v-model="showForm" />
  </aside>
</template>

<script setup lang="ts">
const showForm = ref(false);

const colorMode = useColorMode();
const route = useRoute();
const { logout } = useAuth();

const serverId = computed(() => route.params.serverId as string | undefined);
const agentId = computed(() => route.params.id as string | undefined);

interface Server {
  serverId: string;
  status: string;
  name?: string;
}

interface ServerNavItem {
  label: string;
  to: string;
  icon: string;
  requiresRunning?: boolean;
}

const { servers: agentServers, refresh: refreshServers } = useServers(agentId);

const currentServerStatus = ref("");
const serverFilesInitialized = ref(true);

watch(
  [agentServers, serverId],
  () => {
    if (!serverId.value) {
      currentServerStatus.value = "";
      return;
    }
    const server = agentServers.value.find(
      (s: Server) => s.serverId === serverId.value,
    );
    currentServerStatus.value = server?.status ?? "";
  },
  { immediate: true },
);

const currentServerName = computed(() => {
  if (!serverId.value) return "";
  const server = agentServers.value.find(
    (s: Server) => s.serverId === serverId.value,
  );
  return server?.name || server?.serverId || "";
});

const { lastEvent } = useEventSource();

watch(lastEvent, (evt) => {
  if (!evt) return;

  if (evt.type === "server.status") {
    if (evt.agentId !== agentId.value) return;

    if (evt.serverId === serverId.value && evt.newStatus) {
      currentServerStatus.value = evt.newStatus;
    }

    const server = agentServers.value.find((s) => s.serverId === evt.serverId);
    if (server) {
      server.status = evt.newStatus || server.status;
    }
  }
});

watch(
  [agentId, serverId],
  async ([newAgentId, newServerId]) => {
    if (!newAgentId || !newServerId) {
      serverFilesInitialized.value = true;
      return;
    }

    try {
      const res = await $fetch<
        { initialized?: boolean } | { body?: { initialized?: boolean } }
      >(`/agent/${newAgentId}/servers/${newServerId}/config`, {
        baseURL: useApiBase(),
        credentials: "include",
      });

      const initializedRoot = (res as { initialized?: boolean })?.initialized;
      const initializedBody = (res as { body?: { initialized?: boolean } })
        ?.body?.initialized;
      const initialized: boolean = initializedRoot ?? initializedBody ?? true;

      serverFilesInitialized.value = initialized;
    } catch (err: any) {
      serverFilesInitialized.value = true;
    }
  },
  { immediate: true },
);

function getServerStatusColor(status: string) {
  switch (status) {
    case "running":
      return "bg-green-500";
    case "exited":
      return "bg-red-500";
    default:
      return "bg-gray-400 dark:bg-neutral-500";
  }
}

const serverNav = computed(() => {
  if (!agentId.value || !serverId.value) return [];
  const base = `/agent/${agentId.value}/servers/${serverId.value}`;
  return [
    {
      label: "Overview",
      to: base,
      icon: "M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6",
    },
    {
      label: "Mods",
      to: `${base}/mods`,
      icon: "M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10",
    },
    {
      label: "Files",
      to: `${base}/files`,
      icon: "M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z",
    },
    {
      label: "Logs",
      to: `${base}/logs`,
      icon: "M4 6h16M4 12h16M4 18h16",
    },
    {
      label: "Console",
      to: `${base}/console`,
      icon: "M8 9l3 3-3 3m5 0h3",
      requiresRunning: true,
    },
    {
      label: "Players",
      to: `${base}/players`,
      icon: "M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z",
      requiresRunning: true,
    },
  ] as ServerNavItem[];
});

function isServerNavItemDisabled(item: ServerNavItem) {
  if (item.label !== "Overview" && !serverFilesInitialized.value) {
    return true;
  }
  if (item.requiresRunning && currentServerStatus.value !== "running") {
    return true;
  }
  return false;
}

function getDisabledReason(item: ServerNavItem) {
  if (item.label !== "Overview" && !serverFilesInitialized.value) {
    return `${item.label} is available only after server files are initialized`;
  }
  if (item.requiresRunning && currentServerStatus.value !== "running") {
    return `${item.label} is available only when the server is running`;
  }
  return `${item.label} is currently unavailable`;
}

function toggleColorMode() {
  colorMode.preference = colorMode.value === "dark" ? "light" : "dark";
}

const { data: agentsData } = await useFetch("/agents", {
  baseURL: useApiBase(),
  credentials: "include",
  key: "agents",
});

const agents = computed(() => {
  const val = agentsData.value;
  if (Array.isArray(val)) return val as any[];
  if (val && typeof val === "object" && "body" in val) {
    return (val as any).body || [];
  }
  return [];
});

function isAgentsActive() {
  return route.path === "/agents" || route.path.startsWith("/agents/");
}

// Agent statuses via SSE
const agentStatuses = ref<Record<string, boolean>>({});

onMounted(async () => {
  try {
    const statuses = await $fetch<Array<{ id: string; online: boolean }>>(
      "/agents/status",
      { baseURL: useApiBase(), credentials: "include" },
    );
    for (const s of statuses) {
      agentStatuses.value[s.id] = s.online;
    }
  } catch (e) {
    console.error("Failed to load initial agent statuses:", e);
  }
});

watch(lastEvent, (evt) => {
  if (!evt || evt.type !== "agent.status") return;
  const isOnline = evt.newStatus === "online";
  if (evt.agentId) {
    agentStatuses.value[evt.agentId] = isOnline;
  }
});

function getAgentStatusColor(agentId: string) {
  const isOnline = agentStatuses.value[agentId];
  if (isOnline === true) return "bg-green-500";
  if (isOnline === false) return "bg-red-500";
  return "bg-gray-400 dark:bg-neutral-600";
}
</script>
