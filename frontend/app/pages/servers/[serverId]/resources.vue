<template>
  <div class="p-6 h-[calc(100vh-4rem)] flex flex-col">
    <div class="mb-4">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
        Resources
      </h1>
      <p class="text-sm text-gray-500 dark:text-neutral-400 mt-1">
        Manage server and client content.
      </p>
    </div>

    <div class="mb-4 flex items-center gap-2 flex-wrap">
      <input
        ref="fileInput"
        type="file"
        accept=".jar,.zip"
        multiple
        class="hidden"
        @change="onFileSelect"
      />
      <button
        class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-primary hover:bg-primary/90 text-white text-sm font-medium transition-colors disabled:opacity-50"
        :disabled="anyUploadLoading"
        @click="showUploadModal = true"
      >
        <Upload class="w-4 h-4 text-white" />
        {{ anyUploadLoading ? "Uploading..." : "Upload .jar or .zip" }}
      </button>
      <button
        class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
        @click="showLibraryPanel = true"
      >
        <Search class="w-4 h-4 text-indigo-500" />
        Mod Library
      </button>
      <button
        v-if="activeMainTab === 'client'"
        class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
        @click="showArchiveModal = true"
      >
        <Archive class="w-4 h-4 text-amber-500" />
        Export
      </button>
      <button
        v-if="activeMainTab === 'server'"
        class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors disabled:opacity-50"
        :disabled="copyAllLoading || mods.length === 0"
        @click="openCopyAllConfirm"
      >
        <Copy class="w-4 h-4 text-blue-500" />
        {{ copyAllLoading ? "Copying..." : "Copy all to client" }}
      </button>
    </div>

    <!-- Upload / Download Modal -->
    <div
      v-if="showUploadModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
      @click.self="closeUploadModal"
    >
      <div
        class="bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 w-full max-w-md flex flex-col max-h-[80vh]"
      >
        <div
          class="p-6 border-b border-gray-200 dark:border-neutral-700 shrink-0"
        >
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">
            Add {{ uploadTargetLabel }}
          </h2>
        </div>
        <div class="p-6 space-y-4 overflow-y-auto">
          <div>
            <label
              class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
              >URL</label
            >
            <input
              v-model="modUrl"
              type="url"
              placeholder="https://example.com/file.jar"
              class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow text-sm"
              @keyup.enter="handleUploadConfirm"
            />
          </div>
          <p class="text-xs text-gray-500 dark:text-neutral-400">
            Will be downloaded to: <strong>{{ downloadTargetLabel }}</strong>
          </p>

          <div class="relative">
            <div class="absolute inset-0 flex items-center">
              <div
                class="w-full border-t border-gray-200 dark:border-neutral-700"
              />
            </div>
            <div class="relative flex justify-center text-xs">
              <span
                class="bg-white dark:bg-neutral-800 px-2 text-gray-500 dark:text-neutral-400"
                >or</span
              >
            </div>
          </div>

          <div
            class="border-2 border-dashed rounded-xl p-6 text-center transition-colors"
            :class="
              isDraggingOver
                ? 'border-primary bg-primary/5'
                : 'border-gray-300 dark:border-neutral-600'
            "
            @dragenter.prevent="isDraggingOver = true"
            @dragover.prevent="isDraggingOver = true"
            @dragleave.prevent="isDraggingOver = false"
            @drop.prevent="onDrop"
          >
            <Upload
              class="w-8 h-8 mx-auto text-gray-400 dark:text-neutral-500 mb-2"
            />
            <p class="text-sm text-gray-600 dark:text-neutral-300 mb-1">
              Drag & drop files here
            </p>
            <p class="text-xs text-gray-400 dark:text-neutral-500 mb-2">
              Supports .jar and .zip
            </p>
            <button
              class="px-3 py-1.5 rounded-lg bg-gray-100 dark:bg-neutral-700 text-gray-700 dark:text-neutral-300 text-sm font-medium hover:bg-gray-200 dark:hover:bg-neutral-600 transition-colors"
              @click="fileInput?.click()"
            >
              Select files
            </button>
          </div>

          <div v-if="pendingUploadFiles.length > 0" class="space-y-2">
            <div
              v-for="(file, idx) in pendingUploadFiles"
              :key="file.name + idx"
              class="flex items-center gap-2 p-2 rounded-lg bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
            >
              <File
                class="w-4 h-4 text-gray-400 dark:text-neutral-500 shrink-0"
              />
              <div class="flex-1 min-w-0 text-left">
                <p class="text-sm text-gray-900 dark:text-white truncate">
                  {{ file.name }}
                </p>
                <p
                  v-if="zipContentsMap[file.name]?.length"
                  class="text-xs text-gray-500 dark:text-neutral-400"
                >
                  Archive: {{ zipContentsMap[file.name]?.length || 0 }} mod(s)
                </p>
              </div>
              <button
                class="p-1 rounded hover:bg-gray-200 dark:hover:bg-neutral-600 text-gray-500 dark:text-neutral-400"
                @click="removePendingFile(idx)"
              >
                <X class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
        <div
          class="p-6 border-t border-gray-200 dark:border-neutral-700 flex gap-3 justify-end shrink-0"
        >
          <button
            class="px-4 py-2 rounded-lg text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors font-medium text-sm"
            @click="closeUploadModal"
          >
            Cancel
          </button>
          <button
            class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-primary hover:bg-primary/90 text-white text-sm font-medium transition-colors disabled:opacity-50"
            :disabled="
              bulkUploadLoading || (!modUrl && pendingUploadFiles.length === 0)
            "
            @click="handleUploadConfirm"
          >
            <Upload class="w-4 h-4" />
            {{ bulkUploadLoading ? "Uploading..." : "Upload" }}
          </button>
        </div>
      </div>
    </div>

    <!-- Copy All Modal -->
    <div
      v-if="showCopyAllModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
      @click.self="showCopyAllModal = false"
    >
      <div
        class="bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 w-full max-w-lg flex flex-col max-h-[80vh]"
      >
        <div class="p-6 border-b border-gray-200 dark:border-neutral-700">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">
            Copy all server mods to client
          </h2>
          <p class="text-sm text-gray-500 dark:text-neutral-400 mt-1">
            The following mods will be copied to the client mods folder. Already
            existing mods will be skipped.
          </p>
        </div>
        <div class="p-6 overflow-y-auto flex-1 space-y-3">
          <div
            v-if="modsToCopy.length === 0"
            class="text-center text-gray-500 dark:text-neutral-400 text-sm"
          >
            No mods to copy.
          </div>
          <div
            v-for="mod in modsToCopy"
            :key="mod.filename"
            class="flex items-center gap-3 p-2 rounded-lg bg-gray-50 dark:bg-neutral-700/50"
          >
            <div
              class="w-8 h-8 rounded bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400 shrink-0"
            >
              <Box class="w-4 h-4" />
            </div>
            <div class="flex-1 min-w-0">
              <p
                class="text-sm font-medium text-gray-900 dark:text-white truncate"
              >
                {{ mod.name || mod.filename }}
              </p>
              <p class="text-xs text-gray-500 dark:text-neutral-400 truncate">
                {{ mod.filename }}
              </p>
            </div>
          </div>
        </div>
        <div
          class="p-6 border-t border-gray-200 dark:border-neutral-700 flex gap-3 justify-end"
        >
          <button
            class="px-4 py-2 rounded-lg text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors font-medium text-sm"
            @click="showCopyAllModal = false"
          >
            Cancel
          </button>
          <button
            class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-primary hover:bg-primary/90 text-white text-sm font-medium transition-colors disabled:opacity-50"
            :disabled="copyAllLoading || modsToCopy.length === 0"
            @click="handleCopyAll"
          >
            <Copy class="w-4 h-4" />
            {{
              copyAllLoading
                ? "Copying..."
                : "Copy " +
                  modsToCopy.length +
                  " mod" +
                  (modsToCopy.length !== 1 ? "s" : "")
            }}
          </button>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <div
      class="mb-4 flex items-center gap-1 border-b border-gray-200 dark:border-neutral-700"
    >
      <button
        class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[1px] inline-flex items-center gap-1.5"
        :class="
          activeMainTab === 'server'
            ? 'border-primary text-primary'
            : 'border-transparent text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-300'
        "
        @click="activeMainTab = 'server'"
      >
        <Server
          class="w-4 h-4"
          :class="
            activeMainTab === 'server'
              ? 'text-blue-500'
              : 'text-gray-400 dark:text-neutral-500'
          "
        />
        Server
      </button>

      <button
        class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[1px] inline-flex items-center gap-1.5"
        :class="
          activeMainTab === 'client'
            ? 'border-primary text-primary'
            : 'border-transparent text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-300'
        "
        @click="activeMainTab = 'client'"
      >
        <Monitor
          class="w-4 h-4"
          :class="
            activeMainTab === 'client'
              ? 'text-emerald-500'
              : 'text-gray-400 dark:text-neutral-500'
          "
        />
        Client
      </button>

      <Transition name="tab-appear">
        <div v-if="activeMainTab === 'client'" class="flex items-center gap-1">
          <span class="text-gray-400 dark:text-neutral-600 mx-1">·</span>
          <button
            v-for="tab in clientTabs"
            :key="tab.key"
            class="px-3 py-2 text-sm font-medium transition-colors inline-flex items-center gap-1.5 border-b-2 -mb-[1px]"
            :class="
              activeClientSubTab === tab.key
                ? tab.key === 'mods'
                  ? 'border-blue-500 text-blue-500'
                  : tab.key === 'resourcepacks'
                    ? 'border-emerald-500 text-emerald-500'
                    : 'border-purple-500 text-purple-500'
                : 'border-transparent text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-300'
            "
            @click="activeClientSubTab = tab.key"
          >
            <component
              :is="tab.icon"
              class="w-4 h-4"
              :class="
                activeClientSubTab === tab.key
                  ? tab.key === 'mods'
                    ? 'text-blue-500'
                    : tab.key === 'resourcepacks'
                      ? 'text-emerald-500'
                      : 'text-purple-500'
                  : 'text-gray-400 dark:text-neutral-500'
              "
            />
            {{ tab.label }}
          </button>
        </div>
      </Transition>
    </div>

    <!-- Server Tab -->
    <div
      v-if="activeMainTab === 'server'"
      class="flex-1 min-h-0 overflow-hidden bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm"
    >
      <div class="p-4 md:p-6 h-full">
        <installed-mods
          :mods="mods"
          :loading="loading"
          :server-id="serverId"
          v-model:search-query="installedSearchQuery"
          v-model:side-filter="installedSideFilter"
          @delete="deleteMod"
          @upload="handleUpload"
          @toggle="handleToggle"
          @move="handleClientMove"
          @copy="handleCopy"
        />
      </div>
    </div>

    <!-- Client Tab -->
    <div
      v-else
      class="flex flex-col flex-1 min-h-0 overflow-hidden bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm"
    >
      <div class="p-4 md:p-6 h-full flex flex-col">
        <div class="flex-1 min-h-0 overflow-hidden">
          <client-mods
            v-if="activeClientSubTab === 'mods'"
            :mods="clientModList"
            :loading="clientLoading"
            :server-id="serverId"
            :archive-loading="clientArchiveLoading"
            :archive-result="archiveResult"
            :archive-links="archiveLinks"
            :links-loading="linksLoading"
            :show-archive-modal="showArchiveModal"
            v-model:search-query="clientSearchQuery"
            @update:show-archive-modal="showArchiveModal = $event"
            @delete="handleClientDelete"
            @upload="handleClientUpload"
            @toggle="handleClientToggle"
            @move="handleClientMove"
            @copy="handleCopy"
            @generate-archive="handleGenerateArchive"
            @refresh-archive-links="handleRefreshLinks"
            @delete-archive-link="handleDeleteLink"
          />
          <client-assets
            v-else-if="activeClientSubTab === 'resourcepacks'"
            :server-id="serverId"
            type="resourcepacks"
            title="Resource Packs"
            :icon="Image"
            :search-query="clientSearchQuery"
          />
          <client-assets
            v-else-if="activeClientSubTab === 'shaderpacks'"
            :server-id="serverId"
            type="shaderpacks"
            title="Shader Packs"
            :icon="Sparkles"
            :search-query="clientSearchQuery"
          />
        </div>
      </div>
    </div>

    <!-- Slide-over: Mod Library -->
    <Transition name="slide-over">
      <div v-if="showLibraryPanel" class="fixed inset-0 z-50">
        <div
          class="absolute inset-0 bg-black/30 backdrop-blur-sm"
          @click="showLibraryPanel = false"
        />
        <div
          class="absolute inset-y-0 right-0 w-full max-w-xl bg-white dark:bg-neutral-800 border-l border-gray-200 dark:border-neutral-700 shadow-2xl flex flex-col"
        >
          <div
            class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-neutral-700 shrink-0"
          >
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">
              Mod Library
            </h2>
            <button
              class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-neutral-700 text-gray-500 dark:text-neutral-400 transition-colors"
              @click="showLibraryPanel = false"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="flex-1 min-h-0 overflow-hidden p-4 md:p-6">
            <mod-library
              :search-query="modrinth.searchQuery.value"
              :search-results="modrinth.searchResults.value"
              :search-loading="modrinth.searchLoading.value"
              :loader="serverEngine"
              :game-version="serverGameVersion"
              :install-loading="modrinth.installLoading.value"
              :versions="versionsMap"
              :has-more="modrinth.hasMore.value"
              :version-details="versionDetailsMap"
              :dep-projects="depProjectsMap"
              :project-type="modrinth.projectType.value"
              :installed-mod-basenames="installedModBasenames"
              :installed-asset-basenames="installedAssetBasenames"
              :locked-type="libraryProjectType"
              v-model:side-filter="librarySideFilter"
              @update:search-query="onSearchInput"
              @update:project-type="onProjectTypeChange"
              @search="modrinth.search(serverEngine, serverGameVersion)"
              @load-more="modrinth.searchMore(serverEngine, serverGameVersion)"
              @install="handleInstallFromLibrary"
              @load-versions="handleLoadVersions"
              @load-version-details="handleLoadVersionDetails"
              @load-project="handleLoadProject"
            />
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.slide-over-enter-active,
.slide-over-leave-active {
  transition: opacity 0.3s ease;
}
.slide-over-enter-from,
.slide-over-leave-to {
  opacity: 0;
}
.slide-over-enter-active .absolute.inset-y-0.right-0,
.slide-over-leave-active .absolute.inset-y-0.right-0 {
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.slide-over-enter-from .absolute.inset-y-0.right-0,
.slide-over-leave-to .absolute.inset-y-0.right-0 {
  transform: translateX(100%);
}

/* Inline sub-tabs appear animation */
.tab-appear-enter-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}
.tab-appear-leave-active {
  transition: all 0.15s ease;
}
.tab-appear-enter-from,
.tab-appear-leave-to {
  opacity: 0;
  transform: translateX(-10px);
}
</style>

<script setup lang="ts">
import { debounce } from "~/utils/debounce";
import {
  Upload,
  Link,
  Download,
  Copy,
  Box,
  File,
  Server,
  Monitor,
  Image,
  Sparkles,
  Search,
  X,
  Archive,
} from "lucide-vue-next";
import JSZip from "jszip";
import { useClientAssetsRefresh } from "~/composables/useClientAssetsRefresh";
import { useClientAssets } from "~/composables/useClientAssets";

usePageTitle("Resources");

definePageMeta({ middleware: "auth" });

const route = useRoute();
const serverId = route.params.serverId as string;

const {
  mods,
  loading,
  uploadLoading,
  downloadLoading,
  refresh,
  deleteMod,
  uploadFile,
  downloadFromURL,
  toggleMod,
} = useMods(serverId);
const {
  mods: clientModList,
  loading: clientLoading,
  uploadLoading: clientUploadLoading,
  downloadLoading: clientDownloadLoading,
  archiveLoading: clientArchiveLoading,
  refresh: refreshClient,
  deleteMod: deleteClientMod,
  uploadFile: uploadClientFile,
  downloadFromURL: downloadClientFromURL,
  toggleMod: toggleClientMod,
  moveMod: moveClientMod,
  copyMod: copyClientMod,
  createArchive,
  listArchives,
  deleteArchive,
} = useClientMods(serverId);

const resourcePacks = useClientAssets(serverId, "resourcepacks");
const shaderPacks = useClientAssets(serverId, "shaderpacks");

const fileInput = ref<HTMLInputElement | null>(null);
const showUploadModal = ref(false);
const showLibraryPanel = ref(false);
const showArchiveModal = ref(false);
const modUrl = ref("");
const installedSearchQuery = ref("");
const clientSearchQuery = ref("");
const pendingUploadFiles = ref<File[]>([]);
const zipContentsMap = ref<Record<string, string[]>>({});
const isDraggingOver = ref(false);
const bulkUploadLoading = ref(false);

const activeMainTab = ref<"server" | "client">("server");
const activeClientSubTab = ref<"mods" | "resourcepacks" | "shaderpacks">(
  "mods",
);

const clientTabs = [
  { key: "mods" as const, label: "Mods", icon: Box },
  { key: "resourcepacks" as const, label: "Resource Packs", icon: Image },
  { key: "shaderpacks" as const, label: "Shader Packs", icon: Sparkles },
];

const { trigger: triggerClientAssetsRefresh } =
  useClientAssetsRefresh(serverId);

const assetUploadLoading = computed(() => {
  if (activeClientSubTab.value === "resourcepacks")
    return resourcePacks.uploadLoading.value;
  if (activeClientSubTab.value === "shaderpacks")
    return shaderPacks.uploadLoading.value;
  return false;
});

const assetDownloadLoading = computed(() => false);

const anyUploadLoading = computed(
  () =>
    uploadLoading.value ||
    clientUploadLoading.value ||
    assetUploadLoading.value ||
    bulkUploadLoading.value,
);

const uploadTargetLabel = computed(() => {
  if (activeMainTab.value === "server") return "Server Mod";
  if (activeClientSubTab.value === "mods") return "Client Mod";
  if (activeClientSubTab.value === "resourcepacks") return "Resource Pack";
  return "Shader Pack";
});

const downloadTargetLabel = computed(() => {
  if (activeMainTab.value === "server") return "server mods";
  if (activeClientSubTab.value === "mods") return "client mods";
  if (activeClientSubTab.value === "resourcepacks") return "resource packs";
  return "shader packs";
});

const libraryProjectType = computed<"mod" | "resourcepack" | "shaderpack">(
  () => {
    if (activeMainTab.value === "server") return "mod";
    if (activeClientSubTab.value === "resourcepacks") return "resourcepack";
    if (activeClientSubTab.value === "shaderpacks") return "shaderpack";
    return "mod";
  },
);

watch(showLibraryPanel, (open) => {
  if (!open) return;
  const lt = libraryProjectType.value;
  if (lt && modrinth.projectType.value !== lt) {
    modrinth.projectType.value = lt;
    modrinth.search(serverEngine.value, serverGameVersion.value);
  }
});

const installedModBasenames = computed(() => {
  const server = new Set<string>(),
    client = new Set<string>();
  const add = (set: Set<string>, filename: string) => {
    const base = filename.replace(/\.[^.]+$/, "").toLowerCase();
    if (base) set.add(base);
  };
  mods.value.forEach((m) => add(server, m.filename));
  clientModList.value.forEach((m) => add(client, m.filename));
  return { server, client };
});

const installedAssetBasenames = computed(() => {
  const names = new Set<string>();
  const add = (filename: string) => {
    const base = filename.replace(/\.[^.]+$/, "").toLowerCase();
    if (base) names.add(base);
  };
  resourcePacks.assets.value.forEach((a) => add(a.filename));
  shaderPacks.assets.value.forEach((a) => add(a.filename));
  return names;
});

const installedSideFilter = ref<"all" | "server" | "client">("all");
const librarySideFilter = ref<"all" | "server" | "client">("all");

const showCopyAllModal = ref(false);
const copyAllLoading = ref(false);
const modsToCopy = ref<typeof mods.value>([]);

function openCopyAllConfirm() {
  const clientFilenames = new Set(clientModList.value.map((m) => m.filename));
  modsToCopy.value = mods.value.filter((m) => !clientFilenames.has(m.filename));
  showCopyAllModal.value = true;
}

async function handleCopyAll() {
  if (modsToCopy.value.length === 0) {
    showCopyAllModal.value = false;
    return;
  }
  copyAllLoading.value = true;
  try {
    const res: any = await $fetch(`/servers/${serverId}/mods/copy-all`, {
      baseURL: useApiBase(),
      method: "POST",
      credentials: "include",
    });
    useToast().show(
      "success",
      `Copied ${res.copied?.length || 0} mod(s) to client`,
      {
        description: res.skipped?.length
          ? `${res.skipped.length} skipped (already exist)`
          : undefined,
      },
    );
    await refresh();
    await refreshClient();
    showCopyAllModal.value = false;
  } catch (err: any) {
    useToast().show("error", "Failed to copy mods", {
      description: err?.data?.detail || err?.message || "Unknown error",
    });
  } finally {
    copyAllLoading.value = false;
  }
}

const versionDetailsMap = ref<
  Record<string, import("~/composables/useModrinth").ModrinthVersion>
>({});
const depProjectsMap = ref<
  Record<string, import("~/composables/useModrinth").ModrinthProject>
>({});

async function onFileSelect(e: Event) {
  const input = e.target as HTMLInputElement;
  if (!input.files) return;
  for (let i = 0; i < input.files.length; i++) {
    const file = input.files[i];
    if (!file) continue;
    if (
      !file.name.toLowerCase().endsWith(".jar") &&
      !file.name.toLowerCase().endsWith(".zip")
    )
      continue;
    if (pendingUploadFiles.value.some((f) => f.name === file.name)) continue;
    pendingUploadFiles.value.push(file);
    if (file.name.toLowerCase().endsWith(".zip")) {
      try {
        const zip = await JSZip.loadAsync(file);
        const jars: string[] = [];
        zip.forEach((relativePath, entry) => {
          if (!entry.dir && relativePath.toLowerCase().endsWith(".jar"))
            jars.push(relativePath);
        });
        zipContentsMap.value[file.name] = jars;
      } catch {
        /* ignore */
      }
    }
  }
  input.value = "";
}

async function handleUploadConfirm() {
  if (!modUrl.value && pendingUploadFiles.value.length === 0) return;
  bulkUploadLoading.value = true;
  try {
    if (modUrl.value) {
      await performDownloadFromURL();
      modUrl.value = "";
    }
    for (const file of pendingUploadFiles.value) {
      if (activeMainTab.value === "server") {
        await uploadFile(file);
      } else if (activeClientSubTab.value === "mods") {
        await uploadClientFile(file);
      } else {
        const assetType = activeClientSubTab.value;
        const formData = new FormData();
        formData.append("file", file);
        await $fetch(
          `/servers/${serverId}/client-assets/upload?type=${assetType}`,
          {
            baseURL: useApiBase(),
            method: "POST",
            body: formData,
            credentials: "include",
          },
        );
        useToast().show("success", "Upload complete", {
          description: file.name,
        });
        triggerClientAssetsRefresh();
      }
    }
    if (
      activeClientSubTab.value === "mods" &&
      pendingUploadFiles.value.length > 0
    ) {
      await refreshClient();
    }
    if (
      activeMainTab.value === "server" &&
      pendingUploadFiles.value.length > 0
    ) {
      await refresh();
    }
  } catch (err: any) {
    useToast().show("error", "Failed to process files", {
      description: err?.data?.detail || err?.message || "Unknown error",
    });
  } finally {
    bulkUploadLoading.value = false;
    closeUploadModal();
  }
}

async function downloadResourcePackFromURL(url: string, filename?: string) {
  await $fetch(
    `/servers/${serverId}/client-assets/download?type=resourcepacks`,
    {
      baseURL: useApiBase(),
      method: "POST",
      credentials: "include",
      body: { url, filename },
    },
  );
}

async function downloadShaderPackFromURL(url: string, filename?: string) {
  await $fetch(`/servers/${serverId}/client-assets/download?type=shaderpacks`, {
    baseURL: useApiBase(),
    method: "POST",
    credentials: "include",
    body: { url, filename },
  });
}

async function performDownloadFromURL() {
  if (!modUrl.value) return;
  if (activeMainTab.value === "server") await downloadFromURL(modUrl.value);
  else if (activeClientSubTab.value === "mods") {
    await downloadClientFromURL(modUrl.value);
    await refreshClient();
  } else if (activeClientSubTab.value === "resourcepacks") {
    await downloadResourcePackFromURL(modUrl.value);
    triggerClientAssetsRefresh();
  } else if (activeClientSubTab.value === "shaderpacks") {
    await downloadShaderPackFromURL(modUrl.value);
    triggerClientAssetsRefresh();
  }
}

function closeUploadModal() {
  showUploadModal.value = false;
  pendingUploadFiles.value = [];
  zipContentsMap.value = {};
  modUrl.value = "";
}

function onDrop(e: DragEvent) {
  isDraggingOver.value = false;
  const files = e.dataTransfer?.files;
  if (!files) return;
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    if (!file) continue;
    if (
      !file.name.toLowerCase().endsWith(".jar") &&
      !file.name.toLowerCase().endsWith(".zip")
    )
      continue;
    if (pendingUploadFiles.value.some((f) => f.name === file.name)) continue;
    pendingUploadFiles.value.push(file);
    if (file.name.toLowerCase().endsWith(".zip")) {
      JSZip.loadAsync(file)
        .then((zip) => {
          const jars: string[] = [];
          zip.forEach((relativePath, entry) => {
            if (!entry.dir && relativePath.toLowerCase().endsWith(".jar"))
              jars.push(relativePath);
          });
          zipContentsMap.value[file.name] = jars;
        })
        .catch(() => {
          /* ignore */
        });
    }
  }
}

function removePendingFile(idx: number) {
  const file = pendingUploadFiles.value[idx];
  pendingUploadFiles.value.splice(idx, 1);
  if (file) delete zipContentsMap.value[file.name];
}

const modrinth = useModrinth();
const serverEngine = ref("");
const serverGameVersion = ref("");
const versionsMap = ref<
  Record<string, import("~/composables/useModrinth").ModrinthVersion[]>
>({});

const debouncedSearch = debounce(() => {
  modrinth.search(serverEngine.value, serverGameVersion.value);
}, 1300);
onBeforeUnmount(() => {
  debouncedSearch.cancel();
});

function onSearchInput(v: string) {
  modrinth.searchQuery.value = v;
  debouncedSearch();
}
function onProjectTypeChange(v: "mod" | "resourcepack" | "shaderpack") {
  modrinth.projectType.value = v;
  debouncedSearch();
}

async function fetchServerInfo() {
  try {
    const res = await $fetch<
      | { body: { engineType: string; gameVersion: string } }
      | { engineType: string; gameVersion: string }
    >(`/servers/${serverId}`, {
      baseURL: useApiBase(),
      credentials: "include",
    });
    const data = (res as any).body ?? res;
    const engine = (data.engineType ?? "").toUpperCase();
    serverEngine.value = engine;
    serverGameVersion.value = data.gameVersion ?? "";
    if (engine === "PAPERMC" || engine === "PAPER") {
      await navigateTo(`/servers/${serverId}/plugins`, { replace: true });
      return;
    }
    if (engine === "VANILLA") {
      await navigateTo(`/servers/${serverId}`, { replace: true });
      return;
    }
  } catch {
    /* ignore */
  }
}

async function handleUpload(file: File) {
  await uploadFile(file);
}
async function handleToggle(filename: string) {
  await toggleMod(filename);
}
async function handleClientUpload(file: File) {
  await uploadClientFile(file);
}
async function handleClientDelete(filename: string) {
  await deleteClientMod(filename);
}
async function handleClientToggle(filename: string) {
  await toggleClientMod(filename);
}
async function handleClientMove(filename: string, target: "server" | "client") {
  await moveClientMod(filename, target);
  await refresh();
  await refreshClient();
}
async function handleCopy(
  filename: string,
  source: "server" | "client",
  target: "server" | "client",
) {
  await copyClientMod(filename, source, target);
  await refresh();
  await refreshClient();
}

async function handleInstallFromLibrary(
  projectId: string,
  versionId: string,
  depProjectIds?: string[],
  _target?: "server" | "client" | "both",
) {
  const file = await modrinth.install(projectId, versionId);
  if (!file) return;

  const projectType = modrinth.projectType.value;
  const isAsset =
    projectType === "resourcepack" || projectType === "shaderpack";
  const assetType =
    projectType === "shaderpack" ? "shaderpacks" : "resourcepacks";

  const allDeps = file.dependencies.filter(
    (d) => d.dependency_type !== "incompatible",
  );
  const selectedDeps = depProjectIds
    ? allDeps.filter((d) => depProjectIds.includes(d.project_id))
    : allDeps.filter((d) => d.dependency_type === "required");

  async function installDepsAndMod(
    modFile: NonNullable<typeof file>,
    doDownload: (url: string, filename?: string) => Promise<void>,
    doRefresh: () => Promise<void>,
  ) {
    if (selectedDeps.length > 0) {
      useToast().show(
        "info",
        `Installing ${selectedDeps.length} dependencies...`,
      );
      for (const dep of selectedDeps) {
        const depFile = await modrinth.resolveDependency(
          dep.project_id,
          dep.version_id,
          serverEngine.value,
          serverGameVersion.value,
        );
        if (depFile) await doDownload(depFile.url, depFile.filename);
      }
    }
    await doDownload(modFile.url, modFile.filename);
    await doRefresh();
  }

  if (isAsset) {
    const doDownload =
      assetType === "shaderpacks"
        ? downloadShaderPackFromURL
        : downloadResourcePackFromURL;
    await installDepsAndMod(file, doDownload, async () => {
      triggerClientAssetsRefresh();
    });
    useToast().show("success", "Installed to client", {
      description: file.filename,
    });
    return;
  }

  if (activeMainTab.value === "server") {
    await installDepsAndMod(file, downloadFromURL, refresh);
    useToast().show("success", "Mod installed to server", {
      description: file.filename,
    });
  } else {
    await installDepsAndMod(file, downloadClientFromURL, refreshClient);
    useToast().show("success", "Mod installed to client", {
      description: file.filename,
    });
  }
}

async function handleLoadVersions(projectId: string) {
  const loader = modrinth.projectType.value === "mod" ? serverEngine.value : "";
  const list = await modrinth.getVersions(
    projectId,
    loader,
    serverGameVersion.value,
  );
  if (list.length > 0)
    versionsMap.value = { ...versionsMap.value, [projectId]: list };
}

async function handleLoadVersionDetails(versionId: string) {
  const details = await modrinth.getVersionDetails(versionId);
  if (details)
    versionDetailsMap.value = {
      ...versionDetailsMap.value,
      [versionId]: details,
    };
}

async function handleLoadProject(projectId: string) {
  const project = await modrinth.getProject(projectId);
  if (project)
    depProjectsMap.value = { ...depProjectsMap.value, [projectId]: project };
}

const archiveResult = ref<
  import("~/composables/useClientMods").ArchiveInfo | null
>(null);
const archiveLinks = ref<
  import("~/composables/useClientMods").ArchiveLinkEntry[]
>([]);
const linksLoading = ref(false);

async function handleGenerateArchive(
  ttl: number,
  include: string[] = ["mods"],
) {
  const result = await createArchive(ttl, include);
  if (result) {
    archiveResult.value = result;
    await handleRefreshLinks();
  }
}

async function handleRefreshLinks() {
  linksLoading.value = true;
  try {
    archiveLinks.value = await listArchives();
  } finally {
    linksLoading.value = false;
  }
}

async function handleDeleteLink(token: string) {
  const ok = await deleteArchive(token);
  if (ok)
    archiveLinks.value = archiveLinks.value.filter((l) => l.token !== token);
}

onMounted(async () => {
  await fetchServerInfo();
  if (!serverEngine.value) return;
  await Promise.all([
    refresh(),
    refreshClient(),
    resourcePacks.refresh(),
    shaderPacks.refresh(),
    modrinth.search(serverEngine.value, serverGameVersion.value),
  ]);
});
</script>
