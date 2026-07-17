<template>
  <div
    class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden"
  >
    <button
      @click="expanded = !expanded"
      class="w-full flex items-center justify-between p-4 md:p-6 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
    >
      <div class="flex items-center gap-3">
        <Database class="w-5 h-5 text-green-500" />
        <h2 class="text-xl font-bold text-gray-900 dark:text-white">Backups</h2>
      </div>
      <ChevronDown
        class="w-5 h-5 text-gray-500 dark:text-neutral-400 transition-transform"
        :class="{ 'rotate-180': expanded }"
      />
    </button>

    <div
      v-show="expanded"
      class="px-4 pt-4 pb-4 md:px-6 md:pt-6 md:pb-6 space-y-4"
    >
      <div class="flex justify-end">
        <button
          @click="createBackup"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
        >
          Create Backup
        </button>
      </div>

      <div
        v-if="loading"
        class="text-center py-12 text-gray-500 dark:text-neutral-400"
      >
        Loading backups...
      </div>
      <div
        v-else-if="error"
        class="text-center py-12 text-red-500 dark:text-red-400"
      >
        {{ error }}
      </div>
      <div
        v-else-if="backups.length === 0"
        class="text-center py-12 text-gray-500 dark:text-neutral-400"
      >
        No backups yet.
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="bk in backups"
          :key="bk.name"
          class="bg-gray-50 dark:bg-neutral-700/50 border border-gray-200 dark:border-neutral-700 rounded-xl p-4 flex items-center gap-4"
        >
          <div class="flex-1 min-w-0">
            <h3 class="font-semibold text-gray-900 dark:text-white">
              {{ bk.name }}
            </h3>
            <p class="text-sm text-gray-500 dark:text-neutral-400">
              {{ formatBytes(bk.sizeBytes) }} · {{ formatDate(bk.createdAt) }}
            </p>
          </div>
          <div class="flex items-center gap-2 flex-shrink-0">
            <button
              @click="doDownload(bk.name)"
              class="px-3 py-1.5 text-sm rounded-lg bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors"
            >
              Download
            </button>
            <button
              @click="doCopyLink(bk.name)"
              title="Copy direct link"
              class="p-1.5 rounded-lg bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors"
            >
              <Link class="w-4 h-4" />
            </button>
            <button
              @click="restoreBackup(bk.name)"
              class="px-3 py-1.5 text-sm rounded-lg bg-amber-600 text-white hover:bg-amber-700 shadow-sm transition-colors"
            >
              Restore
            </button>
            <button
              @click="deleteBackup(bk.name)"
              class="px-3 py-1.5 text-sm rounded-lg bg-red-600 text-white hover:bg-red-700 shadow-sm transition-colors"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ChevronDown, Database, Link } from "lucide-vue-next";
import { addBackupTask } from "~/composables/useBackupProgress";

const props = defineProps<{
  serverId: string;
}>();

const {
  backups,
  loading,
  error,
  fetchBackups,
  createBackup: doCreate,
  restoreBackup: doRestore,
  deleteBackup: doDelete,
  downloadBackup: doDownload,
  copyDownloadLink: doCopyLink,
} = useBackups(props.serverId);

const expanded = ref(true);

onMounted(() => {
  fetchBackups();
});

function createBackup() {
  addBackupTask(props.serverId, doCreate);
}

async function restoreBackup(name: string) {
  if (
    !confirm(`Restore backup "${name}"? This will overwrite the current world.`)
  )
    return;
  await doRestore(name);
}

async function deleteBackup(name: string) {
  if (!confirm(`Delete backup "${name}"?`)) return;
  await doDelete(name);
}

function formatBytes(bytes: number) {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString();
}
</script>
