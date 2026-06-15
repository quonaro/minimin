<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    @click.self="emit('cancel')"
  >
    <div
      class="bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 w-full max-w-md flex flex-col max-h-[80vh]"
    >
      <div
        class="p-6 border-b border-gray-200 dark:border-neutral-700 shrink-0"
      >
        <h2 class="text-xl font-bold text-gray-900 dark:text-white">
          ZIP Preview
        </h2>
        <p class="text-sm text-gray-500 dark:text-neutral-400 mt-1 truncate">
          {{ fileName }}
          <span v-if="fileSize" class="text-xs ml-1">
            {{ formatBytes(fileSize) }}
          </span>
        </p>
      </div>

      <div class="flex-1 min-h-0 overflow-y-auto p-6">
        <template v-if="entries.length > 0">
          <p class="text-sm text-gray-700 dark:text-neutral-300 mb-3">
            Contains {{ entries.length }} file{{
              entries.length !== 1 ? "s" : ""
            }}:
          </p>
          <div class="space-y-1">
            <div
              v-for="entry in entries"
              :key="entry"
              class="flex items-center gap-2 px-3 py-2 rounded-lg bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
            >
              <File
                class="w-4 h-4 text-gray-400 dark:text-neutral-500 shrink-0"
              />
              <p class="text-sm text-gray-900 dark:text-white truncate">
                {{ entry }}
              </p>
            </div>
          </div>
        </template>
        <template v-else>
          <div
            class="flex flex-col items-center justify-center text-center py-8"
          >
            <Archive class="w-10 h-10 text-amber-500 mb-3" />
            <p class="text-sm font-medium text-gray-900 dark:text-white">
              Large archive
            </p>
            <p
              class="text-xs text-gray-500 dark:text-neutral-400 mt-1 max-w-xs"
            >
              Archive is too large to preview contents on the client. The server
              will extract it after upload.
            </p>
          </div>
        </template>
      </div>

      <div
        class="p-6 border-t border-gray-200 dark:border-neutral-700 flex gap-3 justify-end shrink-0"
      >
        <button
          class="px-4 py-2 rounded-lg text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors font-medium text-sm"
          @click="emit('cancel')"
        >
          Cancel
        </button>
        <button
          class="px-4 py-2 rounded-lg bg-primary hover:bg-primary/90 text-white font-medium text-sm transition-colors"
          @click="emit('confirm')"
        >
          Upload
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Archive, File } from "lucide-vue-next";

interface Props {
  fileName: string;
  entries: string[];
  fileSize?: number;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  confirm: [];
  cancel: [];
}>();

function formatBytes(n: number): string {
  if (n >= 1024 * 1024 * 1024)
    return (n / (1024 * 1024 * 1024)).toFixed(1) + " GB";
  if (n >= 1024 * 1024) return (n / (1024 * 1024)).toFixed(1) + " MB";
  if (n >= 1024) return (n / 1024).toFixed(1) + " KB";
  return n + " B";
}
</script>
