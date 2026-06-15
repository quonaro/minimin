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
      class="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 content-start flex-1 min-h-0 overflow-y-auto pr-1 p-2"
    >
      <div
        v-for="mod in filteredMods"
        :key="mod.filename"
        class="flex flex-col p-4 rounded-xl border transition-colors"
        :class="
          mod.corrupted
            ? 'bg-red-50 dark:bg-red-900/10 border-red-200 dark:border-red-800 hover:border-red-300 dark:hover:border-red-700'
            : mod.enabled !== false
              ? 'bg-white dark:bg-neutral-800 border-gray-200 dark:border-neutral-700 hover:border-gray-300 dark:hover:border-neutral-600'
              : 'bg-gray-100 dark:bg-neutral-800 border-gray-200 dark:border-neutral-700 opacity-60'
        "
        draggable="true"
        @dragstart="onDragStart($event, mod)"
        @contextmenu.prevent="openContextMenu(mod, $event)"
      >
        <div
          class="self-center relative w-14 h-14 rounded-xl overflow-hidden"
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
            <AlertTriangle v-if="mod.corrupted" class="w-6 h-6" />
            <Box v-else class="w-6 h-6" />
          </div>
        </div>
        <div class="mt-3 text-center min-w-0">
          <p
            class="text-sm font-semibold truncate"
            :class="
              mod.corrupted
                ? 'text-red-700 dark:text-red-300'
                : 'text-gray-900 dark:text-white'
            "
          >
            {{ mod.name || mod.filename }}
          </p>
          <p
            class="text-xs truncate mt-0.5"
            :class="
              mod.corrupted
                ? 'text-red-500 dark:text-red-400'
                : 'text-gray-500 dark:text-neutral-400'
            "
          >
            <template v-if="mod.corrupted"> Corrupted JAR </template>
            <template v-else>
              {{ mod.version }} &middot; {{ mod.modid }}
              <span v-if="mod.authors">&middot; {{ mod.authors }}</span>
            </template>
          </p>
          <p
            v-if="mod.size"
            class="text-[11px] text-gray-400 dark:text-neutral-500 mt-0.5"
          >
            {{ formatBytes(mod.size) }}
          </p>
          <p
            v-if="mod.description && !mod.corrupted"
            class="text-xs text-gray-400 dark:text-neutral-500 line-clamp-2 mt-1"
          >
            {{ mod.description }}
          </p>
        </div>
        <div class="mt-3 flex items-center justify-center gap-1">
          <button
            class="text-gray-400 hover:text-primary transition-colors p-1"
            title="Show in File Manager"
            @click="emit('show-in-files', mod.filename)"
          >
            <FolderOpen class="w-4 h-4" />
          </button>
          <button
            class="text-gray-400 hover:text-primary transition-colors p-1"
            :title="mod.enabled !== false ? 'Disable mod' : 'Enable mod'"
            @click="emit('toggle', mod.filename)"
          >
            <Eye v-if="mod.enabled !== false" class="w-4 h-4" />
            <EyeOff v-else class="w-4 h-4" />
          </button>
          <button
            class="text-gray-400 hover:text-primary transition-colors p-1"
            title="Move to client mods"
            @click="emit('move', mod.filename, 'client')"
          >
            <Server class="w-4 h-4" />
          </button>
          <button
            class="text-gray-400 hover:text-red-500 transition-colors p-1"
            :disabled="loading"
            @click="emit('delete', mod.filename)"
          >
            <Trash2 class="w-4 h-4" />
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
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: "",
  sideFilter: "all",
});
const emit = defineEmits<{
  delete: [filename: string];
  upload: [file: File];
  toggle: [filename: string];
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

function onDragStart(e: DragEvent, mod: ModInfo) {
  e.dataTransfer?.setData("application/x-mod-filename", mod.filename);
  e.dataTransfer?.setData("application/x-mod-source", "server");
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
