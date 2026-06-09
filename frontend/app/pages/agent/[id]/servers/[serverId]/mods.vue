<template>
  <div class="flex flex-col h-[calc(100vh-4rem)]">
    <div class="flex-1 min-h-0 px-6 pt-6 pb-6 flex flex-col">
      <div class="mb-4">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Mods</h1>
        <p class="text-sm text-gray-500 dark:text-neutral-400 mt-1">
          Manage installed mods and browse Modrinth library.
        </p>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 flex-1 min-h-0">
        <installed-mods
          :mods="mods"
          :loading="loading"
          :upload-loading="uploadLoading"
          @delete="deleteMod"
          @upload="handleUpload"
        />
        <mod-library
          :search-query="modrinth.searchQuery.value"
          :search-results="modrinth.searchResults.value"
          :search-loading="modrinth.searchLoading.value"
          :loader="serverEngine"
          :game-version="serverGameVersion"
          :install-loading="modrinth.installLoading.value"
          :versions="versionsMap"
          @update:search-query="(v) => (modrinth.searchQuery.value = v)"
          @search="modrinth.search(serverEngine, serverGameVersion)"
          @install="handleInstall"
          @load-versions="handleLoadVersions"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

const route = useRoute();
const { agentId } = useCurrentAgent();
const serverId = route.params.serverId as string;

const { mods, loading, uploadLoading, refresh, deleteMod, uploadFile } =
  useMods(agentId, serverId);

const modrinth = useModrinth();

const serverEngine = ref("");
const serverGameVersion = ref("");
const versionsMap = ref<
  Record<string, import("~/composables/useModrinth").ModrinthVersion[]>
>({});

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
