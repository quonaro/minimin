<template>
  <div
    class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden"
  >
    <div
      class="w-full flex items-center justify-between p-4 md:p-6 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
    >
      <button
        @click="expanded = !expanded"
        class="flex items-center gap-3 flex-1 text-left"
      >
        <Database class="w-5 h-5 text-green-500" />
        <h2 class="text-xl font-bold text-gray-900 dark:text-white">Backups</h2>
      </button>
      <div class="flex items-center gap-3">
        <button
          @click.stop="createBackup"
          class="px-3 py-1.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
        >
          Create Backup
        </button>
        <button
          @click="expanded = !expanded"
          class="p-1 text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-200 transition-colors"
        >
          <ChevronDown
            class="w-5 h-5 transition-transform"
            :class="{ 'rotate-180': expanded }"
          />
        </button>
      </div>
    </div>

    <div
      v-show="expanded"
      class="px-4 pt-4 pb-4 md:px-6 md:pt-6 md:pb-6 space-y-4"
    >
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

      <div
        v-else
        class="max-h-[32rem] overflow-y-auto pr-1 space-y-3 no-scrollbar"
      >
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
              class="h-9 px-3 py-1.5 text-sm rounded-lg bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors"
            >
              Download
            </button>
            <button
              @click="doCopyLink(bk.name)"
              title="Copy direct link"
              class="h-9 w-9 flex items-center justify-center rounded-lg bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors"
            >
              <Link class="w-4 h-4" />
            </button>
            <button
              @click="openRestoreDialog(bk.name)"
              class="h-9 px-3 py-1.5 text-sm rounded-lg bg-amber-600 text-white hover:bg-amber-700 shadow-sm transition-colors"
            >
              Restore
            </button>
            <button
              @click="openDeleteDialog(bk.name)"
              class="h-9 px-3 py-1.5 text-sm rounded-lg bg-red-600 text-white hover:bg-red-700 shadow-sm transition-colors"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>

    <confirm-dialog
      v-model="showRestoreDialog"
      title="Restore Backup"
      :description="`Restore backup &quot;${selectedBackup}&quot;? This will overwrite the current world.`"
      :expected-text="serverName"
      confirm-label="Restore"
      @confirm="onRestoreConfirmed"
    />
    <confirm-dialog
      v-model="showDeleteDialog"
      title="Delete Backup"
      :description="`Delete backup &quot;${selectedBackup}&quot;?`"
      :expected-text="selectedBackup"
      confirm-label="Delete"
      danger
      @confirm="onDeleteConfirmed"
    />
  </div>
</template>

<script setup lang="ts">
import { ChevronDown, Database, Link } from "lucide-vue-next";
import ConfirmDialog from "~/components/ConfirmDialog.vue";
import { addBackupTask } from "~/composables/useBackupProgress";
import { useServerEvents } from "~/composables/useServerEvents";

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
const showRestoreDialog = ref(false);
const showDeleteDialog = ref(false);
const selectedBackup = ref("");

const { serverMap } = useServerEvents();
const serverName = computed(
  () => serverMap.value[props.serverId]?.name || props.serverId,
);

onMounted(() => {
  fetchBackups();
});

function createBackup() {
  addBackupTask(props.serverId, doCreate);
}

function openRestoreDialog(name: string) {
  selectedBackup.value = name;
  showRestoreDialog.value = true;
}

function openDeleteDialog(name: string) {
  selectedBackup.value = name;
  showDeleteDialog.value = true;
}

async function onRestoreConfirmed() {
  await doRestore(selectedBackup.value);
}

async function onDeleteConfirmed() {
  await doDelete(selectedBackup.value);
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

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
