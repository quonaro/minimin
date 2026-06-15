<template>
  <div class="p-6">
    <div v-if="server" class="space-y-6">
      <!-- Main server card -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl p-6 md:p-8 shadow-sm"
      >
        <div class="flex flex-col md:flex-row gap-8">
          <server-header
            :server="server"
            :server-id="serverId"
            :icon-url="iconUrl"
            :icon-error="iconError"
            :current-action="currentAction"
            :is-pending="isPending"
            :remove-before-start="removeBeforeStart"
            :delete-loading="deleteLoading"
            :recreate-loading="recreateLoading"
            v-model:show-icon-editor="showIconEditor"
            v-model:selected-icon-file="selectedIconFile"
            @update:icon-error="iconError = $event"
            @icon-select="onIconFileSelect"
            @icon-save="uploadProcessedIcon"
            @update:remove-before-start="removeBeforeStart = $event"
            @action="doAction($event)"
            @delete="promptDelete"
            @recreate="promptRecreate"
          />

          <div class="flex-1 min-w-0">
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-4">
              {{ server.serverId }}
            </h1>
            <server-status-tiles
              :server="server"
              :container-uptime="containerUptime"
              :server-uptime="serverUptime"
            />
            <server-config-tiles :server="server" />
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <server-storage-tiles
                :server-id="serverId"
                :host-path="server.hostPath"
                :server-disk="serverDisk"
              />
              <server-metrics-panel
                :server-id="serverId"
                :container-status="server.containerStatus"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Crash Reports Alert -->
      <div
        v-if="hasCrashReports"
        class="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-2xl p-4 flex items-start gap-3"
      >
        <AlertTriangle
          class="w-5 h-5 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5"
        />
        <div class="flex-1">
          <p class="text-sm font-semibold text-amber-800 dark:text-amber-300">
            {{ crashReports.length }} crash report{{
              crashReports.length === 1 ? "" : "s"
            }}
            found
          </p>
          <p class="text-xs text-amber-700 dark:text-amber-400 mt-0.5">
            Latest: {{ formatTimestamp(latestCrashReportDate || undefined) }}
          </p>
        </div>
        <NuxtLink
          :to="`/servers/${serverId}/crash-reports`"
          class="text-xs font-medium text-amber-700 dark:text-amber-400 hover:underline flex-shrink-0"
        >
          View →
        </NuxtLink>
      </div>

      <!-- Players Online -->
      <div
        v-if="
          server?.serverStatus === 'running' && onlinePlayers.players.length > 0
        "
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl p-4 md:p-6 shadow-sm"
      >
        <div class="flex items-center justify-between mb-4">
          <h2
            class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2"
          >
            <Users class="w-5 h-5 text-primary" />
            Players Online
          </h2>
          <span class="text-sm text-gray-500 dark:text-neutral-400">
            {{ onlinePlayers.players.length }} / {{ onlinePlayers.max }}
          </span>
        </div>
        <div class="flex flex-wrap gap-3">
          <div
            v-for="name in onlinePlayers.players"
            :key="name"
            class="flex items-center gap-2 px-3 py-2 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
          >
            <img
              :src="`https://mc-heads.net/avatar/${name}/32`"
              alt=""
              class="w-6 h-6 rounded"
              loading="lazy"
            />
            <span class="text-sm font-medium text-gray-900 dark:text-white">{{
              name
            }}</span>
          </div>
        </div>
      </div>

      <!-- Logs -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden"
      >
        <button
          class="w-full flex items-center justify-between p-4 md:p-6 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
          @click="toggleLogs"
        >
          <div class="flex items-center gap-3">
            <Terminal class="w-5 h-5 text-gray-500 dark:text-neutral-400" />
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">
              Logs
            </h2>
          </div>
          <ChevronDown
            class="w-5 h-5 text-gray-500 dark:text-neutral-400 transition-transform"
            :class="{ 'rotate-180': logsExpanded }"
          />
        </button>
        <div
          v-show="logsExpanded"
          class="px-4 pt-4 pb-4 md:px-6 md:pt-6 md:pb-6 h-96"
        >
          <server-logs
            ref="serverLogsRef"
            :server-id="serverId"
            class="h-full"
          />
        </div>
      </div>

      <!-- Server Properties -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden"
      >
        <button
          class="w-full flex items-center justify-between p-4 md:p-6 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
          @click="propertiesExpanded = !propertiesExpanded"
        >
          <div class="flex items-center gap-3">
            <Pencil class="w-5 h-5 text-gray-500 dark:text-neutral-400" />
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">
              Server Properties
            </h2>
          </div>
          <ChevronDown
            class="w-5 h-5 text-gray-500 dark:text-neutral-400 transition-transform"
            :class="{ 'rotate-180': propertiesExpanded }"
          />
        </button>
        <div
          v-show="propertiesExpanded"
          class="px-4 pt-4 pb-4 md:px-6 md:pt-6 md:pb-6"
        >
          <server-properties :server-id="serverId" />
        </div>
      </div>
    </div>

    <div v-else class="text-gray-500 dark:text-neutral-400">Loading...</div>

    <confirm-dialog
      v-if="server"
      v-model="showDeleteDialog"
      title="Delete Server"
      :description="`Delete server &quot;${serverId}&quot;? The container will be removed.`"
      confirm-label="Delete"
      danger
      show-wipe
      @confirm="onDeleteConfirmed"
    />
    <confirm-dialog
      v-if="server"
      v-model="showRecreateDialog"
      title="Recreate World"
      :description="`This will reset the world for server &quot;${serverId}&quot;. All world data will be lost.`"
      confirm-label="Recreate"
      danger
      @confirm="onRecreateConfirmed"
    />
  </div>
</template>

<script setup lang="ts">
import {
  AlertTriangle,
  ChevronDown,
  Pencil,
  Terminal,
  Users,
} from "lucide-vue-next";
import { nextTick, watch } from "vue";
import type { Server } from "~/composables/useServers";
import { useServerEvents } from "~/composables/useServerEvents";

definePageMeta({ middleware: "auth" });

const route = useRoute();
const serverId = route.params.serverId as string;
const { servers, refresh: refreshServers } = useServers();
const { playersMap } = useServerEvents();
const { disk: serverDisk } = useServerDisk(serverId);

const server = computed<Server | null>(
  () => servers.value.find((s: Server) => s.serverId === serverId) ?? null,
);
usePageTitle(() => server.value?.serverId || serverId);

const iconTimestamp = ref(Date.now());
const iconError = ref(false);
const showIconEditor = ref(false);
const selectedIconFile = ref<File | null>(null);
const logsExpanded = ref(false);
const propertiesExpanded = ref(false);

const serverLogsRef = ref<{
  scrollToBottom: () => void;
  reconnect: (tail?: number) => void;
} | null>(null);

const iconUrl = computed(() => {
  if (!server.value) return "";
  return `${useApiBase()}/servers/${serverId}/icon?t=${iconTimestamp.value}`;
});

const containerStartedAt = computed(() => {
  if (
    !server.value ||
    server.value.containerStatus !== "running" ||
    !server.value.containerStartedAt
  ) {
    return undefined;
  }
  return server.value.containerStartedAt;
});

const serverStartedAt = computed(() => {
  if (
    !server.value ||
    server.value.serverStatus !== "running" ||
    !server.value.serverStartedAt
  ) {
    return undefined;
  }
  return server.value.serverStartedAt;
});

const containerUptime = useUptime(containerStartedAt);
const serverUptime = useUptime(serverStartedAt);

function onIconFileSelect(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0] ?? null;
  if (!file) return;
  selectedIconFile.value = file;
  showIconEditor.value = true;
  input.value = "";
}

async function uploadProcessedIcon(blob: Blob) {
  const formData = new FormData();
  formData.append("icon", blob, "icon.png");
  try {
    await $fetch(`/servers/${serverId}/icon`, {
      baseURL: useApiBase(),
      method: "POST",
      body: formData,
      credentials: "include",
    });
    iconTimestamp.value = Date.now();
    iconError.value = false;
  } catch (err: any) {
    const msg = err?.data?.detail || err?.message || "Failed to upload icon";
    const { show } = useToast();
    show("error", "Upload failed", { description: msg });
  }
}

function toggleLogs() {
  logsExpanded.value = !logsExpanded.value;
  if (logsExpanded.value && serverLogsRef.value) {
    nextTick(() => serverLogsRef.value!.scrollToBottom());
  }
}

function formatTimestamp(iso: string | undefined): string {
  if (!iso) return "";
  return new Date(iso).toLocaleString();
}

const onlinePlayers = computed(() => {
  return playersMap.value[serverId] ?? { players: [], max: 0 };
});

const {
  actionLoading,
  currentAction,
  removeBeforeStart,
  isPending,
  deleteLoading,
  recreateLoading,
  showDeleteDialog,
  showRecreateDialog,
  doAction,
  promptDelete,
  onDeleteConfirmed,
  promptRecreate,
  onRecreateConfirmed,
} = useServerActions(serverId, server);

const { crashReports, hasCrashReports, latestCrashReportDate } =
  useCrashReports(serverId);
</script>
