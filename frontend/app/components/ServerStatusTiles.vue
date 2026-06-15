<script setup lang="ts">
import {
  Activity,
  Check,
  CloudDownload,
  Globe,
  Pencil,
  Terminal,
  X as XIcon,
} from "lucide-vue-next";
import type { Server } from "~/composables/useServers";

const props = defineProps<{
  server: Server;
  containerUptime: string;
  serverUptime: string;
}>();

const serverId = computed(() => props.server.serverId);

const serverStatus = computed(() => props.server.serverStatus);
const { percentage: pullPercentage } = useImagePullProgress(
  serverId,
  serverStatus,
);

const {
  editingPort,
  tempPort,
  portLoading,
  savePort,
  editingPublicRcon,
  tempPublicRcon,
  tempRconPort,
  rconLoading,
  savePublicRcon,
} = useServerConfigEdits(serverId.value, toRef(props, "server"));

function formatServerStatus(status: string): string {
  switch (status) {
    case "pulling_image":
      return "Pulling image...";
    case "starting":
      return "Starting...";
    case "running":
      return "Running";
    case "stopped":
      return "Stopped";
    default:
      return status;
  }
}

function formatContainerStatus(status: string): string {
  switch (status) {
    case "running":
      return "Running";
    case "starting":
      return "Starting...";
    case "exited":
      return "Exited";
    default:
      return status || "Created";
  }
}
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-6">
    <!-- Server Status -->
    <div class="space-y-2">
      <p
        class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
      >
        Server Status
      </p>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <!-- Container -->
        <div
          class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
        >
          <div
            :class="[
              'w-9 h-9 shrink-0 rounded-lg flex items-center justify-center',
              server.containerStatus === 'running'
                ? 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400'
                : server.containerStatus === 'starting'
                  ? 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 dark:text-yellow-400'
                  : server.containerStatus === 'exited'
                    ? 'bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400'
                    : 'bg-gray-100 dark:bg-neutral-600/50 text-gray-600 dark:text-gray-400',
            ]"
          >
            <Activity
              v-if="server.containerStatus === 'running'"
              class="w-4 h-4 animate-pulse-icon"
            />
            <Activity v-else class="w-4 h-4" />
          </div>
          <div class="min-w-0">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              Container
            </p>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ formatContainerStatus(server.containerStatus) }}
              <span
                v-if="containerUptime"
                class="text-gray-500 dark:text-neutral-400 font-normal"
              >
                &middot; {{ containerUptime }}
              </span>
            </p>
          </div>
        </div>
        <!-- Server -->
        <div
          class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
        >
          <div
            :class="[
              'w-9 h-9 shrink-0 rounded-lg flex items-center justify-center',
              server.serverStatus === 'running'
                ? 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400'
                : server.serverStatus === 'starting'
                  ? 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 dark:text-yellow-400'
                  : server.serverStatus === 'pulling_image'
                    ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400'
                    : 'bg-gray-100 dark:bg-neutral-600/50 text-gray-600 dark:text-gray-400',
            ]"
          >
            <Activity
              v-if="server.serverStatus === 'running'"
              class="w-4 h-4 animate-pulse-icon"
            />
            <CloudDownload
              v-else-if="server.serverStatus === 'pulling_image'"
              class="w-4 h-4 animate-bounce"
            />
            <Activity v-else class="w-4 h-4" />
          </div>
          <div class="min-w-0">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              Server
            </p>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ formatServerStatus(server.serverStatus) }}
              <span
                v-if="serverUptime"
                class="text-gray-500 dark:text-neutral-400 font-normal"
              >
                &middot; {{ serverUptime }}
              </span>
            </p>
            <div v-if="server.serverStatus === 'pulling_image'" class="mt-1.5">
              <div
                class="w-full bg-gray-100 dark:bg-neutral-700 rounded-full h-1.5"
              >
                <div
                  class="h-1.5 rounded-full bg-blue-500 transition-all duration-300"
                  :style="{ width: pullPercentage + '%' }"
                />
              </div>
              <p class="text-[10px] text-gray-400 dark:text-neutral-500 mt-0.5">
                {{ pullPercentage }}%
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Network -->
    <div class="space-y-2">
      <p
        class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
      >
        Network
      </p>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <!-- Game Port -->
        <div
          class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400"
          >
            <Globe class="w-4 h-4" />
          </div>
          <div class="min-w-0 flex-1">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              Game Port (host)
            </p>
            <div v-if="!editingPort" class="flex items-center gap-2">
              <p class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ server.gamePort }}
              </p>
              <button
                v-if="server.containerStatus !== 'running'"
                class="text-gray-400 hover:text-primary transition-colors"
                :disabled="portLoading"
                @click="
                  tempPort = server.gamePort;
                  editingPort = true;
                "
              >
                <Pencil class="w-3.5 h-3.5" />
              </button>
            </div>
            <div v-else class="flex items-center gap-2">
              <number-input
                v-model="tempPort"
                :min="1024"
                :max="65535"
                size="sm"
                class="w-24"
                @keyup.enter="savePort"
              />
              <button
                class="text-green-500 hover:text-green-600 transition-colors"
                :disabled="portLoading"
                @click="savePort"
              >
                <Check class="w-4 h-4" />
              </button>
              <button
                class="text-red-500 hover:text-red-600 transition-colors"
                :disabled="portLoading"
                @click="editingPort = false"
              >
                <XIcon class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <!-- Public RCON -->
        <div
          class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-600 dark:text-emerald-400"
          >
            <Terminal class="w-4 h-4" />
          </div>
          <div class="min-w-0 flex-1">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              Public RCON
            </p>
            <div v-if="!editingPublicRcon" class="flex items-center gap-2">
              <p class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ server.publicRcon ? `Yes (${server.rconPort})` : "No" }}
              </p>
              <button
                v-if="server.containerStatus !== 'running'"
                class="text-gray-400 hover:text-primary transition-colors"
                :disabled="rconLoading"
                @click="
                  tempPublicRcon = server.publicRcon;
                  tempRconPort = server.rconPort || server.gamePort + 10;
                  editingPublicRcon = true;
                "
              >
                <Pencil class="w-3.5 h-3.5" />
              </button>
            </div>
            <div v-else class="flex items-center gap-2">
              <select
                v-model="tempPublicRcon"
                class="text-sm bg-white dark:bg-neutral-700 border border-gray-200 dark:border-neutral-600 rounded-lg px-2 py-1 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:outline-none"
              >
                <option :value="false">No</option>
                <option :value="true">Yes</option>
              </select>
              <number-input
                v-if="tempPublicRcon"
                v-model="tempRconPort"
                :min="1024"
                :max="65535"
                size="sm"
                class="w-24"
                @keyup.enter="savePublicRcon"
              />
              <button
                class="text-green-500 hover:text-green-600 transition-colors"
                :disabled="rconLoading"
                @click="savePublicRcon"
              >
                <Check class="w-4 h-4" />
              </button>
              <button
                class="text-red-500 hover:text-red-600 transition-colors"
                :disabled="rconLoading"
                @click="editingPublicRcon = false"
              >
                <XIcon class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
