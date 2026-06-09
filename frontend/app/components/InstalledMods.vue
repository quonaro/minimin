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
        <span class="text-xs">Drag & drop .jar or .zip here, or use the button below.</span>
      </div>

      <div
        v-for="mod in mods"
        :key="mod.filename"
        class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 hover:border-gray-300 dark:hover:border-neutral-600 transition-colors"
      >
        <div
          class="w-10 h-10 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400 shrink-0"
        >
          <Box class="w-5 h-5" />
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold text-gray-900 dark:text-white truncate">
            {{ mod.name || mod.filename }}
          </p>
          <p class="text-xs text-gray-500 dark:text-neutral-400 truncate">
            {{ mod.version }} &middot; {{ mod.modid }}
            <span v-if="mod.authors">&middot; {{ mod.authors }}</span>
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

    <div class="mt-4 pt-4 border-t border-gray-200 dark:border-neutral-700">
      <input
        ref="fileInput"
        type="file"
        accept=".jar,.zip"
        class="hidden"
        @change="onFileSelect"
      />
      <button
        class="w-full inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-primary hover:bg-primary/90 text-white text-sm font-medium transition-colors disabled:opacity-50"
        :disabled="uploadLoading"
        @click="fileInput?.click()"
      >
        <Upload class="w-4 h-4" />
        {{ uploadLoading ? "Uploading..." : "Upload .jar or .zip" }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Box, Trash2, Upload } from "lucide-vue-next";
import type { ModInfo } from "~/composables/useMods";

interface Props {
  mods: ModInfo[];
  loading: boolean;
  uploadLoading: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  delete: [filename: string];
  upload: [file: File];
}>();

const fileInput = ref<HTMLInputElement | null>(null);

function onFileSelect(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  emit("upload", file);
  input.value = "";
}

function onDrop(e: DragEvent) {
  const file = e.dataTransfer?.files[0];
  if (!file) return;
  const ext = file.name.split(".").pop()?.toLowerCase();
  if (ext !== "jar" && ext !== "zip") {
    useToast().show("error", "Invalid file type", { description: "Only .jar and .zip files are allowed." });
    return;
  }
  emit("upload", file);
}
</script>
