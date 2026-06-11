<template>
  <div class="flex flex-col h-full overflow-hidden">
    <div class="flex items-center justify-between mb-2">
      <h2 class="text-lg font-bold text-gray-900 dark:text-white">Client</h2>
      <span class="text-xs text-gray-500 dark:text-neutral-400">
        {{
          filteredMods.length +
          assetCounts.resourcepacks +
          assetCounts.shaderpacks
        }}
        items
      </span>
    </div>
    <div class="mb-4 flex items-center gap-2">
      <input
        :value="rawQuery"
        type="text"
        placeholder="Search client mods, resource packs, shader packs..."
        class="flex-1 px-3 py-1.5 rounded-lg bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 text-sm text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-neutral-500 focus:ring-2 focus:ring-primary focus:outline-none"
        @input="onSearchInput"
      />
      <button
        class="inline-flex items-center gap-1 px-2 py-1.5 rounded-lg bg-primary/10 text-primary hover:bg-primary/20 text-xs font-medium transition-colors"
        @click="showArchiveModal = true"
      >
        <Archive class="w-3.5 h-3.5" />
        Export
      </button>
    </div>

    <div class="flex flex-col flex-1 min-h-0 gap-4 overflow-hidden">
      <div
        class="border border-gray-200 dark:border-neutral-700 rounded-xl overflow-hidden flex flex-col flex-1 min-h-0"
      >
        <button
          class="w-full flex items-center justify-between px-4 py-3 bg-gray-50 dark:bg-neutral-700/50 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors shrink-0"
          @click="modsOpen = !modsOpen"
        >
          <div class="flex items-center gap-2">
            <Box class="w-5 h-5 text-gray-500 dark:text-neutral-400" />
            <span
              class="text-sm font-medium text-gray-700 dark:text-neutral-300"
              >Mods</span
            >
            <span class="text-xs text-gray-400 dark:text-neutral-500"
              >({{ filteredMods.length }})</span
            >
          </div>
          <ChevronDown
            class="w-4 h-4 text-gray-400 transition-transform"
            :class="modsOpen ? 'rotate-180' : ''"
          />
        </button>
        <div
          v-if="modsOpen"
          class="p-4 space-y-4 flex-1 min-h-0 overflow-y-auto"
        >
          <div
            class="space-y-2 pr-1"
            :class="
              isDragOver
                ? 'border-2 border-dashed border-primary rounded-xl bg-primary/5'
                : ''
            "
            @dragover.prevent="isDragOver = true"
            @dragleave="isDragOver = false"
            @drop.prevent="onDrop"
          >
            <div
              v-if="filteredMods.length === 0 && !loading"
              class="text-center text-gray-500 dark:text-neutral-400 py-12 text-sm"
            >
              <p>No client mods.</p>
              <p class="text-xs mt-1">
                Drag server mods here, or drop .jar files to upload.
              </p>
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
                  <span v-if="mod.size"
                    >&middot; {{ formatBytes(mod.size) }}</span
                  >
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
                  class="text-gray-400 hover:text-amber-500 transition-colors p-1"
                  title="Move to server mods"
                  @click="emit('move', mod.filename, 'server')"
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
                  Move to server mods
                </button>
                <button
                  class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700 flex items-center gap-2"
                  @click="contextCopy"
                >
                  <Copy class="w-4 h-4" />
                  Copy to server mods
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

          <client-archive-modal
            v-model:show="showArchiveModal"
            :archive-loading="archiveLoading"
            :archive-result="archiveResult"
            @generate="generateArchive"
          />
        </div>
      </div>

      <client-assets
        v-if="serverId"
        class="flex flex-col flex-1 min-h-0"
        :server-id="serverId"
        type="resourcepacks"
        title="Resource Packs"
        :icon="Image"
        :search-query="searchQuery"
      />
      <client-assets
        v-if="serverId"
        class="flex flex-col flex-1 min-h-0"
        :server-id="serverId"
        type="shaderpacks"
        title="Shader Packs"
        :icon="Sparkles"
        :search-query="searchQuery"
      />
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
  Archive,
  Copy,
  ExternalLink,
  Image,
  Sparkles,
  ChevronDown,
} from "lucide-vue-next";
import type { ModInfo } from "~/composables/useMods";
import type { ArchiveInfo } from "~/composables/useClientMods";

interface Props {
  mods: ModInfo[];
  loading: boolean;
  serverId?: string;
  searchQuery?: string;
  archiveLoading?: boolean;
  archiveResult?: ArchiveInfo | null;
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: "",
  archiveLoading: false,
  archiveResult: null,
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
  "generate-archive": [formats: string[], ttl: number, include: string[]];
}>();

const modsOpen = ref(true);

const rawQuery = ref(props.searchQuery);
let debounceTimer: ReturnType<typeof setTimeout> | null = null;

function onSearchInput(e: Event) {
  const val = (e.target as HTMLInputElement).value;
  rawQuery.value = val;
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    emit("update:searchQuery", val);
  }, 150);
}

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer);
});
const isDragOver = ref(false);
const showArchiveModal = ref(false);

const assetCounts = ref({ resourcepacks: 0, shaderpacks: 0 });

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
    emit("move", contextTarget.value.filename, "server");
  }
  closeContextMenu();
}

function contextCopy() {
  if (contextTarget.value) {
    emit("copy", contextTarget.value.filename, "client", "server");
  }
  closeContextMenu();
}

function contextDelete() {
  if (contextTarget.value) {
    emit("delete", contextTarget.value.filename);
  }
  closeContextMenu();
}

const iconLoaded = ref<Record<string, boolean>>({});

function getIconUrl(filename: string): string {
  if (!props.serverId) return "";
  return `${useApiBase()}/servers/${props.serverId}/client-mods/${encodeURIComponent(filename)}/icon`;
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
  return list;
});

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
  if (internalFilename && internalSource === "server") {
    if (e.ctrlKey) {
      emit("copy", internalFilename, "server", "client");
    } else {
      emit("move", internalFilename, "client");
    }
    return;
  }

  const file = dt.files[0];
  if (!file) return;
  const ext = file.name.split(".").pop()?.toLowerCase();
  if (ext !== "jar") {
    useToast().show("error", "Invalid file type", {
      description: "Only .jar files are allowed for client mods.",
    });
    return;
  }
  emit("upload", file);
}

function onDragStart(e: DragEvent, mod: ModInfo) {
  e.dataTransfer?.setData("application/x-mod-filename", mod.filename);
  e.dataTransfer?.setData("application/x-mod-source", "client");
}

function generateArchive(formats: string[], ttl: number, include: string[]) {
  emit("generate-archive", formats, ttl, include);
}
</script>
