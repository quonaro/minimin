<script setup lang="ts">
import {
  Camera,
  Loader2,
  OctagonAlert,
  Play,
  RefreshCw,
  RotateCcw,
  Server as ServerIcon,
  Square,
  Trash2,
} from "lucide-vue-next";
import type { Server } from "~/composables/useServers";

defineProps<{
  server: Server;
  serverId: string;
  iconUrl: string;
  iconError: boolean;
  currentAction: "start" | "stop" | "force-stop" | "restart" | null;
  isPending: boolean;
  removeBeforeStart: boolean;
  deleteLoading: boolean;
  recreateLoading: boolean;
}>();

const emit = defineEmits<{
  (e: "update:iconError", v: boolean): void;
  (e: "iconSelect", event: Event): void;
  (e: "iconSave", blob: Blob): void;
  (e: "update:removeBeforeStart", v: boolean): void;
  (e: "action", action: "start" | "stop" | "restart" | "force-stop"): void;
  (e: "delete"): void;
  (e: "recreate"): void;
}>();

const showIconEditor = defineModel<boolean>("showIconEditor", {
  default: false,
});
const selectedIconFile = defineModel<File | null>("selectedIconFile", {
  default: null,
});

const fileInput = ref<HTMLInputElement | null>(null);
</script>

<template>
  <div class="flex flex-col items-center gap-4 shrink-0 max-w-[200px]">
    <div class="relative w-full">
      <div
        class="w-full aspect-square rounded-2xl overflow-hidden shadow-lg bg-gradient-to-br from-gray-200 to-gray-300 dark:from-neutral-700 dark:to-neutral-600 flex items-center justify-center ring-4 ring-white dark:ring-neutral-800"
      >
        <img
          v-if="iconUrl && !iconError"
          :src="iconUrl"
          alt="Server icon"
          class="w-full h-full object-cover"
          @error="$emit('update:iconError', true)"
        />
        <ServerIcon
          v-else
          class="w-10 h-10 text-indigo-500 dark:text-indigo-400"
        />
      </div>
      <button
        v-if="server.hostPath"
        class="absolute inset-0 flex items-center justify-center bg-black/40 opacity-0 hover:opacity-100 transition-opacity rounded-2xl cursor-pointer"
        @click="fileInput?.click()"
      >
        <Camera class="w-6 h-6 text-white" />
      </button>
      <span
        :class="[
          'absolute -bottom-1 -right-1 w-5 h-5 rounded-full border-2 border-white dark:border-neutral-800',
          server.serverStatus === 'running'
            ? 'bg-green-500'
            : server.serverStatus === 'starting'
              ? 'bg-yellow-500'
              : server.serverStatus === 'pulling_image'
                ? 'bg-blue-500'
                : server.containerStatus === 'exited'
                  ? 'bg-red-500'
                  : 'bg-gray-400',
        ]"
      />
    </div>
    <input
      ref="fileInput"
      type="file"
      accept="image/png,image/jpeg,image/jpg,image/gif,image/webp,image/bmp"
      class="hidden"
      @change="$emit('iconSelect', $event)"
    />
    <server-icon-editor
      v-model="showIconEditor"
      :file="selectedIconFile"
      @save="$emit('iconSave', $event)"
    />

    <!-- Actions -->
    <div class="flex flex-col items-stretch gap-2 w-full">
      <div class="flex flex-col gap-2">
        <div class="flex items-center gap-2">
          <button
            :disabled="
              currentAction === 'start' ||
              server?.containerStatus === 'running' ||
              isPending
            "
            class="flex items-center gap-2 flex-1 justify-center bg-gray-100 dark:bg-neutral-700/50 hover:bg-gray-200 dark:hover:bg-neutral-700 text-gray-700 dark:text-gray-200 border border-gray-200 dark:border-neutral-600 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed active:scale-95"
            @click="$emit('action', 'start')"
          >
            <Loader2
              v-if="currentAction === 'start'"
              class="w-4 h-4 animate-spin"
            />
            <Play v-else class="w-4 h-4" />
            Start
          </button>
          <label
            class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-neutral-400 cursor-pointer select-none shrink-0"
          >
            <input
              :checked="removeBeforeStart"
              type="checkbox"
              class="w-3.5 h-3.5 rounded border-gray-300 dark:border-neutral-600 text-indigo-600 dark:text-indigo-400 focus:ring-indigo-500 dark:focus:ring-indigo-400 bg-white dark:bg-neutral-700 cursor-pointer"
              @change="
                $emit(
                  'update:removeBeforeStart',
                  ($event.target as HTMLInputElement).checked,
                )
              "
            />
            Recreate
          </label>
        </div>
        <button
          :disabled="
            currentAction === 'stop' ||
            server?.containerStatus !== 'running' ||
            isPending
          "
          class="flex items-center gap-2 w-full justify-center bg-gray-100 dark:bg-neutral-700/50 hover:bg-gray-200 dark:hover:bg-neutral-700 text-gray-700 dark:text-gray-200 border border-gray-200 dark:border-neutral-600 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed active:scale-95"
          @click="$emit('action', 'stop')"
        >
          <Loader2
            v-if="currentAction === 'stop'"
            class="w-4 h-4 animate-spin"
          />
          <Square v-else class="w-4 h-4" />
          Stop
        </button>
        <button
          :disabled="
            currentAction === 'force-stop' ||
            server?.containerStatus !== 'running' ||
            isPending
          "
          class="flex items-center gap-2 w-full justify-center bg-gray-100 dark:bg-neutral-700/50 hover:bg-gray-200 dark:hover:bg-neutral-700 text-gray-700 dark:text-gray-200 border border-gray-200 dark:border-neutral-600 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed active:scale-95"
          @click="$emit('action', 'force-stop')"
        >
          <Loader2
            v-if="currentAction === 'force-stop'"
            class="w-4 h-4 animate-spin"
          />
          <OctagonAlert v-else class="w-4 h-4" />
          Force Stop
        </button>
        <button
          :disabled="
            currentAction === 'restart' ||
            server?.containerStatus !== 'running' ||
            isPending
          "
          class="flex items-center gap-2 w-full justify-center bg-gray-100 dark:bg-neutral-700/50 hover:bg-gray-200 dark:hover:bg-neutral-700 text-gray-700 dark:text-gray-200 border border-gray-200 dark:border-neutral-600 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed active:scale-95"
          @click="$emit('action', 'restart')"
        >
          <Loader2
            v-if="currentAction === 'restart'"
            class="w-4 h-4 animate-spin"
          />
          <RotateCcw v-else class="w-4 h-4" />
          Restart
        </button>
      </div>

      <div
        class="mt-3 pt-3 border-t border-gray-200 dark:border-neutral-700 flex flex-col gap-2"
      >
        <button
          v-if="
            server.containerStatus === 'running' ||
            server.containerStatus === 'exited'
          "
          :disabled="recreateLoading || isPending"
          class="flex items-center gap-2 w-full justify-center bg-amber-500/10 hover:bg-amber-500/20 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-800/50 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed active:scale-95"
          @click="$emit('recreate')"
        >
          <RefreshCw class="w-4 h-4" />
          Recreate World
        </button>
        <button
          :disabled="deleteLoading || isPending"
          class="flex items-center gap-2 w-full justify-center bg-red-500/10 hover:bg-red-500/20 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800/50 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed active:scale-95"
          @click="$emit('delete')"
        >
          <Trash2 class="w-4 h-4" />
          Delete Server
        </button>
      </div>
    </div>
  </div>
</template>
