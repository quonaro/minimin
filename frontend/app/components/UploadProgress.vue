<template>
  <div
    v-if="uploads.length > 0 && !closed"
    class="fixed bottom-4 right-4 z-50 w-80 bg-white dark:bg-neutral-800 rounded-xl shadow-2xl border border-gray-200 dark:border-neutral-700 overflow-hidden"
  >
    <div
      class="px-4 py-3 flex items-center justify-between border-b border-gray-100 dark:border-neutral-700"
    >
      <span class="text-sm font-medium text-gray-900 dark:text-white">
        <template v-if="activeCount > 0">
          Uploading {{ activeCount }} file{{ activeCount !== 1 ? "s" : "" }}
        </template>
        <template v-else> Uploads </template>
      </span>
      <div class="flex items-center gap-1">
        <button
          class="p-1 rounded hover:bg-gray-100 dark:hover:bg-neutral-700 text-gray-500"
          @click="expanded = !expanded"
        >
          <ChevronDown v-if="expanded" class="w-4 h-4" />
          <ChevronUp v-else class="w-4 h-4" />
        </button>
        <button
          class="p-1 rounded hover:bg-gray-100 dark:hover:bg-neutral-700 text-gray-500"
          @click="closed = true"
        >
          <X class="w-4 h-4" />
        </button>
      </div>
    </div>

    <div v-if="expanded" class="px-4 py-3 space-y-3">
      <div
        v-if="remainingText && activeCount > 0"
        class="flex items-center justify-between text-xs text-gray-500 dark:text-neutral-400"
      >
        <span>{{ remainingText }}</span>
        <button class="text-primary hover:underline" @click="cancelAll">
          Cancel all
        </button>
      </div>

      <div
        v-for="task in uploads"
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
            {{ task.fileName }}
          </p>
          <div
            class="w-full bg-gray-100 dark:bg-neutral-700 rounded-full h-1.5 mt-1"
          >
            <div
              class="h-1.5 rounded-full transition-all duration-300"
              :class="barClassFor(task.status)"
              :style="{ width: task.percentage + '%' }"
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
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ChevronDown,
  ChevronUp,
  X,
  FileText,
  Check,
  AlertTriangle,
  Loader2,
} from "lucide-vue-next";
import type { UploadTask } from "~/composables/useUploadQueue";
import { useUploadQueue } from "~/composables/useUploadQueue";

const { uploads, cancelUpload, removeUpload } = useUploadQueue();
const expanded = ref(true);
const closed = ref(false);

const activeCount = computed(
  () =>
    uploads.value.filter(
      (u) => u.status === "pending" || u.status === "uploading",
    ).length,
);

watch(
  () => uploads.value.length,
  () => {
    if (uploads.value.length > 0) closed.value = false;
  },
  { immediate: true },
);

const remainingText = computed(() => {
  const active = uploads.value.filter((u) => u.status === "uploading");
  const maxRemaining = Math.max(0, ...active.map((u) => u.remainingSeconds));
  if (!maxRemaining || !isFinite(maxRemaining)) return "";
  return formatTime(maxRemaining);
});

function isFinished(status: UploadTask["status"]) {
  return status === "done" || status === "error" || status === "cancelled";
}

function iconFor(status: UploadTask["status"]) {
  switch (status) {
    case "done":
      return Check;
    case "error":
      return AlertTriangle;
    case "cancelled":
      return X;
    case "pending":
      return Loader2;
    default:
      return FileText;
  }
}

function iconClassFor(status: UploadTask["status"]) {
  switch (status) {
    case "done":
      return "text-green-500";
    case "error":
      return "text-red-500";
    case "cancelled":
      return "text-gray-400";
    case "pending":
      return "text-gray-400 animate-spin";
    default:
      return "text-gray-400";
  }
}

function barClassFor(status: UploadTask["status"]) {
  switch (status) {
    case "done":
      return "bg-green-500";
    case "error":
      return "bg-red-500";
    case "cancelled":
      return "bg-gray-400";
    default:
      return "bg-primary";
  }
}

function textClassFor(status: UploadTask["status"]) {
  switch (status) {
    case "done":
      return "text-green-600 dark:text-green-400";
    case "error":
      return "text-red-600 dark:text-red-400";
    case "cancelled":
      return "text-gray-500 dark:text-neutral-400";
    default:
      return "text-gray-500 dark:text-neutral-400";
  }
}

function buttonClassFor(status: UploadTask["status"]) {
  switch (status) {
    case "done":
      return "text-green-500 hover:text-green-700";
    case "error":
      return "text-red-500 hover:text-red-700";
    default:
      return "text-gray-500";
  }
}

function statusText(task: UploadTask): string {
  switch (task.status) {
    case "done":
      return task.duration
        ? `Uploaded in ${formatDuration(task.duration)}`
        : "Uploaded";
    case "error":
      return task.duration
        ? `Failed after ${formatDuration(task.duration)}`
        : "Failed";
    case "cancelled":
      return "Cancelled";
    case "pending":
      return "Waiting…";
    case "uploading":
      return `${task.percentage}% · ${formatSpeed(task.speed)} · ${formatTime(task.remainingSeconds)}`;
    default:
      return "";
  }
}

function handleAction(task: UploadTask) {
  if (isFinished(task.status)) {
    removeUpload(task.id);
  } else {
    cancelUpload(task.id);
  }
}

function cancelAll() {
  uploads.value
    .filter((u) => !isFinished(u.status))
    .forEach((u) => cancelUpload(u.id));
}

function formatSpeed(bps: number): string {
  if (!bps || !isFinite(bps)) return "-";
  if (bps > 1024 * 1024) return (bps / (1024 * 1024)).toFixed(1) + " MB/s";
  if (bps > 1024) return (bps / 1024).toFixed(1) + " KB/s";
  return bps.toFixed(0) + " B/s";
}

function formatTime(seconds: number): string {
  if (!seconds || !isFinite(seconds)) return "";
  if (seconds < 60) return Math.ceil(seconds) + " sec left";
  if (seconds < 3600) return Math.ceil(seconds / 60) + " min left";
  return Math.ceil(seconds / 3600) + " hr left";
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}
</script>
