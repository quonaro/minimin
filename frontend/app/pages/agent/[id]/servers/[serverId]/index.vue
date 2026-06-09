<template>
  <div class="p-6">
    <div v-if="server" class="space-y-6">
      <!-- Main server card -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl p-6 md:p-8 shadow-sm"
      >
        <div class="flex flex-col md:flex-row gap-8">
          <!-- Avatar column -->
          <div class="flex flex-col items-center gap-4 shrink-0">
            <div class="relative">
              <div
                class="w-24 h-24 rounded-2xl overflow-hidden shadow-lg bg-gradient-to-br from-gray-200 to-gray-300 dark:from-neutral-700 dark:to-neutral-600 flex items-center justify-center ring-4 ring-white dark:ring-neutral-800"
              >
                <img
                  v-if="iconUrl && !iconError"
                  :src="iconUrl"
                  alt="Server icon"
                  class="w-full h-full object-cover"
                  @error="iconError = true"
                />
                <ServerIcon
                  v-else
                  class="w-10 h-10 text-indigo-500 dark:text-indigo-400"
                />
              </div>
              <span
                :class="[
                  'absolute -bottom-1 -right-1 w-5 h-5 rounded-full border-2 border-white dark:border-neutral-800',
                  server.status === 'running'
                    ? 'bg-green-500 animate-pulse'
                    : server.status === 'exited'
                      ? 'bg-red-500'
                      : 'bg-gray-400',
                ]"
              />
            </div>
            <button
              class="text-xs font-medium text-gray-500 dark:text-neutral-400 hover:text-primary transition-colors"
              @click="fileInput?.click()"
            >
              Change Icon
            </button>
            <input
              ref="fileInput"
              type="file"
              accept="image/png,image/jpeg,image/jpg,image/gif,image/webp,image/bmp"
              class="hidden"
              @change="onIconFileSelect"
            />
            <server-icon-editor
              v-model="showIconEditor"
              :file="selectedIconFile"
              @save="uploadProcessedIcon"
            />
          </div>

          <!-- Details column -->
          <div class="flex-1 min-w-0">
            <!-- Title row -->
            <div class="mb-6">
              <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
                {{ server.serverId }}
              </h1>
              <div class="flex items-center gap-3 flex-wrap">
                <span
                  :class="[
                    getStatusColor(server.status),
                    'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium',
                    server.status === 'running' &&
                      'animate-heartbeat dark:animate-heartbeat-dark',
                  ]"
                >
                  <Activity
                    v-if="server.status === 'running'"
                    :class="server.status === 'running' && 'animate-pulse-icon'"
                    class="w-3.5 h-3.5"
                  />
                  {{ server.status }}
                </span>
                <span
                  v-if="reactiveUptime"
                  class="text-sm text-gray-500 dark:text-neutral-400 flex items-center gap-1.5"
                >
                  <Clock
                    class="w-3.5 h-3.5 text-amber-500 dark:text-amber-400"
                  />
                  Uptime: {{ reactiveUptime }}
                </span>
                <span
                  v-if="isPending"
                  class="text-xs text-gray-500 dark:text-neutral-400 italic"
                >
                  ({{ server.desiredStatus }}…)
                </span>
              </div>
            </div>

            <!-- Info tiles -->
            <div
              class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 mb-6"
            >
              <!-- Agent ID -->
              <div
                class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
              >
                <div
                  class="w-9 h-9 shrink-0 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center text-purple-600 dark:text-purple-400"
                >
                  <Hash class="w-4 h-4" />
                </div>
                <div class="min-w-0">
                  <p
                    class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                  >
                    Agent ID
                  </p>
                  <p
                    class="text-sm font-semibold text-gray-900 dark:text-white truncate font-mono"
                  >
                    {{ agentId }}
                  </p>
                </div>
              </div>

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
                    Game Port
                  </p>
                  <div v-if="!editingPort" class="flex items-center gap-2">
                    <p
                      class="text-sm font-semibold text-gray-900 dark:text-white"
                    >
                      {{ server.gamePort }}
                    </p>
                    <button
                      v-if="server.status !== 'running'"
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
                <div class="min-w-0">
                  <p
                    class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                  >
                    Public RCON
                  </p>
                  <p
                    class="text-sm font-semibold text-gray-900 dark:text-white"
                  >
                    {{ server.publicRcon ? `Yes (${server.rconPort})` : "No" }}
                  </p>
                </div>
              </div>

              <!-- Restart Policy -->
              <div
                class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
              >
                <div
                  class="w-9 h-9 shrink-0 rounded-lg bg-teal-100 dark:bg-teal-900/30 flex items-center justify-center text-teal-600 dark:text-teal-400"
                >
                  <RefreshCw class="w-4 h-4" />
                </div>
                <div class="min-w-0 flex-1">
                  <p
                    class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                  >
                    Restart Policy
                  </p>
                  <div
                    v-if="!editingRestartPolicy"
                    class="flex items-center gap-2"
                  >
                    <p
                      class="text-sm font-semibold text-gray-900 dark:text-white"
                    >
                      {{ server.restartPolicy || "no" }}
                    </p>
                    <button
                      v-if="server.status !== 'running'"
                      class="text-gray-400 hover:text-primary transition-colors"
                      :disabled="restartPolicyLoading"
                      @click="
                        tempRestartPolicy = server.restartPolicy || 'no';
                        editingRestartPolicy = true;
                      "
                    >
                      <Pencil class="w-3.5 h-3.5" />
                    </button>
                  </div>
                  <div v-else class="flex items-center gap-2">
                    <select
                      v-model="tempRestartPolicy"
                      class="text-sm bg-white dark:bg-neutral-700 border border-gray-200 dark:border-neutral-600 rounded-lg px-2 py-1 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:outline-none"
                    >
                      <option value="no">no</option>
                      <option value="always">always</option>
                      <option value="unless-stopped">unless-stopped</option>
                      <option value="on-failure">on-failure</option>
                    </select>
                    <button
                      class="text-green-500 hover:text-green-600 transition-colors"
                      :disabled="restartPolicyLoading"
                      @click="saveRestartPolicy"
                    >
                      <Check class="w-4 h-4" />
                    </button>
                    <button
                      class="text-red-500 hover:text-red-600 transition-colors"
                      :disabled="restartPolicyLoading"
                      @click="editingRestartPolicy = false"
                    >
                      <XIcon class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>

              <!-- Engine -->
              <div
                class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
              >
                <div
                  class="w-9 h-9 shrink-0 rounded-lg bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center text-orange-600 dark:text-orange-400"
                >
                  <Box class="w-4 h-4" />
                </div>
                <div class="min-w-0">
                  <p
                    class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                  >
                    Engine
                  </p>
                  <p
                    class="text-sm font-semibold text-gray-900 dark:text-white"
                  >
                    {{ server.engineType }}
                  </p>
                </div>
              </div>

              <!-- Version -->
              <div
                class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
              >
                <div
                  class="w-9 h-9 shrink-0 rounded-lg bg-pink-100 dark:bg-pink-900/30 flex items-center justify-center text-pink-600 dark:text-pink-400"
                >
                  <Tag class="w-4 h-4" />
                </div>
                <div class="min-w-0">
                  <p
                    class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                  >
                    Version
                  </p>
                  <p
                    class="text-sm font-semibold text-gray-900 dark:text-white"
                  >
                    {{ server.gameVersion }}
                  </p>
                </div>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex flex-wrap gap-3">
              <button
                :disabled="
                  actionLoading || server?.status === 'running' || isPending
                "
                class="inline-flex items-center gap-2 bg-emerald-500 hover:bg-emerald-600 text-white px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed shadow-sm hover:shadow-md active:scale-95"
                @click="doAction('start')"
              >
                <Play class="w-4 h-4" />
                Start
              </button>
              <button
                :disabled="
                  actionLoading || server?.status !== 'running' || isPending
                "
                class="inline-flex items-center gap-2 bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed shadow-sm hover:shadow-md active:scale-95"
                @click="doAction('stop')"
              >
                <Square class="w-4 h-4" />
                Stop
              </button>
              <button
                :disabled="
                  actionLoading || server?.status !== 'running' || isPending
                "
                class="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed shadow-sm hover:shadow-md active:scale-95"
                @click="doAction('restart')"
              >
                <RotateCcw class="w-4 h-4" />
                Restart
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Server Properties -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl p-6 md:p-8 shadow-sm"
      >
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">
            Server Properties
          </h2>
          <span
            class="text-xs font-medium text-gray-500 dark:text-neutral-400 bg-gray-100 dark:bg-neutral-700 px-2.5 py-1 rounded-lg"
          >
            server.properties
          </span>
        </div>
        <server-properties :agent-id="agentId" :server-id="serverId" />
      </div>
    </div>

    <div v-else class="text-gray-500 dark:text-neutral-400">Loading...</div>
  </div>
</template>

<script setup lang="ts">
import { formatDuration } from "~/composables/useDuration";
import {
  Activity,
  Box,
  Check,
  Clock,
  Globe,
  Hash,
  Pencil,
  Play,
  RefreshCw,
  RotateCcw,
  Server as ServerIcon,
  Square,
  Tag,
  Terminal,
  X as XIcon,
} from "lucide-vue-next";

definePageMeta({
  middleware: "auth",
});

interface Server {
  serverId: string;
  status: string;
  desiredStatus?: string;
  gamePort: number;
  publicRcon: boolean;
  rconPort?: number;
  restartPolicy?: string;
  engineType: string;
  gameVersion: string;
  startedAt?: string;
}

const route = useRoute();
const { agentId } = useCurrentAgent();
const { show: showToast } = useToast();
const { lastEvent } = useEventSource();

const serverId = route.params.serverId as string;
const server = ref<Server | null>(null);
const actionLoading = ref(false);
const iconTimestamp = ref(Date.now());
const iconError = ref(false);
const fileInput = ref<HTMLInputElement | null>(null);
const editingPort = ref(false);
const tempPort = ref<number | null>(null);
const portLoading = ref(false);
const editingRestartPolicy = ref(false);
const tempRestartPolicy = ref<string>("");
const restartPolicyLoading = ref(false);
const showIconEditor = ref(false);
const selectedIconFile = ref<File | null>(null);

const iconUrl = computed(() => {
  if (!server.value) return "";
  return `${useApiBase()}/agent/${agentId.value}/servers/${serverId}/icon?t=${iconTimestamp.value}`;
});

const startedAt = computed(() => {
  if (
    !server.value ||
    server.value.status !== "running" ||
    !server.value.startedAt
  ) {
    return undefined;
  }
  return server.value.startedAt;
});

const reactiveUptime = useUptime(startedAt);

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
    await $fetch(`/agent/${agentId.value}/servers/${serverId}/icon`, {
      baseURL: useApiBase(),
      method: "POST",
      body: formData,
      credentials: "include",
    });
    showToast("info", "Icon updated");
    iconTimestamp.value = Date.now();
    iconError.value = false;
  } catch (err: any) {
    const msg = err?.data?.detail || err?.message || "Failed to upload icon";
    showToast("error", "Upload failed", { description: msg });
  }
}

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
      return "bg-gray-100 text-gray-800 dark:bg-neutral-700 dark:text-neutral-300";
  }
}

async function savePort() {
  if (
    !server.value ||
    tempPort.value == null ||
    tempPort.value === server.value.gamePort
  ) {
    editingPort.value = false;
    return;
  }
  if (tempPort.value < 1024 || tempPort.value > 65535) {
    showToast("error", "Invalid port", {
      description: "Port must be between 1024 and 65535.",
    });
    return;
  }
  portLoading.value = true;
  try {
    await $fetch(`/agent/${agentId.value}/servers/${serverId}`, {
      baseURL: useApiBase(),
      method: "PATCH",
      credentials: "include",
      body: { gamePort: tempPort.value },
    });
    showToast("success", "Port updated", {
      description: `Game port changed to ${tempPort.value}.`,
    });
    editingPort.value = false;
    await refresh();
  } catch (err: any) {
    const status = err?.status || err?.statusCode;
    const msg = err?.data?.detail || err?.message || "Failed to update port";
    if (status === 409) {
      showToast("error", "Port unavailable", { description: msg });
    } else {
      showToast("error", "Update failed", { description: msg });
    }
  } finally {
    portLoading.value = false;
  }
}

async function saveRestartPolicy() {
  if (
    !server.value ||
    tempRestartPolicy.value === (server.value.restartPolicy || "no")
  ) {
    editingRestartPolicy.value = false;
    return;
  }
  restartPolicyLoading.value = true;
  try {
    await $fetch(`/agent/${agentId.value}/servers/${serverId}`, {
      baseURL: useApiBase(),
      method: "PATCH",
      credentials: "include",
      body: { restartPolicy: tempRestartPolicy.value },
    });
    showToast("success", "Restart policy updated", {
      description: `Policy changed to ${tempRestartPolicy.value}.`,
    });
    editingRestartPolicy.value = false;
    await refresh();
  } catch (err: any) {
    const msg =
      err?.data?.detail || err?.message || "Failed to update restart policy";
    showToast("error", "Update failed", { description: msg });
  } finally {
    restartPolicyLoading.value = false;
  }
}

async function doAction(action: "start" | "stop" | "restart") {
  if (actionLoading.value) return;
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
