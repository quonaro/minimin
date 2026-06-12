<template>
  <div class="flex flex-col h-full overflow-hidden">
    <div class="flex items-center justify-between mb-2">
      <h2 class="text-lg font-bold text-gray-900 dark:text-white">Client</h2>
      <span class="text-xs text-gray-500 dark:text-neutral-400">
        {{ filteredMods.length }} items
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

    <div
      class="mb-4 flex items-center gap-1 border-b border-gray-200 dark:border-neutral-700"
    >
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[1px] inline-flex items-center gap-1.5"
        :class="
          activeTab === tab.key
            ? 'border-primary text-primary'
            : 'border-transparent text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-300'
        "
        @click="activeTab = tab.key"
      >
        <component :is="tab.icon" class="w-4 h-4" />
        {{ tab.label }}
      </button>
    </div>

    <div class="flex-1 min-h-0 overflow-hidden">
      <div
        v-if="activeTab === 'mods'"
        class="flex flex-col h-full overflow-hidden"
      >
        <div class="flex-1 min-h-0 overflow-y-auto p-4 space-y-4">
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
            :archive-links="archiveLinks"
            :links-loading="linksLoading"
            @generate="generateArchive"
            @refresh-links="emit('refresh-archive-links')"
            @delete-link="emit('delete-archive-link', $event)"
          />
        </div>
      </div>

      <div
        v-else-if="activeTab === 'resourcepacks'"
        class="flex flex-col h-full overflow-hidden"
      >
        <client-assets
          v-if="serverId"
          embedded
          :server-id="serverId"
          type="resourcepacks"
          title="Resource Packs"
          :icon="Image"
          :search-query="searchQuery"
        />
      </div>

      <div
        v-else-if="activeTab === 'shaderpacks'"
        class="flex flex-col h-full overflow-hidden"
      >
        <client-assets
          v-if="serverId"
          embedded
          :server-id="serverId"
          type="shaderpacks"
          title="Shader Packs"
          :icon="Sparkles"
          :search-query="searchQuery"
        />
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
  Archive,
  Copy,
  ExternalLink,
  Image,
  Sparkles,
} from "lucide-vue-next";
import type { ModInfo } from "~/composables/useMods";
import type {
  ArchiveInfo,
  ArchiveLinkEntry,
} from "~/composables/useClientMods";

interface Props {
  mods: ModInfo[];
  loading: boolean;
  serverId?: string;
  searchQuery?: string;
  archiveLoading?: boolean;
  archiveResult?: ArchiveInfo | null;
  archiveLinks?: ArchiveLinkEntry[];
  linksLoading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: "",
  archiveLoading: false,
  archiveResult: null,
  archiveLinks: () => [],
  linksLoading: false,
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
  "generate-archive": [ttl: number, include: string[]];
  "refresh-archive-links": [];
  "delete-archive-link": [token: string];
}>();

const activeTab = ref<"mods" | "resourcepacks" | "shaderpacks">("mods");

const tabs = [
  { key: "mods" as const, label: "Mods", icon: Box },
  { key: "resourcepacks" as const, label: "Resource Packs", icon: Image },
  { key: "shaderpacks" as const, label: "Shader Packs", icon: Sparkles },
];

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

function generateArchive(ttl: number, include: string[]) {
  emit("generate-archive", ttl, include);
}
</script>
