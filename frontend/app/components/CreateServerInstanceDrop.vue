<script setup lang="ts">
import { UploadCloud } from "lucide-vue-next";

export interface World {
  name: string;
  archivePath: string;
}

export interface InstancePreview {
  token: string;
  format: string;
  instanceName?: string;
  gameVersion?: string;
  engineType?: string;
  loaderVersion?: string;
  detectedPaths: string[];
  worlds: World[];
}

const emit = defineEmits<{
  (e: "preview", preview: InstancePreview): void;
  (e: "error", message: string): void;
}>();

const { show: showToast } = useToast();

const isDragOver = ref(false);
const uploading = ref(false);
const percentage = ref(0);

const statusText = computed(() => {
  if (percentage.value === 0) return "Uploading…";
  if (percentage.value < 100) return `${percentage.value}% uploaded`;
  return "Analyzing archive…";
});

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
  const file = e.dataTransfer?.files?.[0];
  if (!file) return;
  handleFile(file);
}

function onFileInput(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  handleFile(file);
  input.value = "";
}

function handleFile(file: File) {
  const ext = file.name.split(".").pop()?.toLowerCase();
  if (!ext || !["zip", "mrpack"].includes(ext)) {
    showToast("error", "Unsupported file", {
      description: "Please drop a .zip or .mrpack archive.",
    });
    emit("error", "unsupported file type");
    return;
  }

  uploading.value = true;
  percentage.value = 0;
  const fd = new FormData();
  fd.append("file", file);

  const xhr = new XMLHttpRequest();
  xhr.upload.addEventListener("progress", (e) => {
    if (!e.lengthComputable) return;
    percentage.value = Math.min(100, Math.round((e.loaded / e.total) * 100));
  });
  xhr.addEventListener("load", () => {
    uploading.value = false;
    if (xhr.status >= 200 && xhr.status < 300) {
      try {
        const res: InstancePreview = JSON.parse(xhr.responseText);
        emit("preview", res);
      } catch (err) {
        showToast("error", "Invalid preview response", {
          description: "Server returned malformed JSON.",
        });
        emit("error", "invalid response");
      }
    } else {
      let detail = `Upload failed: ${xhr.status} ${xhr.statusText}`;
      try {
        const parsed = JSON.parse(xhr.responseText);
        if (parsed && parsed.detail) detail = parsed.detail;
      } catch {
        /* ignore */
      }
      showToast("error", "Archive preview failed", { description: detail });
      emit("error", detail);
    }
  });
  xhr.addEventListener("error", () => {
    uploading.value = false;
    const msg = "Network error during upload";
    showToast("error", "Archive preview failed", { description: msg });
    emit("error", msg);
  });
  xhr.addEventListener("abort", () => {
    uploading.value = false;
    const msg = "Upload cancelled";
    showToast("error", "Archive preview failed", { description: msg });
    emit("error", msg);
  });

  xhr.open("POST", `${useApiBase()}/servers/prepare-instance`);
  xhr.withCredentials = true;
  xhr.send(fd);
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
      <div v-if="uploading" class="mt-4">
        <div
          class="h-2 w-full max-w-xs mx-auto rounded-full bg-gray-200 dark:bg-neutral-600 overflow-hidden"
        >
          <div
            class="h-full bg-primary transition-all duration-150"
            :style="{ width: `${percentage}%` }"
          />
        </div>
        <p class="text-xs text-primary mt-2 font-medium">{{ statusText }}</p>
      </div>
    </div>
  </div>
</template>
