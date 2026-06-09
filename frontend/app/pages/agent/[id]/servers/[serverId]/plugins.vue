<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Plugins</h1>
    </div>
    <div class="text-gray-500 dark:text-neutral-400">
      Plugins management coming soon.
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute();
const { agentId } = useCurrentAgent();
const serverId = route.params.serverId as string;

definePageMeta({
  middleware: "auth",
});

async function fetchServerInfo() {
  if (!agentId.value) return;
  try {
    const res = await $fetch<
      { body: { engineType: string } } | { engineType: string }
    >(`/agent/${agentId.value}/servers/${serverId}`, {
      baseURL: useApiBase(),
      credentials: "include",
    });
    const data = (res as any).body ?? res;
    const engine = (data.engineType ?? "").toUpperCase();

    if (engine === "FABRIC" || engine === "FORGE") {
      await navigateTo(`/agent/${agentId.value}/servers/${serverId}/mods`, {
        replace: true,
      });
      return;
    }
    if (engine === "VANILLA") {
      await navigateTo(`/agent/${agentId.value}/servers/${serverId}`, {
        replace: true,
      });
      return;
    }
  } catch {
    // ignore
  }
}

onMounted(() => {
  fetchServerInfo();
});
</script>
