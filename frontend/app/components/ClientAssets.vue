<template>
  <div class="flex flex-col flex-1 min-h-0">
    <div
      v-if="filteredAssets.length === 0 && !loading"
      class="flex-1 flex flex-col items-center justify-center text-center"
    >
      <component
        :is="icon"
        class="w-12 h-12 mb-3"
        :class="
          type === 'resourcepacks' ? 'text-emerald-500' : 'text-purple-500'
        "
      />
      <p class="text-gray-900 dark:text-white font-medium text-base">
        {{ title }}
      </p>
      <p class="text-xs text-gray-500 dark:text-neutral-400 mt-1 max-w-xs">
        No
        {{ type === "resourcepacks" ? "resource packs" : "shader packs" }}
        installed.
      </p>
    </div>
    <div v-else-if="loading" class="flex-1 flex items-center justify-center">
      <Loader2 class="w-5 h-5 animate-spin text-primary" />
    </div>
    <div
      v-else
      class="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 content-start flex-1 min-h-0 overflow-y-auto p-2"
    >
      <div
        v-for="asset in filteredAssets"
        :key="asset.filename"
        class="flex flex-col p-4 rounded-xl border bg-white dark:bg-neutral-800 border-gray-200 dark:border-neutral-700 hover:border-gray-300 dark:hover:border-neutral-600 transition-colors"
      >
        <div
          class="self-center w-14 h-14 rounded-xl flex items-center justify-center"
          :class="
            type === 'resourcepacks'
              ? 'bg-emerald-50 dark:bg-emerald-900/20 text-emerald-500'
              : 'bg-purple-50 dark:bg-purple-900/20 text-purple-500'
          "
        >
          <component :is="icon" class="w-7 h-7" />
        </div>
        <div class="mt-3 text-center min-w-0">
          <p
            class="text-sm font-semibold truncate text-gray-900 dark:text-white"
          >
            {{ asset.filename }}
          </p>
          <p class="text-[11px] text-gray-400 dark:text-neutral-500 mt-1">
            {{ formatBytes(asset.size) }}
          </p>
        </div>
        <div class="mt-3 flex items-center justify-center gap-1">
          <button
            class="text-gray-400 hover:text-primary transition-colors p-1"
            title="Show in File Manager"
            @click="emit('show-in-files', asset.filename)"
          >
            <FolderOpen class="w-4 h-4" />
          </button>
          <button
            class="text-gray-400 hover:text-primary transition-colors p-1"
            :title="asset.enabled ? 'Disable' : 'Enable'"
            @click="toggleAsset(asset.filename)"
          >
            <Eye v-if="asset.enabled" class="w-4 h-4" />
            <EyeOff v-else class="w-4 h-4" />
          </button>
          <button
            class="text-gray-400 hover:text-red-500 transition-colors p-1"
            @click="deleteAsset(asset.filename)"
          >
            <Trash2 class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Loader2,
  Eye,
  EyeOff,
  Trash2,
  FolderOpen,
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
const emit = defineEmits<{
  "show-in-files": [filename: string];
}>();

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
