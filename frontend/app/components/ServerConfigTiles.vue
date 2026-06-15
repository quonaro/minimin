<script setup lang="ts">
import {
  Check,
  ChevronDown,
  Code,
  Coffee,
  MemoryStick,
  Package,
  Pencil,
  Plus,
  RefreshCw,
  Tag,
  X as XIcon,
} from "lucide-vue-next";
import type { Server } from "~/composables/useServers";

const props = defineProps<{
  server: Server;
}>();

const serverId = computed(() => props.server.serverId);

const {
  editingRestartPolicy,
  tempRestartPolicy,
  restartPolicyLoading,
  saveRestartPolicy,
  editingRam,
  tempRamGb,
  ramLoading,
  saveRam,
  editingExternalJavaArgs,
  tempExternalJavaArgs,
  newExternalArg,
  externalJavaArgsLoading,
  handlePasteExternalJavaArgs,
  saveExternalJavaArgs,
} = useServerConfigEdits(serverId.value, toRef(props, "server"));

function getEngineAbbreviation(engineType: string) {
  switch (engineType.toUpperCase()) {
    case "FORGE":
      return "FO";
    case "FABRIC":
      return "FA";
    case "PAPERMC":
      return "PA";
    case "VANILLA":
      return "MC";
    default:
      return engineType.substring(0, 2).toUpperCase();
  }
}

function getEngineIconColor(engineType: string) {
  switch (engineType.toUpperCase()) {
    case "FORGE":
      return "bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400";
    case "FABRIC":
      return "bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400";
    case "PAPERMC":
      return "bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400";
    case "VANILLA":
      return "bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400";
    default:
      return "bg-gray-100 text-gray-600 dark:bg-neutral-700 dark:text-neutral-400";
  }
}

function getJavaVersion(imageName?: string): string {
  if (!imageName) return "—";
  const tag = imageName.split(":")[1] || "";
  const match = tag.match(/^java(\d+)$/);
  return match && match[1] ? match[1] : tag;
}
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-6">
    <!-- Server Config -->
    <div class="space-y-2">
      <p
        class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
      >
        Server
      </p>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
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
            <div v-if="!editingRestartPolicy" class="flex items-center gap-2">
              <p class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ server.restartPolicy || "no" }}
              </p>
              <button
                v-if="server.containerStatus !== 'running'"
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
            :class="[
              'w-9 h-9 shrink-0 rounded-lg flex items-center justify-center text-xs font-bold',
              getEngineIconColor(server.engineType),
            ]"
          >
            {{ getEngineAbbreviation(server.engineType) }}
          </div>
          <div class="min-w-0">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              Engine
            </p>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
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
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ server.gameVersion }}
            </p>
          </div>
        </div>

        <!-- Loader -->
        <div
          class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400"
          >
            <Package class="w-4 h-4" />
          </div>
          <div class="min-w-0 flex-1">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              Loader
            </p>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ server.loaderVersion || "—" }}
            </p>
          </div>
        </div>

        <!-- Java Version -->
        <div
          class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center text-orange-600 dark:text-orange-400"
          >
            <Coffee class="w-4 h-4" />
          </div>
          <div class="min-w-0">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              Java Version
            </p>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ getJavaVersion(server.imageName) }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Resources -->
    <div class="space-y-2">
      <p
        class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
      >
        Resources
      </p>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <!-- RAM -->
        <div
          class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400"
          >
            <MemoryStick class="w-4 h-4" />
          </div>
          <div class="min-w-0 flex-1">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              RAM
            </p>
            <div v-if="!editingRam" class="flex items-center gap-2">
              <p class="text-sm font-semibold text-gray-900 dark:text-white">
                {{
                  server.ramBytes
                    ? (server.ramBytes / (1024 * 1024 * 1024)).toFixed(0) +
                      " GB"
                    : "-"
                }}
              </p>
              <button
                v-if="server.containerStatus !== 'running'"
                class="text-gray-400 hover:text-primary transition-colors"
                :disabled="ramLoading"
                @click="
                  tempRamGb = server.ramBytes
                    ? Math.round(server.ramBytes / (1024 * 1024 * 1024))
                    : 2;
                  editingRam = true;
                "
              >
                <Pencil class="w-3.5 h-3.5" />
              </button>
            </div>
            <div v-else class="flex items-center gap-2">
              <number-input
                v-model="tempRamGb"
                :min="1"
                :max="128"
                size="sm"
                class="w-24"
                @keyup.enter="saveRam"
              />
              <span class="text-sm text-gray-500 dark:text-neutral-400"
                >GB</span
              >
              <button
                class="text-green-500 hover:text-green-600 transition-colors"
                :disabled="ramLoading"
                @click="saveRam"
              >
                <Check class="w-4 h-4" />
              </button>
              <button
                class="text-red-500 hover:text-red-600 transition-colors"
                :disabled="ramLoading"
                @click="editingRam = false"
              >
                <XIcon class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <!-- External Java Args -->
        <div
          class="col-span-full flex items-start gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-sky-100 dark:bg-sky-900/30 flex items-center justify-center text-sky-600 dark:text-sky-400"
          >
            <Code class="w-4 h-4" />
          </div>
          <div class="min-w-0 flex-1">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              External Java Args
            </p>
            <div
              v-if="!editingExternalJavaArgs"
              class="flex items-center gap-2"
            >
              <div class="flex flex-wrap gap-1">
                <span
                  v-if="
                    !server.externalJavaArgs ||
                    server.externalJavaArgs.length === 0
                  "
                  class="text-sm font-semibold text-gray-900 dark:text-white"
                >
                  None
                </span>
                <span
                  v-for="(arg, idx) in server.externalJavaArgs"
                  :key="idx"
                  class="inline-flex items-center px-1.5 py-0.5 rounded text-[11px] font-medium bg-sky-100 dark:bg-sky-900/30 text-sky-700 dark:text-sky-400 border border-sky-200 dark:border-sky-800"
                >
                  {{ arg }}
                </span>
              </div>
              <button
                v-if="server.containerStatus !== 'running'"
                class="text-gray-400 hover:text-primary transition-colors"
                :disabled="externalJavaArgsLoading"
                @click="
                  tempExternalJavaArgs = server.externalJavaArgs
                    ? [...server.externalJavaArgs]
                    : [];
                  newExternalArg = '';
                  editingExternalJavaArgs = true;
                "
              >
                <Pencil class="w-3.5 h-3.5" />
              </button>
            </div>
            <div v-else class="flex flex-col gap-2">
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="(arg, idx) in tempExternalJavaArgs"
                  :key="idx"
                  class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-medium bg-sky-100 dark:bg-sky-900/30 text-sky-700 dark:text-sky-400 border border-sky-200 dark:border-sky-800"
                >
                  {{ arg }}
                  <button
                    class="hover:text-red-500 transition-colors"
                    @click="tempExternalJavaArgs.splice(idx, 1)"
                  >
                    <XIcon class="w-3 h-3" />
                  </button>
                </span>
              </div>
              <div class="flex items-center gap-2">
                <input
                  v-model="newExternalArg"
                  type="text"
                  class="flex-1 min-w-0 text-sm bg-white dark:bg-neutral-700 border border-gray-200 dark:border-neutral-600 rounded-lg px-2 py-1 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:outline-none font-mono"
                  placeholder="-XX:+UseG1GC"
                  @paste="handlePasteExternalJavaArgs($event)"
                  @keyup.enter.prevent="
                    if (newExternalArg.trim()) {
                      tempExternalJavaArgs.push(newExternalArg.trim());
                      newExternalArg = '';
                    }
                  "
                />
                <button
                  class="text-primary hover:text-primary/80 transition-colors"
                  @click="
                    if (newExternalArg.trim()) {
                      tempExternalJavaArgs.push(newExternalArg.trim());
                      newExternalArg = '';
                    }
                  "
                >
                  <Plus class="w-4 h-4" />
                </button>
                <button
                  class="text-green-500 hover:text-green-600 transition-colors"
                  :disabled="externalJavaArgsLoading"
                  @click="saveExternalJavaArgs"
                >
                  <Check class="w-4 h-4" />
                </button>
                <button
                  class="text-red-500 hover:text-red-600 transition-colors"
                  :disabled="externalJavaArgsLoading"
                  @click="editingExternalJavaArgs = false"
                >
                  <XIcon class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
