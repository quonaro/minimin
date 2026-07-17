<template>
  <div
    v-if="hasVisible"
    class="fixed bottom-0 right-4 z-50 w-96 bg-white dark:bg-neutral-800 rounded-t-xl shadow-2xl border border-gray-200 dark:border-neutral-700 border-b-0 overflow-hidden"
  >
    <div
      class="px-4 py-3 flex items-center justify-between border-b border-gray-100 dark:border-neutral-700"
    >
      <span class="text-sm font-medium text-gray-900 dark:text-white">
        <template v-if="activeCount > 0">
          Backing up {{ activeCount }} server{{ activeCount !== 1 ? "s" : "" }}
        </template>
        <template v-else> Backups </template>
      </span>
      <button
        class="p-1 rounded hover:bg-gray-100 dark:hover:bg-neutral-700 text-gray-500"
        @click="expanded = !expanded"
      >
        <ChevronDown v-if="expanded" class="w-4 h-4" />
        <ChevronUp v-else class="w-4 h-4" />
      </button>
    </div>

    <Transition name="backup-collapse">
      <div v-if="expanded" class="px-4 py-3 space-y-3 overflow-hidden">
        <div
          v-if="tasks.length === 0"
          class="flex items-center justify-center text-sm text-gray-400 dark:text-neutral-500 py-4"
        >
          No backups in progress
        </div>
        <template v-else>
          <div
            v-for="task in tasks"
            :key="task.id"
            class="flex items-center gap-3"
          >
            <component
              :is="iconFor(task.status)"
              class="w-4 h-4 shrink-0"
              :class="iconClassFor(task.status)"
            />
            <div class="flex-1 min-w-0">
              <p class="text-sm text-gray-900 dark:text-white truncate">
                {{ task.name || "Backup" }}
              </p>
              <div
                class="w-full bg-gray-100 dark:bg-neutral-700 rounded-full h-1.5 mt-1"
              >
                <div
                  v-if="task.status === 'running'"
                  class="h-1.5 rounded-full bg-blue-500 animate-[loading-bar_1.5s_ease-in-out_infinite]"
                  style="width: 60%"
                />
                <div
                  v-else
                  class="h-1.5 rounded-full transition-all duration-300"
                  :class="barClassFor(task.status)"
                  :style="{ width: '100%' }"
                />
              </div>
              <p class="text-xs mt-0.5" :class="textClassFor(task.status)">
                {{ statusText(task) }}
              </p>
            </div>
            <button
              class="p-1 rounded hover:bg-gray-100 dark:hover:bg-neutral-700 shrink-0"
              :class="buttonClassFor(task.status)"
              @click="handleAction(task)"
            >
              <X v-if="isFinished(task.status)" class="w-3.5 h-3.5" />
              <X v-else class="w-3.5 h-3.5" />
            </button>
          </div>
        </template>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import {
  ChevronDown,
  ChevronUp,
  X,
  Check,
  AlertTriangle,
  Loader2,
} from "lucide-vue-next";
import type { BackupTask } from "~/composables/useBackupProgress";
import { useBackupProgress, removeBackupTask } from "~/composables/useBackupProgress";

const { tasks } = useBackupProgress();
const expanded = ref(true);

const activeCount = computed(
  () => tasks.value.filter((t) => t.status === "running").length,
);

const hasVisible = computed(
  () => tasks.value.length > 0,
);

function isFinished(status: BackupTask["status"]) {
  return status === "done" || status === "error";
}

function iconFor(status: BackupTask["status"]) {
  switch (status) {
    case "done":
      return Check;
    case "error":
      return AlertTriangle;
    case "running":
      return Loader2;
    default:
      return Loader2;
  }
}

function iconClassFor(status: BackupTask["status"]) {
  switch (status) {
    case "done":
      return "text-green-500";
    case "error":
      return "text-red-500";
    case "running":
      return "text-blue-500 animate-spin";
    default:
      return "text-gray-400";
  }
}

function barClassFor(status: BackupTask["status"]) {
  switch (status) {
    case "done":
      return "bg-green-500";
    case "error":
      return "bg-red-500";
    default:
      return "bg-gray-400";
  }
}

function textClassFor(status: BackupTask["status"]) {
  switch (status) {
    case "done":
      return "text-green-600 dark:text-green-400";
    case "error":
      return "text-red-600 dark:text-red-400";
    default:
      return "text-gray-500 dark:text-neutral-400";
  }
}

function buttonClassFor(status: BackupTask["status"]) {
  switch (status) {
    case "done":
      return "text-green-500 hover:text-green-700";
    case "error":
      return "text-red-500 hover:text-red-700";
    default:
      return "text-gray-500";
  }
}

function statusText(task: BackupTask): string {
  switch (task.status) {
    case "done":
      return `Done in ${formatDuration(task.elapsedMs)}`;
    case "error":
      return task.error || "Failed";
    case "running":
      return `Backing up… ${formatDuration(task.elapsedMs)}`;
    default:
      return "";
  }
}

function handleAction(task: BackupTask) {
  if (isFinished(task.status)) {
    removeBackupTask(task.id);
  } else {
    removeBackupTask(task.id);
  }
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}
</script>

<style scoped>
.backup-collapse-enter-active,
.backup-collapse-leave-active {
  transition:
    max-height 0.3s ease,
    opacity 0.3s ease;
}

.backup-collapse-enter-from,
.backup-collapse-leave-to {
  max-height: 0;
  opacity: 0;
}

.backup-collapse-enter-to,
.backup-collapse-leave-from {
  max-height: 500px;
  opacity: 1;
}

@keyframes loading-bar {
  0% {
    transform: translateX(-100%);
  }
  50% {
    transform: translateX(0%);
  }
  100% {
    transform: translateX(100%);
  }
}
</style>
