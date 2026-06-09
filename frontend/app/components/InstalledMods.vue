<template>
  <div class="flex flex-col h-full">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-bold text-gray-900 dark:text-white">
        Installed Mods
      </h2>
      <span class="text-xs text-gray-500 dark:text-neutral-400">
        {{ mods.length }} mods
      </span>
    </div>

    <div
      class="flex-1 overflow-y-auto space-y-2 pr-1"
      @dragover.prevent
      @drop.prevent="onDrop"
    >
      <div
        v-if="mods.length === 0 && !loading"
        class="text-center text-gray-500 dark:text-neutral-400 py-12 text-sm"
      >
        No mods installed.
        <br />
        <span class="text-xs"
          >Drag & drop .jar or .zip here, or use the button below.</span
        >
      </div>

      <div
        v-for="mod in mods"
        :key="mod.filename"
        class="flex items-start gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 hover:border-gray-300 dark:hover:border-neutral-600 transition-colors"
      >
        <div
          class="relative w-10 h-10 shrink-0 rounded-lg overflow-hidden bg-gray-200 dark:bg-neutral-600"
        >
          <img
            v-if="agentId && serverId"
            v-show="iconLoaded[mod.filename]"
            :src="getIconUrl(mod.filename)"
            class="w-full h-full object-cover"
            alt=""
            @load="iconLoaded[mod.filename] = true"
            @error="iconLoaded[mod.filename] = false"
          />
          <div
            v-show="!iconLoaded[mod.filename]"
            class="absolute inset-0 flex items-center justify-center bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400"
          >
            <Box class="w-5 h-5" />
          </div>
        </div>
        <div class="flex-1 min-w-0">
          <p
            class="text-sm font-semibold text-gray-900 dark:text-white truncate"
          >
            {{ mod.name || mod.filename }}
          </p>
          <p class="text-xs text-gray-500 dark:text-neutral-400 truncate">
            {{ mod.version }} &middot; {{ mod.modid }}
            <span v-if="mod.authors">&middot; {{ mod.authors }}</span>
            <span v-if="mod.size">&middot; {{ formatBytes(mod.size) }}</span>
          </p>
          <p
            v-if="mod.description"
            class="text-xs text-gray-500 dark:text-neutral-400 line-clamp-2 mt-0.5"
          >
            {{ mod.description }}
          </p>
        </div>
        <button
          class="text-gray-400 hover:text-red-500 transition-colors shrink-0 p-1"
          :disabled="loading"
          @click="emit('delete', mod.filename)"
        >
          <Trash2 class="w-4 h-4" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Box, Trash2 } from "lucide-vue-next";
import type { ModInfo } from "~/composables/useMods";

interface Props {
  mods: ModInfo[];
  loading: boolean;
  agentId?: string;
  serverId?: string;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  delete: [filename: string];
  upload: [file: File];
}>();

const iconLoaded = ref<Record<string, boolean>>({});

function getIconUrl(filename: string): string {
  if (!props.agentId || !props.serverId) return "";
  return `${useApiBase()}/agent/${props.agentId}/servers/${props.serverId}/mods/${encodeURIComponent(filename)}/icon`;
}

function formatBytes(n: number): string {
  if (n >= 1024 * 1024 * 1024)
    return (n / (1024 * 1024 * 1024)).toFixed(1) + " GB";
  if (n >= 1024 * 1024) return (n / (1024 * 1024)).toFixed(1) + " MB";
  if (n >= 1024) return (n / 1024).toFixed(1) + " KB";
  return n + " B";
}

function onDrop(e: DragEvent) {
  const file = e.dataTransfer?.files[0];
  if (!file) return;
  const ext = file.name.split(".").pop()?.toLowerCase();
  if (ext !== "jar" && ext !== "zip") {
    useToast().show("error", "Invalid file type", {
      description: "Only .jar and .zip files are allowed.",
    });
    return;
  }
  emit("upload", file);
}
</script>
