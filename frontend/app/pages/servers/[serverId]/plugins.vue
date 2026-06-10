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
const serverId = route.params.serverId as string;

definePageMeta({
  middleware: "auth",
});

async function fetchServerInfo() {
  try {
    const res = await $fetch<
      { body: { engineType: string } } | { engineType: string }
    >(`/servers/${serverId}`, {
      baseURL: useApiBase(),
      credentials: "include",
    });
    const data = (res as any).body ?? res;
    const engine = (data.engineType ?? "").toUpperCase();

    if (engine === "FABRIC" || engine === "FORGE") {
      await navigateTo(`/servers/${serverId}/mods`, {
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

onMounted(() => {
  fetchServerInfo();
});
</script>
