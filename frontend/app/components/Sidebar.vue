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
            class="group flex-1 flex items-center gap-3 px-3 py-2 rounded-xl transition-colors hover:bg-gray-100 dark:hover:bg-gray-800"
            :class="
              isAgentsActive()
                ? 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white'
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
            class="group relative px-2 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors text-gray-500 dark:text-gray-400 shrink-0 flex items-center justify-center w-10"
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
          <NuxtLink
            v-for="agent in agents"
            :key="agent.id"
            :to="`/agent/${agent.id}/servers`"
            class="flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm transition-colors hover:bg-gray-100 dark:hover:bg-gray-800"
            :class="
              route.params.id === agent.id
                ? 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white'
                : 'text-gray-600 dark:text-gray-400'
            "
          >
            <span
              class="w-2 h-2 rounded-full flex-shrink-0"
              :class="getAgentStatusColor(agent.id)"
            ></span>
            <span class="truncate">{{ agent.name }}</span>
          </NuxtLink>
        </div>
      </div>

      <div v-if="serverId && agentId" class="space-y-1 mt-4">
        <div
          class="px-3 py-1 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
          Server
        </div>
        <NuxtLink
          v-for="item in serverNav"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-3 px-3 py-2 rounded-xl text-sm transition-colors hover:bg-gray-100 dark:hover:bg-gray-800"
          :class="
            route.path === item.to
              ? 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white'
              : 'text-gray-600 dark:text-gray-400'
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
        </NuxtLink>
      </div>

      <div
        v-if="nextCheckAt"
        class="mt-auto pt-2 text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1.5"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="w-3.5 h-3.5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>Next check in {{ countdown }}</span>
      </div>
    </nav>

    <div
      class="p-4 border-t border-gray-200 dark:border-border-dark flex items-center justify-center gap-2"
    >
      <button
        @click="toggleColorMode"
        class="group flex items-center gap-2 px-2 py-2 rounded-xl transition-all hover:bg-gray-100 dark:hover:bg-gray-800"
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
      label: "Settings",
      to: `${base}/settings`,
      icon: "M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z",
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
    },
    {
      label: "Players",
      to: `${base}/players`,
      icon: "M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z",
    },
  ];
});

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

// WebSocket connection for agent statuses
const agentStatuses = ref<Record<string, boolean>>({});
const nextCheckAt = ref<Date | null>(null);
const countdown = ref<string>("");

let countdownTimer: ReturnType<typeof setInterval> | null = null;

function updateCountdown() {
  if (!nextCheckAt.value) {
    countdown.value = "";
    return;
  }
  const diff = Math.max(
    0,
    Math.ceil((nextCheckAt.value.getTime() - Date.now()) / 1000),
  );
  const m = Math.floor(diff / 60);
  const s = diff % 60;
  countdown.value = m > 0 ? `${m}:${s.toString().padStart(2, "0")}` : `${s}s`;
}

onMounted(() => {
  const wsUrl = "ws://localhost:8081/ws/agents";
  const ws = new WebSocket(wsUrl);

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      if (Array.isArray(data)) {
        // Legacy format fallback
        const newStatuses: Record<string, boolean> = {};
        for (const status of data) {
          newStatuses[status.id] = status.online;
        }
        agentStatuses.value = newStatuses;
        return;
      }
      const statuses = data.statuses as Array<{ id: string; online: boolean }>;
      const newStatuses: Record<string, boolean> = {};
      for (const status of statuses) {
        newStatuses[status.id] = status.online;
      }
      agentStatuses.value = newStatuses;
      if (data.next_check_at) {
        nextCheckAt.value = new Date(data.next_check_at as string);
        updateCountdown();
      }
    } catch (e) {
      console.error("Failed to parse agent statuses:", e);
    }
  };

  ws.onerror = (error) => {
    console.error("WebSocket error:", error);
  };

  ws.onclose = () => {
    console.log("WebSocket closed, reconnecting in 5s...");
    setTimeout(() => {
      // Simple reconnection - component will remount on navigation
    }, 5000);
  };

  countdownTimer = setInterval(updateCountdown, 1000);

  onUnmounted(() => {
    ws.close();
    if (countdownTimer) {
      clearInterval(countdownTimer);
    }
  });
});

function getAgentStatusColor(agentId: string) {
  const isOnline = agentStatuses.value[agentId];
  if (isOnline === true) return "bg-green-500";
  if (isOnline === false) return "bg-red-500";
  return "bg-gray-400 dark:bg-gray-600";
}
</script>
