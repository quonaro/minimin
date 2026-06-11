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
        <NuxtLink
          to="/servers"
          class="group flex items-center gap-3 px-3 py-2 rounded-xl transition-colors hover:bg-gray-100 dark:hover:bg-neutral-800"
          :class="
            route.path === '/servers' || route.path.startsWith('/servers/')
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
              d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"
            />
          </svg>
          <span>Servers</span>
        </NuxtLink>

        <div v-if="servers.length > 0" class="ml-4 space-y-0.5">
          <NuxtLink
            v-for="server in servers"
            :key="server.serverId"
            :to="`/servers/${server.serverId}`"
            class="flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm transition-colors hover:bg-gray-100 dark:hover:bg-neutral-800"
            :class="
              route.params.serverId === server.serverId
                ? 'bg-gray-200 dark:bg-neutral-700 text-gray-900 dark:text-white'
                : 'text-gray-600 dark:text-neutral-400'
            "
          >
            <span
              class="w-2 h-2 rounded-full flex-shrink-0"
              :class="getServerStatusColor(server.serverStatus)"
            ></span>
            <span class="truncate">{{ server.serverId }}</span>
          </NuxtLink>
        </div>
      </div>

      <div v-if="serverId" class="space-y-1 mt-4">
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
  </aside>
</template>

<script setup lang="ts">
import type { Server } from "~/composables/useServers";

const colorMode = useColorMode();
const route = useRoute();
const { logout } = useAuth();

const serverId = computed(() => route.params.serverId as string | undefined);

interface ServerNavItem {
  label: string;
  to: string;
  icon: string;
  requiresRunning?: boolean;
  requiresContainerRunning?: boolean;
  requiresFilesInitialized?: boolean;
}

const { servers, refresh: refreshServers } = useServers();

const currentServerStatus = ref("");
const currentContainerStatus = ref("");
const serverFilesInitialized = ref(true);

watch(
  [servers, serverId],
  () => {
    if (!serverId.value) {
      currentServerStatus.value = "";
      return;
    }
    const server = servers.value.find(
      (s: Server) => s.serverId === serverId.value,
    );
    currentServerStatus.value = server?.serverStatus ?? "";
    currentContainerStatus.value = server?.containerStatus ?? "";
  },
  { immediate: true },
);

const currentServerName = computed(() => {
  if (!serverId.value) return "";
  const server = servers.value.find(
    (s: Server) => s.serverId === serverId.value,
  );
  return server?.name || server?.serverId || "";
});

const currentServerEngineType = computed(() => {
  if (!serverId.value) return "";
  const server = servers.value.find(
    (s: Server) => s.serverId === serverId.value,
  );
  return server?.engineType || "";
});

watch(
  serverId,
  async (newServerId) => {
    if (!newServerId) {
      serverFilesInitialized.value = true;
      return;
    }

    try {
      const res = await $fetch<
        { initialized?: boolean } | { body?: { initialized?: boolean } }
      >(`/servers/${newServerId}/config`, {
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
    case "starting":
      return "bg-yellow-500";
    case "exited":
      return "bg-red-500";
    default:
      return "bg-gray-400 dark:bg-neutral-500";
  }
}

function getServerLink(targetServerId: string) {
  if (!serverId.value) {
    return `/servers/${targetServerId}`;
  }
  return route.path.replace(
    `/servers/${serverId.value}`,
    `/servers/${targetServerId}`,
  );
}

const serverNav = computed(() => {
  if (!serverId.value) return [];
  const base = `/servers/${serverId.value}`;
  const engine = currentServerEngineType.value.toUpperCase();
  const items: ServerNavItem[] = [
    {
      label: "Overview",
      to: base,
      icon: "M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6",
    },
  ];
  if (engine === "FABRIC" || engine === "FORGE") {
    items.push({
      label: "Resources",
      to: `${base}/mods`,
      icon: "M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10",
      requiresFilesInitialized: true,
    });
  }
  if (engine === "PAPERMC" || engine === "PAPER") {
    items.push({
      label: "Plugins",
      to: `${base}/plugins`,
      icon: "M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10",
      requiresFilesInitialized: true,
    });
  }
  items.push(
    {
      label: "Files",
      to: `${base}/files`,
      icon: "M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z",
      requiresFilesInitialized: true,
    },
    {
      label: "Logs",
      to: `${base}/logs`,
      icon: "M4 6h16M4 12h16M4 18h16",
      requiresContainerRunning: true,
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
  );
  return items;
});

function isServerNavItemDisabled(item: ServerNavItem) {
  if (item.requiresFilesInitialized && !serverFilesInitialized.value) {
    return true;
  }
  if (
    item.requiresContainerRunning &&
    currentContainerStatus.value !== "running"
  ) {
    return true;
  }
  if (item.requiresRunning && currentServerStatus.value !== "running") {
    return true;
  }
  return false;
}

function getDisabledReason(item: ServerNavItem) {
  if (item.requiresFilesInitialized && !serverFilesInitialized.value) {
    return `${item.label} is available only after server files are initialized`;
  }
  if (
    item.requiresContainerRunning &&
    currentContainerStatus.value !== "running"
  ) {
    return `${item.label} is available only when the container is running`;
  }
  if (item.requiresRunning && currentServerStatus.value !== "running") {
    return `${item.label} is available only when the server is running`;
  }
  return `${item.label} is currently unavailable`;
}

function toggleColorMode() {
  colorMode.preference = colorMode.value === "dark" ? "light" : "dark";
}

onMounted(() => {
  const interval = setInterval(() => {
    refreshServers();
  }, 5000);
  onBeforeUnmount(() => clearInterval(interval));
});
</script>
