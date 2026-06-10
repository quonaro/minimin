<template>
  <div class="p-6 h-[calc(100vh-4rem)] flex flex-col">
    <div class="mb-4">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Mods</h1>
      <p class="text-sm text-gray-500 dark:text-neutral-400 mt-1">
        Manage installed mods and browse Modrinth library.
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
        :disabled="uploadLoading"
        @click="fileInput?.click()"
      >
        <Upload class="w-4 h-4" />
        {{ uploadLoading ? "Uploading..." : "Upload .jar or .zip" }}
      </button>
      <button
        class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors disabled:opacity-50"
        :disabled="downloadLoading"
        @click="showDownloadModal = true"
      >
        <Link class="w-4 h-4" />
        {{ downloadLoading ? "Downloading..." : "Download from URL" }}
      </button>
    </div>

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
            >
              Mod URL
            </label>
            <input
              v-model="modUrl"
              type="url"
              placeholder="https://example.com/mod.jar"
              class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow text-sm"
              @keyup.enter="handleDownloadFromURL"
            />
          </div>
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
            :disabled="downloadLoading || !modUrl"
            @click="handleDownloadFromURL"
          >
            <Download class="w-4 h-4" />
            {{ downloadLoading ? "Downloading..." : "Download" }}
          </button>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 flex-1 min-h-0">
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden flex flex-col"
      >
        <div class="p-4 md:p-6 flex-1 min-h-0">
          <installed-mods
            :mods="mods"
            :loading="loading"
            :server-id="serverId"
            v-model:search-query="installedSearchQuery"
            v-model:side-filter="installedSideFilter"
            @delete="deleteMod"
            @upload="handleUpload"
            @toggle="handleToggle"
          />
        </div>
      </div>

      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden flex flex-col"
      >
        <div class="p-4 md:p-6 flex-1 min-h-0">
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
            v-model:side-filter="librarySideFilter"
            @update:search-query="onSearchInput"
            @search="modrinth.search(serverEngine, serverGameVersion)"
            @load-more="modrinth.searchMore(serverEngine, serverGameVersion)"
            @install="handleInstall"
            @load-versions="handleLoadVersions"
            @load-version-details="handleLoadVersionDetails"
            @load-project="handleLoadProject"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { debounce } from "~/utils/debounce";
import { Upload, Link, Download } from "lucide-vue-next";

definePageMeta({
  middleware: "auth",
});

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

const fileInput = ref<HTMLInputElement | null>(null);
const showDownloadModal = ref(false);
const modUrl = ref("");
const installedSearchQuery = ref("");
const installedSideFilter = ref<"all" | "server" | "client">("all");
const librarySideFilter = ref<"all" | "server" | "client">("all");
const versionDetailsMap = ref<
  Record<string, import("~/composables/useModrinth").ModrinthVersion>
>({});
const depProjectsMap = ref<
  Record<string, import("~/composables/useModrinth").ModrinthProject>
>({});

function onFileSelect(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  uploadFile(file);
  input.value = "";
}

async function handleDownloadFromURL() {
  if (!modUrl.value) return;
  await downloadFromURL(modUrl.value);
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
      await navigateTo(`/servers/${serverId}/plugins`, {
        replace: true,
      });
      return;
    }
    if (engine === "VANILLA") {
      await navigateTo(`/servers/${serverId}`, {
        replace: true,
      });
      return;
    }
  } catch {
    // ignore
  }
}

async function handleUpload(file: File) {
  await uploadFile(file);
}

async function handleToggle(filename: string) {
  await toggleMod(filename);
}

async function handleInstall(
  projectId: string,
  versionId: string,
  depProjectIds?: string[],
) {
  const file = await modrinth.install(projectId, versionId);
  if (!file) return;

  const allDeps = file.dependencies.filter(
    (d) => d.dependency_type !== "incompatible",
  );
  const selectedDeps = depProjectIds
    ? allDeps.filter((d) => depProjectIds.includes(d.project_id))
    : allDeps.filter((d) => d.dependency_type === "required");

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
      if (depFile) {
        await downloadFromURL(depFile.url, depFile.filename);
      }
    }
  }

  await downloadFromURL(file.url, file.filename);
  await refresh();
  useToast().show("success", "Mod installed", {
    description: file.filename,
  });
}

async function handleLoadVersions(projectId: string) {
  const list = await modrinth.getVersions(
    projectId,
    serverEngine.value,
    serverGameVersion.value,
  );
  if (list.length > 0) {
    versionsMap.value = { ...versionsMap.value, [projectId]: list };
  }
}

async function handleLoadVersionDetails(versionId: string) {
  const details = await modrinth.getVersionDetails(versionId);
  if (details) {
    versionDetailsMap.value = {
      ...versionDetailsMap.value,
      [versionId]: details,
    };
  }
}

async function handleLoadProject(projectId: string) {
  const project = await modrinth.getProject(projectId);
  if (project) {
    depProjectsMap.value = { ...depProjectsMap.value, [projectId]: project };
  }
}

onMounted(async () => {
  await fetchServerInfo();
  if (!serverEngine.value) return;
  await refresh();
  await modrinth.search(serverEngine.value, serverGameVersion.value);
});
</script>
