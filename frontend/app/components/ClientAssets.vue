<template>
  <div class="flex flex-col flex-1 min-h-0">
    <div class="space-y-2 flex-1 min-h-0 overflow-y-auto">
      <div
        v-if="filteredAssets.length === 0 && !loading"
        class="text-center text-gray-500 dark:text-neutral-400 py-6 text-sm"
      >
        No files.
      </div>
      <div v-else-if="loading" class="py-6 flex justify-center">
        <Loader2 class="w-5 h-5 animate-spin text-primary" />
      </div>
      <div v-else class="space-y-1">
        <div
          v-for="asset in filteredAssets"
          :key="asset.filename"
          class="flex items-center justify-between px-3 py-2 rounded-lg bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700"
        >
          <div class="flex items-center gap-2 min-w-0">
            <FileArchive class="w-4 h-4 text-gray-400 flex-shrink-0" />
            <span class="text-sm text-gray-900 dark:text-white truncate">{{
              asset.filename
            }}</span>
            <span class="text-xs text-gray-400 flex-shrink-0">{{
              formatBytes(asset.size)
            }}</span>
          </div>
          <div class="flex items-center gap-1 flex-shrink-0">
            <button
              class="p-1 rounded hover:bg-gray-100 dark:hover:bg-neutral-700 text-gray-500 dark:text-neutral-400"
              :title="asset.enabled ? 'Disable' : 'Enable'"
              @click="toggleAsset(asset.filename)"
            >
              <Eye v-if="asset.enabled" class="w-4 h-4" />
              <EyeOff v-else class="w-4 h-4" />
            </button>
            <button
              class="p-1 rounded hover:bg-red-500/10 text-red-500"
              @click="deleteAsset(asset.filename)"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Loader2,
  FileArchive,
  Eye,
  EyeOff,
  Trash2,
  type Component,
} from "lucide-vue-next";
import { useClientAssetsRefresh } from "~/composables/useClientAssetsRefresh";

interface Props {
  serverId: string;
  type: "resourcepacks" | "shaderpacks";
  title: string;
  icon: Component;
  searchQuery?: string;
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: "",
});

const filteredAssets = computed(() => {
  const q = props.searchQuery.trim().toLowerCase();
  if (!q) return assets.value;
  return assets.value.filter((a) => a.filename.toLowerCase().includes(q));
});

const { assets, loading, refresh, deleteAsset, toggleAsset } = useClientAssets(
  props.serverId,
  props.type,
);

const { key: refreshKey } = useClientAssetsRefresh(props.serverId);
watch(refreshKey, () => {
  refresh();
});

onMounted(() => {
  refresh();
});

function formatBytes(n: number): string {
  if (n >= 1024 * 1024) return (n / (1024 * 1024)).toFixed(1) + " MB";
  if (n >= 1024) return (n / 1024).toFixed(1) + " KB";
  return n + " B";
}
</script>
