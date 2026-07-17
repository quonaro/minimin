<template>
  <div class="space-y-3">
    <h2
      v-if="showTitle"
      class="text-lg font-semibold text-gray-900 dark:text-white"
    >
      Action Log
    </h2>
    <div v-if="loading" class="text-sm text-gray-500 dark:text-neutral-400">
      Loading...
    </div>
    <div v-else-if="error" class="text-sm text-red-500 dark:text-red-400">
      {{ error }}
    </div>
    <div
      v-else-if="entries.length === 0"
      class="text-sm text-gray-500 dark:text-neutral-400"
    >
      No actions recorded yet.
    </div>
    <div v-else class="space-y-2">
      <div
        v-for="entry in entries"
        :key="entry.id"
        class="flex items-center gap-3 px-3 py-2 rounded-lg bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700"
      >
        <div class="flex-shrink-0">
          <Check
            v-if="entry.status === 'success'"
            class="w-4 h-4 text-green-500"
          />
          <AlertTriangle
            v-else-if="entry.status === 'error'"
            class="w-4 h-4 text-red-500"
          />
          <SkipForward v-else class="w-4 h-4 text-gray-400" />
        </div>
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <span
              class="text-sm font-medium text-gray-900 dark:text-white capitalize"
            >
              {{ entry.action }}
            </span>
            <span
              class="text-xs px-1.5 py-0.5 rounded-full font-medium"
              :class="sourceClass(entry.source)"
            >
              {{ entry.source }}
            </span>
          </div>
          <p
            v-if="entry.detail"
            class="text-xs text-gray-500 dark:text-neutral-400 truncate"
          >
            {{ entry.detail }}
          </p>
          <p class="text-xs text-gray-400 dark:text-neutral-500">
            {{ formatDate(entry.timestamp) }}
            <span v-if="entry.message" class="ml-1">· {{ entry.message }}</span>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Check, AlertTriangle, SkipForward } from "lucide-vue-next";
import type { ActionLogEntry } from "~/composables/useActionLog";

const props = withDefaults(
  defineProps<{
    serverId: string;
    showTitle?: boolean;
  }>(),
  {
    showTitle: true,
  },
);

const { entries, loading, error, fetchLog } = useActionLog(props.serverId);

onMounted(() => {
  fetchLog();
});

function sourceClass(source: string) {
  switch (source) {
    case "manual":
      return "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300";
    case "interval":
    case "cron":
      return "bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300";
    case "event":
      return "bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300";
    default:
      return "bg-gray-100 text-gray-700 dark:bg-neutral-700 dark:text-neutral-300";
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString();
}
</script>
