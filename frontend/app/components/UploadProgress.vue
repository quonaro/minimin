<template>
  <div
    v-if="visibleUploads.length > 0 && !closed"
    class="fixed bottom-4 right-4 z-50 w-80 bg-white dark:bg-neutral-800 rounded-xl shadow-2xl border border-gray-200 dark:border-neutral-700 overflow-hidden"
  >
    <div
      class="px-4 py-3 flex items-center justify-between border-b border-gray-100 dark:border-neutral-700"
    >
      <span class="text-sm font-medium text-gray-900 dark:text-white">
        Uploading {{ visibleUploads.length }} file{{ visibleUploads.length !== 1 ? "s" : "" }}
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
        v-if="remainingText"
        class="flex items-center justify-between text-xs text-gray-500 dark:text-neutral-400"
      >
        <span>{{ remainingText }}</span>
        <button class="text-primary hover:underline" @click="cancelAll">
          Cancel all
        </button>
      </div>

      <div
        v-for="task in visibleUploads"
        :key="task.id"
        class="flex items-center gap-3"
      >
        <FileText class="w-4 h-4 text-gray-400 shrink-0" />
        <div class="flex-1 min-w-0">
          <p class="text-sm text-gray-900 dark:text-white truncate">
            {{ task.fileName }}
          </p>
          <div class="w-full bg-gray-100 dark:bg-neutral-700 rounded-full h-1.5 mt-1">
            <div
              class="bg-primary h-1.5 rounded-full transition-all duration-300"
              :style="{ width: task.percentage + '%' }"
            />
          </div>
          <p class="text-xs text-gray-500 dark:text-neutral-400 mt-0.5">
            {{ task.percentage }}% · {{ formatSpeed(task.speed) }} ·
            {{ formatTime(task.remainingSeconds) }}
          </p>
        </div>
        <button
          class="p-1 rounded hover:bg-gray-100 dark:hover:bg-neutral-700 text-gray-500 shrink-0"
          @click="cancelUpload(task.id)"
        >
          <X class="w-3.5 h-3.5" />
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
} from "lucide-vue-next";
import { useUploadQueue } from "~/composables/useUploadQueue";

const { uploads, cancelUpload } = useUploadQueue();
const expanded = ref(true);
const closed = ref(false);

const visibleUploads = computed(() =>
  uploads.value.filter((u) => u.status === "pending" || u.status === "uploading"),
);

watch(
  visibleUploads,
  (val) => {
    if (val.length > 0) closed.value = false;
  },
  { immediate: true },
);

const remainingText = computed(() => {
  const maxRemaining = Math.max(
    0,
    ...visibleUploads.value.map((u) => u.remainingSeconds),
  );
  if (!maxRemaining || !isFinite(maxRemaining)) return "";
  return formatTime(maxRemaining);
});

function cancelAll() {
  visibleUploads.value.forEach((u) => cancelUpload(u.id));
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
</script>
