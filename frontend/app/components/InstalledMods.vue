<template>
  <div
    class="flex flex-col h-full"
    :class="
      isDragOver
        ? 'border-2 border-dashed border-primary rounded-xl bg-primary/5'
        : ''
    "
    @dragenter.prevent="isDragOver = true"
    @dragover.prevent="isDragOver = true"
    @dragleave="isDragOver = false"
    @drop.prevent="onDrop"
  >
    <div
      v-if="filteredMods.length === 0 && !loading"
      class="flex-1 flex flex-col items-center justify-center text-center"
    >
      <Box class="w-12 h-12 text-indigo-500 mb-3" />
      <p class="text-gray-900 dark:text-white font-medium text-base">
        No mods installed
      </p>
      <p class="text-xs text-gray-500 dark:text-neutral-400 mt-1 max-w-xs">
        Drag & drop .jar or .zip files here, or use the Upload button above.
      </p>
    </div>

    <div
      v-else
      class="relative flex flex-col flex-1 min-h-0 overflow-y-auto pr-1"
    >
      <Transition
        enter-active-class="transition ease-out duration-200"
        enter-from-class="opacity-0 translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition ease-in duration-150"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 translate-y-2"
      >
        <div
          v-if="selected.size > 0"
          class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 px-4 py-2 rounded-2xl border shadow-2xl backdrop-blur-sm bg-white/95 dark:bg-neutral-800/95 border-gray-200 dark:border-neutral-700"
        >
          <span
            class="text-xs font-medium text-gray-700 dark:text-neutral-200 whitespace-nowrap"
          >
            {{ selected.size }} selected
          </span>
          <div class="w-px h-4 bg-gray-200 dark:bg-neutral-600" />
          <button
            class="text-xs font-medium text-primary hover:underline whitespace-nowrap"
            @click="selectAll"
          >
            All
          </button>
          <button
            class="text-xs font-medium text-gray-500 dark:text-neutral-400 hover:underline whitespace-nowrap"
            @click="clearSelection"
          >
            Clear
          </button>
          <div class="w-px h-4 bg-gray-200 dark:bg-neutral-600" />
          <button
            class="inline-flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-medium bg-gray-100 dark:bg-neutral-700 text-gray-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-600 transition-colors"
            @click="emitBatchToggle"
          >
            <EyeOff class="w-3.5 h-3.5" />
            Toggle
          </button>
          <button
            class="inline-flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-medium bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 hover:bg-red-200 dark:hover:bg-red-900/50 transition-colors"
            @click="emitBatchDelete"
          >
            <Trash2 class="w-3.5 h-3.5" />
            Delete
          </button>
        </div>
      </Transition>
      <div
        class="grid grid-cols-4 lg:grid-cols-6 xl:grid-cols-8 2xl:grid-cols-10 gap-2 content-start p-2"
      >
        <div
          v-for="mod in filteredMods"
          :key="mod.filename"
          class="relative flex flex-col p-2 rounded-xl border transition-all cursor-pointer"
          :class="
            selected.has(mod.filename)
              ? 'ring-2 ring-primary ring-offset-1 dark:ring-offset-neutral-800'
              : mod.corrupted
                ? 'bg-red-50 dark:bg-red-900/10 border-red-200 dark:border-red-800 hover:border-red-300 dark:hover:border-red-700'
                : mod.enabled !== false
                  ? 'bg-white dark:bg-neutral-800 border-gray-200 dark:border-neutral-700 hover:border-gray-300 dark:hover:border-neutral-600'
                  : 'bg-gray-100 dark:bg-neutral-800 border-gray-200 dark:border-neutral-700 opacity-60'
          "
          :title="mod.description || ''"
          @click="toggleSelect(mod.filename)"
          @contextmenu.prevent="openContextMenu(mod, $event)"
        >
          <input
            type="checkbox"
            class="absolute top-2 right-2 z-10 w-4 h-4 rounded border-gray-300 dark:border-neutral-600 text-primary focus:ring-primary cursor-pointer bg-white/80 dark:bg-neutral-800/80 backdrop-blur"
            :checked="selected.has(mod.filename)"
            @click.stop
            @change="toggleSelect(mod.filename)"
          />
          <div
            class="self-center relative w-10 h-10 rounded-xl overflow-hidden"
            :class="
              mod.corrupted
                ? 'bg-red-100 dark:bg-red-900/30'
                : 'bg-indigo-50 dark:bg-indigo-900/20'
            "
          >
            <img
              v-if="serverId && !mod.corrupted"
              v-show="iconLoaded[mod.filename]"
              :src="getIconUrl(mod.filename)"
              class="w-full h-full object-cover"
              alt=""
              @load="iconLoaded[mod.filename] = true"
              @error="iconLoaded[mod.filename] = false"
            />
            <div
              v-show="mod.corrupted || !iconLoaded[mod.filename]"
              class="absolute inset-0 flex items-center justify-center"
              :class="mod.corrupted ? 'text-red-500' : 'text-indigo-500'"
            >
              <AlertTriangle v-if="mod.corrupted" class="w-5 h-5" />
              <Box v-else class="w-5 h-5" />
            </div>
          </div>
          <div class="mt-1 flex items-center justify-center gap-1 flex-wrap">
            <template v-if="mod.corrupted">
              <span
                class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400"
                >Corrupted JAR</span
              >
            </template>
            <template v-else>
              <span
                v-if="mod.version"
                class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400 truncate max-w-[4.5rem]"
              >
                {{ mod.version }}
              </span>
              <span
                v-if="mod.size"
                class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-neutral-400"
              >
                {{ formatBytes(mod.size) }}
              </span>
            </template>
          </div>
          <div class="mt-0.5 text-center min-w-0">
            <p
              class="text-xs font-semibold truncate"
              :class="
                mod.corrupted
                  ? 'text-red-700 dark:text-red-300'
                  : 'text-gray-900 dark:text-white'
              "
            >
              {{ mod.name || mod.filename }}
            </p>
          </div>
          <div class="mt-1.5 flex items-center justify-center gap-0.5">
            <button
              class="text-gray-400 hover:text-primary transition-colors p-1"
              title="Show in File Manager"
              @click="emit('show-in-files', mod.filename)"
            >
              <FolderOpen class="w-3.5 h-3.5" />
            </button>
            <button
              class="text-gray-400 hover:text-primary transition-colors p-1"
              :title="mod.enabled !== false ? 'Disable mod' : 'Enable mod'"
              @click="emit('toggle', mod.filename)"
            >
              <Eye v-if="mod.enabled !== false" class="w-3.5 h-3.5" />
              <EyeOff v-else class="w-3.5 h-3.5" />
            </button>
            <button
              class="text-gray-400 hover:text-primary transition-colors p-1"
              title="Move to client mods"
              @click="emit('move', mod.filename, 'client')"
            >
              <Server class="w-3.5 h-3.5" />
            </button>
            <button
              class="text-gray-400 hover:text-red-500 transition-colors p-1"
              :disabled="loading"
              @click="emit('delete', mod.filename)"
            >
              <Trash2 class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <div
          v-if="contextMenuOpen"
          class="fixed inset-0 z-40"
          @click="closeContextMenu"
          @contextmenu.prevent
        >
          <div
            data-mods-context-menu="true"
            class="fixed z-50 min-w-[210px] bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-xl shadow-xl py-1"
            :style="{ left: `${contextMenuX}px`, top: `${contextMenuY}px` }"
            @click.stop
          >
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700 flex items-center gap-2"
              @click="contextShowInFiles"
            >
              <FolderOpen class="w-4 h-4" />
              Show in File Manager
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700 flex items-center gap-2"
              @click="contextToggle"
            >
              <component
                :is="contextTarget?.enabled !== false ? EyeOff : Eye"
                class="w-4 h-4"
              />
              {{ contextTarget?.enabled !== false ? "Disable" : "Enable" }}
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700 flex items-center gap-2"
              @click="contextMove"
            >
              <Server class="w-4 h-4" />
              Move to client mods
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700 flex items-center gap-2"
              @click="contextCopy"
            >
              <Copy class="w-4 h-4" />
              Copy to client mods
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 flex items-center gap-2"
              @click="contextDelete"
            >
              <Trash2 class="w-4 h-4" />
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Box,
  Trash2,
  Eye,
  EyeOff,
  Server,
  Copy,
  AlertTriangle,
  FolderOpen,
} from "lucide-vue-next";
import type { ModInfo } from "~/composables/useMods";

interface Props {
  mods: ModInfo[];
  loading: boolean;
  serverId?: string;
  searchQuery?: string;
  sideFilter?: "all" | "server" | "client";
  sortBy?:
    | "name-asc"
    | "name-desc"
    | "size-asc"
    | "size-desc"
    | "date-asc"
    | "date-desc";
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: "",
  sideFilter: "all",
  sortBy: "name-asc",
});
const emit = defineEmits<{
  delete: [filename: string];
  "batch-delete": [filenames: string[]];
  upload: [file: File];
  toggle: [filename: string];
  "batch-toggle": [filenames: string[]];
  move: [filename: string, target: "server" | "client"];
  copy: [
    filename: string,
    source: "server" | "client",
    target: "server" | "client",
  ];
  "show-in-files": [filename: string];
}>();

const isDragOver = ref(false);

const contextMenuOpen = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);
const contextTarget = ref<ModInfo | null>(null);

function openContextMenu(mod: ModInfo, event: MouseEvent) {
  contextTarget.value = mod;
  contextMenuX.value = event.clientX;
  contextMenuY.value = event.clientY;
  contextMenuOpen.value = true;
}

function closeContextMenu() {
  contextMenuOpen.value = false;
  contextTarget.value = null;
}

function contextToggle() {
  if (contextTarget.value) {
    emit("toggle", contextTarget.value.filename);
  }
  closeContextMenu();
}

function contextMove() {
  if (contextTarget.value) {
    emit("move", contextTarget.value.filename, "client");
  }
  closeContextMenu();
}

function contextCopy() {
  if (contextTarget.value) {
    emit("copy", contextTarget.value.filename, "server", "client");
  }
  closeContextMenu();
}

function contextShowInFiles() {
  if (contextTarget.value) {
    emit("show-in-files", contextTarget.value.filename);
  }
  closeContextMenu();
}

function contextDelete() {
  if (contextTarget.value) {
    emit("delete", contextTarget.value.filename);
  }
  closeContextMenu();
}

const selected = ref<Set<string>>(new Set());

function toggleSelect(filename: string) {
  const s = selected.value;
  if (s.has(filename)) {
    s.delete(filename);
  } else {
    s.add(filename);
  }
  selected.value = new Set(s);
}

function selectAll() {
  selected.value = new Set(filteredMods.value.map((m) => m.filename));
}

function clearSelection() {
  selected.value = new Set();
}

function emitBatchDelete() {
  if (selected.value.size === 0) return;
  emit("batch-delete", [...selected.value]);
  selected.value = new Set();
}

function emitBatchToggle() {
  if (selected.value.size === 0) return;
  emit("batch-toggle", [...selected.value]);
  selected.value = new Set();
}

const filteredMods = computed(() => {
  let list = props.mods;
  const q = props.searchQuery.trim().toLowerCase();
  if (q) {
    list = list.filter(
      (m) =>
        (m.name || "").toLowerCase().includes(q) ||
        (m.filename || "").toLowerCase().includes(q) ||
        (m.modid || "").toLowerCase().includes(q),
    );
  }
  if (props.sideFilter === "server") {
    list = list.filter((m) => m.environment === "server");
  } else if (props.sideFilter === "client") {
    list = list.filter((m) => m.environment === "client");
  }
  list = [...list];
  switch (props.sortBy) {
    case "name-asc":
      list.sort((a, b) =>
        (a.name || a.filename).localeCompare(b.name || b.filename),
      );
      break;
    case "name-desc":
      list.sort((a, b) =>
        (b.name || b.filename).localeCompare(a.name || a.filename),
      );
      break;
    case "size-asc":
      list.sort((a, b) => (a.size ?? 0) - (b.size ?? 0));
      break;
    case "size-desc":
      list.sort((a, b) => (b.size ?? 0) - (a.size ?? 0));
      break;
    case "date-asc":
      list.sort((a, b) => (a.installed_at ?? 0) - (b.installed_at ?? 0));
      break;
    case "date-desc":
      list.sort((a, b) => (b.installed_at ?? 0) - (a.installed_at ?? 0));
      break;
  }
  return list;
});

const iconLoaded = ref<Record<string, boolean>>({});

function getIconUrl(filename: string): string {
  if (!props.serverId) return "";
  return `${useApiBase()}/servers/${props.serverId}/mods/${encodeURIComponent(filename)}/icon`;
}

function formatBytes(n: number): string {
  if (n >= 1024 * 1024 * 1024)
    return (n / (1024 * 1024 * 1024)).toFixed(1) + " GB";
  if (n >= 1024 * 1024) return (n / (1024 * 1024)).toFixed(1) + " MB";
  if (n >= 1024) return (n / 1024).toFixed(1) + " KB";
  return n + " B";
}

function onDrop(e: DragEvent) {
  isDragOver.value = false;
  const dt = e.dataTransfer;
  if (!dt) return;

  const internalFilename = dt.getData("application/x-mod-filename");
  const internalSource = dt.getData("application/x-mod-source");
  if (internalFilename && internalSource === "client") {
    if (e.ctrlKey) {
      emit("copy", internalFilename, "client", "server");
    } else {
      emit("move", internalFilename, "server");
    }
    return;
  }

  const files = [...dt.files].filter((f) => {
    const ext = f.name.split(".").pop()?.toLowerCase();
    return ext === "jar" || ext === "zip";
  });
  if (files.length === 0 && dt.files.length > 0) {
    useToast().show("error", "Invalid file type", {
      description: "Only .jar and .zip files are allowed.",
    });
    return;
  }
  for (const file of files) {
    emit("upload", file);
  }
}
</script>
