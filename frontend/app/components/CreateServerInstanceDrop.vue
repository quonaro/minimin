<script setup lang="ts">
import { UploadCloud } from "lucide-vue-next";

export interface InstancePreview {
  token: string;
  format: string;
  instanceName?: string;
  gameVersion?: string;
  engineType?: string;
  loaderVersion?: string;
  detectedPaths: string[];
}

const emit = defineEmits<{
  (e: "preview", preview: InstancePreview): void;
  (e: "error", message: string): void;
}>();

const { show: showToast } = useToast();

const isDragOver = ref(false);
const previewing = ref(false);

function onDragOver(e: DragEvent) {
  e.preventDefault();
  isDragOver.value = true;
}

function onDragLeave(e: DragEvent) {
  e.preventDefault();
  isDragOver.value = false;
}

function onDrop(e: DragEvent) {
  e.preventDefault();
  isDragOver.value = false;
  const files = e.dataTransfer?.files;
  if (!files || files.length === 0) return;
  handleFile(files[0]);
}

function onFileInput(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  handleFile(file);
  input.value = "";
}

async function handleFile(file: File) {
  const ext = file.name.split(".").pop()?.toLowerCase();
  if (!ext || !["zip", "mrpack"].includes(ext)) {
    showToast("error", "Unsupported file", {
      description: "Please drop a .zip or .mrpack archive.",
    });
    emit("error", "unsupported file type");
    return;
  }

  previewing.value = true;
  try {
    const fd = new FormData();
    fd.append("file", file);
    const res = await $fetch<InstancePreview>("/servers/prepare-instance", {
      baseURL: useApiBase(),
      method: "POST",
      credentials: "include",
      body: fd,
    });
    emit("preview", res);
  } catch (err: any) {
    const msg = err?.data?.detail || err?.message || "Failed to parse archive";
    showToast("error", "Archive preview failed", { description: msg });
    emit("error", msg);
  } finally {
    previewing.value = false;
  }
}
</script>

<template>
  <div
    class="relative rounded-xl border-2 border-dashed transition-colors p-6 text-center"
    :class="
      isDragOver
        ? 'border-primary bg-primary/5'
        : 'border-gray-300 dark:border-neutral-600 bg-gray-50 dark:bg-neutral-700/30'
    "
    @dragover.prevent="onDragOver"
    @dragleave.prevent="onDragLeave"
    @drop.prevent="onDrop"
  >
    <input
      id="instance-file"
      type="file"
      accept=".zip,.mrpack"
      class="hidden"
      @change="onFileInput"
    />
    <label
      for="instance-file"
      class="absolute inset-0 cursor-pointer"
      aria-label="Select instance archive"
    />
    <div class="relative pointer-events-none">
      <UploadCloud
        class="w-10 h-10 mx-auto mb-3"
        :class="
          isDragOver ? 'text-primary' : 'text-gray-400 dark:text-neutral-400'
        "
      />
      <p class="text-sm font-medium text-gray-900 dark:text-white">
        Drop a Prism instance, Modrinth .mrpack, CurseForge modpack, or plain
        .zip here
      </p>
      <p class="text-xs text-gray-500 dark:text-neutral-400 mt-1">
        Should contain mods/ and optionally resourcepacks/, shaderpacks/
      </p>
      <p
        v-if="previewing"
        class="text-xs text-primary mt-3 font-medium animate-pulse"
      >
        Analyzing archive…
      </p>
    </div>
  </div>
</template>
