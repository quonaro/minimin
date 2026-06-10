<template>
  <div class="flex flex-col h-full">
    <div class="mb-4 space-y-2">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-bold text-gray-900 dark:text-white">
          Server Mods
        </h2>
        <span class="text-xs text-gray-500 dark:text-neutral-400">
          {{ filteredMods.length }} mods
        </span>
      </div>
      <div class="flex items-center gap-2">
        <input
          :value="searchQuery"
          type="text"
          placeholder="Search server mods..."
          class="flex-1 px-3 py-1.5 rounded-lg bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 text-sm text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-neutral-500 focus:ring-2 focus:ring-primary focus:outline-none"
          @input="onSearchInput"
        />
        <div
          class="flex rounded-lg border border-gray-300 dark:border-neutral-700 overflow-hidden"
        >
          <button
            v-for="opt in sideOptions"
            :key="opt.value"
            class="px-2 py-1.5 text-xs font-medium transition-colors"
            :class="
              sideFilter === opt.value
                ? 'bg-primary text-white'
                : 'bg-white dark:bg-neutral-800 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700'
            "
            @click="emit('update:sideFilter', opt.value)"
          >
            {{ opt.label }}
          </button>
        </div>
      </div>
    </div>

    <div
      class="flex-1 overflow-y-auto space-y-2 pr-1"
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
        class="text-center text-gray-500 dark:text-neutral-400 py-12 text-sm"
      >
        No mods installed.
        <br />
        <span class="text-xs"
          >Drag & drop .jar or .zip here, or use the button below.</span
        >
      </div>

      <div
        v-for="mod in filteredMods"
        :key="mod.filename"
        class="flex items-start gap-3 p-3 rounded-xl border transition-colors"
        :class="
          mod.enabled !== false
            ? 'bg-gray-50 dark:bg-neutral-700/50 border-gray-100 dark:border-neutral-700 hover:border-gray-300 dark:hover:border-neutral-600'
            : 'bg-gray-100 dark:bg-neutral-800 border-gray-200 dark:border-neutral-700 opacity-60'
        "
        draggable="true"
        @dragstart="onDragStart($event, mod)"
        @contextmenu.prevent="openContextMenu(mod, $event)"
      >
        <div
          class="relative w-10 h-10 shrink-0 rounded-lg overflow-hidden bg-gray-200 dark:bg-neutral-600"
        >
          <img
            v-if="serverId"
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
        <div class="flex items-center gap-1 shrink-0">
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
import { Box, Trash2, Eye, EyeOff, Server, Copy } from "lucide-vue-next";
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
  "update:searchQuery": [value: string];
  "update:sideFilter": [value: "all" | "server" | "client"];
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

function contextDelete() {
  if (contextTarget.value) {
    emit("delete", contextTarget.value.filename);
  }
  closeContextMenu();
}

const sideOptions = [
  { label: "All", value: "all" as const },
  { label: "Server", value: "server" as const },
  { label: "Client", value: "client" as const },
];

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

function onSearchInput(e: Event) {
  emit("update:searchQuery", (e.target as HTMLInputElement).value);
}

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

  const file = dt.files[0];
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
