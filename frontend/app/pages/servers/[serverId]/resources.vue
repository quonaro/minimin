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
        class="hidden"
        @change="onFileSelect"
      />
      <button
        class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-primary hover:bg-primary/90 text-white text-sm font-medium transition-colors disabled:opacity-50"
        :disabled="uploadLoading || clientUploadLoading || assetUploadLoading"
        @click="fileInput?.click()"
      >
        <Upload class="w-4 h-4" />
        {{
          uploadLoading || clientUploadLoading || assetUploadLoading
            ? "Uploading..."
            : "Upload .jar or .zip"
        }}
      </button>
      <button
        class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors disabled:opacity-50"
        :disabled="
          downloadLoading || clientDownloadLoading || assetDownloadLoading
        "
        @click="showDownloadModal = true"
      >
        <Link class="w-4 h-4" />
        {{
          downloadLoading || clientDownloadLoading || assetDownloadLoading
            ? "Downloading..."
            : "Download from URL"
        }}
      </button>
      <button
        class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
        @click="showLibraryPanel = true"
      >
        <Search class="w-4 h-4" />
        Mod Library
      </button>
      <button
        v-if="activeMainTab === 'server'"
        class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors disabled:opacity-50"
        :disabled="copyAllLoading || mods.length === 0"
        @click="openCopyAllConfirm"
      >
        <Copy class="w-4 h-4" />
        {{ copyAllLoading ? "Copying..." : "Copy all to client" }}
      </button>
    </div>

    <!-- Upload Modal -->
    <div
      v-if="showUploadModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
      @click.self="showUploadModal = false"
    >
      <div
        class="bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 w-full max-w-md"
      >
        <div class="p-6 border-b border-gray-200 dark:border-neutral-700">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">
            Upload {{ uploadTargetLabel }}
          </h2>
        </div>
        <div class="p-6 space-y-4">
          <div>
            <label
              class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
              >File</label
            >
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-900 dark:text-white truncate">{{
                pendingUploadFile?.name || "No file selected"
              }}</span>
              <button
                class="text-primary text-sm font-medium hover:underline"
                @click="fileInput?.click()"
              >
                Choose file
              </button>
            </div>
            <div
              v-if="zipContents.length > 0 && activeMainTab === 'server'"
              class="mt-2 max-h-32 overflow-y-auto rounded-lg border border-gray-200 dark:border-neutral-700 bg-gray-50 dark:bg-neutral-800 p-2 space-y-1"
            >
              <p class="text-xs text-gray-500 dark:text-neutral-400 mb-1">
                Archive contains {{ zipContents.length }} mod(s):
              </p>
              <div
                v-for="name in zipContents"
                :key="name"
                class="text-xs text-gray-700 dark:text-neutral-300 truncate"
              >
                {{ name }}
              </div>
            </div>
          </div>
        </div>
        <div
          class="p-6 border-t border-gray-200 dark:border-neutral-700 flex gap-3 justify-end"
        >
          <button
            class="px-4 py-2 rounded-lg text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors font-medium text-sm"
            @click="
              showUploadModal = false;
              pendingUploadFile = null;
              zipContents = [];
            "
          >
            Cancel
          </button>
          <button
            class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-primary hover:bg-primary/90 text-white text-sm font-medium transition-colors disabled:opacity-50"
            :disabled="
              !pendingUploadFile ||
              uploadLoading ||
              clientUploadLoading ||
              assetUploadLoading
            "
            @click="handleUploadConfirm"
          >
            <Upload class="w-4 h-4" />
            {{
              uploadLoading || clientUploadLoading || assetUploadLoading
                ? "Uploading..."
                : "Upload"
            }}
          </button>
        </div>
      </div>
    </div>

    <!-- Download Modal -->
    <div
      v-if="showDownloadModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
      @click.self="showDownloadModal = false"
    >
      <div
        class="bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 w-full max-w-md"
      >
        <div class="p-6 border-b border-gray-200 dark:border-neutral-700">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">
            Download from URL
          </h2>
        </div>
        <div class="p-6 space-y-4">
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
              @keyup.enter="handleDownloadFromURL"
            />
          </div>
          <p class="text-xs text-gray-500 dark:text-neutral-400">
            Will be downloaded to: <strong>{{ downloadTargetLabel }}</strong>
          </p>
        </div>
        <div
          class="p-6 border-t border-gray-200 dark:border-neutral-700 flex gap-3 justify-end"
        >
          <button
            class="px-4 py-2 rounded-lg text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors font-medium text-sm"
            @click="
              showDownloadModal = false;
              modUrl = '';
            "
          >
            Cancel
          </button>
          <button
            class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-primary hover:bg-primary/90 text-white text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="
              downloadLoading ||
              clientDownloadLoading ||
              assetDownloadLoading ||
              !modUrl
            "
            @click="handleDownloadFromURL"
          >
            <Download class="w-4 h-4" />
            {{
              downloadLoading || clientDownloadLoading || assetDownloadLoading
                ? "Downloading..."
                : "Download"
            }}
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

    <!-- Main Tabs -->
    <div
      class="mb-4 flex items-center gap-1 border-b border-gray-200 dark:border-neutral-700"
    >
      <button
        v-for="tab in mainTabs"
        :key="tab.key"
        class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[1px] inline-flex items-center gap-1.5"
        :class="
          activeMainTab === tab.key
            ? 'border-primary text-primary'
            : 'border-transparent text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-300'
        "
        @click="activeMainTab = tab.key"
      >
        <component
          :is="tab.icon"
          class="w-4 h-4"
          :class="
            activeMainTab === tab.key
              ? tab.key === 'server'
                ? 'text-blue-500'
                : 'text-emerald-500'
              : 'text-gray-400 dark:text-neutral-500'
          "
        />
        {{ tab.label }}
      </button>
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
        <div
          class="mb-4 flex items-center gap-1 border-b border-gray-200 dark:border-neutral-700"
        >
          <button
            v-for="tab in clientTabs"
            :key="tab.key"
            class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[1px] inline-flex items-center gap-1.5"
            :class="
              activeClientSubTab === tab.key
                ? 'border-primary text-primary'
                : 'border-transparent text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-300'
            "
            @click="activeClientSubTab = tab.key"
          >
            <component :is="tab.icon" class="w-4 h-4" />
            {{ tab.label }}
          </button>
        </div>
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
            v-model:search-query="clientSearchQuery"
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
            :show-upload="true"
          />
          <client-assets
            v-else-if="activeClientSubTab === 'shaderpacks'"
            :server-id="serverId"
            type="shaderpacks"
            title="Shader Packs"
            :icon="Sparkles"
            :search-query="clientSearchQuery"
            :show-upload="true"
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
</style>

<script setup lang="ts">
import { debounce } from "~/utils/debounce";
import {
  Upload,
  Link,
  Download,
  Copy,
  Box,
  Server,
  Monitor,
  Image,
  Sparkles,
  Search,
  X,
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
const showDownloadModal = ref(false);
const showLibraryPanel = ref(false);
const modUrl = ref("");
const installedSearchQuery = ref("");
const clientSearchQuery = ref("");
const pendingUploadFile = ref<File | null>(null);
const zipContents = ref<string[]>([]);

const activeMainTab = ref<"server" | "client">("server");
const activeClientSubTab = ref<"mods" | "resourcepacks" | "shaderpacks">(
  "mods",
);

const mainTabs = [
  { key: "server" as const, label: "Server", icon: Server },
  { key: "client" as const, label: "Client", icon: Monitor },
];

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
  const file = input.files?.[0];
  if (!file) return;
  pendingUploadFile.value = file;
  zipContents.value = [];
  if (file.name.toLowerCase().endsWith(".zip")) {
    try {
      const zip = await JSZip.loadAsync(file);
      const jars: string[] = [];
      zip.forEach((relativePath, entry) => {
        if (!entry.dir && relativePath.toLowerCase().endsWith(".jar"))
          jars.push(relativePath);
      });
      zipContents.value = jars;
    } catch {
      zipContents.value = [];
    }
  }
  showUploadModal.value = true;
  input.value = "";
}

async function handleUploadConfirm() {
  if (!pendingUploadFile.value) return;
  const file = pendingUploadFile.value;
  if (activeMainTab.value === "server") {
    await uploadFile(file);
  } else if (activeClientSubTab.value === "mods") {
    await uploadClientFile(file);
    await refreshClient();
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
    useToast().show("success", "Upload complete", { description: file.name });
    triggerClientAssetsRefresh();
  }
  pendingUploadFile.value = null;
  zipContents.value = [];
  showUploadModal.value = false;
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

async function handleDownloadFromURL() {
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
  modUrl.value = "";
  showDownloadModal.value = false;
}

const modrinth = useModrinth();
const serverEngine = ref("");
const serverGameVersion = ref("");
const versionsMap = ref<
  Record<string, import("~/composables/useModrinth").ModrinthVersion[]>
>({});

const debouncedSearch = debounce(() => {
  modrinth.search(serverEngine.value, serverGameVersion.value);
}, 400);
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
