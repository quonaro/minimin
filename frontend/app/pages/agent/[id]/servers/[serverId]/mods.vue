<template>
  <div class="p-6 h-[calc(100vh-4rem)] flex flex-col">
    <div class="mb-4">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Mods</h1>
      <p class="text-sm text-gray-500 dark:text-neutral-400 mt-1">
        Manage installed mods and browse Modrinth library.
      </p>
    </div>

    <div class="mb-4">
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
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 flex-1 min-h-0">
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden flex flex-col"
      >
        <div class="p-4 md:p-6 flex-1 min-h-0">
          <installed-mods
            :mods="mods"
            :loading="loading"
            @delete="deleteMod"
            @upload="handleUpload"
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
            @update:search-query="onSearchInput"
            @search="modrinth.search(serverEngine, serverGameVersion)"
            @install="handleInstall"
            @load-versions="handleLoadVersions"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { debounce } from "~/utils/debounce";
import { Upload } from "lucide-vue-next";

definePageMeta({
  middleware: "auth",
});

const route = useRoute();
const { agentId } = useCurrentAgent();
const serverId = route.params.serverId as string;

const { mods, loading, uploadLoading, refresh, deleteMod, uploadFile } =
  useMods(agentId, serverId);

const fileInput = ref<HTMLInputElement | null>(null);

function onFileSelect(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  uploadFile(file);
  input.value = "";
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
  if (!agentId.value) return;
  try {
    const res = await $fetch<
      | { body: { engineType: string; gameVersion: string } }
      | { engineType: string; gameVersion: string }
    >(`/agent/${agentId.value}/servers/${serverId}`, {
      baseURL: useApiBase(),
      credentials: "include",
    });
    const data = (res as any).body ?? res;
    serverEngine.value = data.engineType ?? "";
    serverGameVersion.value = data.gameVersion ?? "";
  } catch {
    // ignore
  }
}

async function handleUpload(file: File) {
  await uploadFile(file);
}

async function handleInstall(projectId: string, versionId: string) {
  if (!agentId.value) return;
  await modrinth.install(agentId.value, serverId, projectId, versionId);
  await refresh();
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

onMounted(async () => {
  await fetchServerInfo();
  await refresh();
});
</script>
