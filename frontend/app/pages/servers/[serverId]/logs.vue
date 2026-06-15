<template>
  <div class="flex flex-col h-[calc(100vh-4rem)]">
    <div class="flex-1 min-h-0 min-w-0 px-6 pt-6 pb-6 flex flex-col">
      <div class="mb-4">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Logs</h1>
        <p class="text-sm text-gray-500 dark:text-neutral-400 mt-1">
          Browse and filter server log files.
        </p>
      </div>
      <server-logs :server-id="serverId" class="flex-1 min-h-0 min-w-0" />
    </div>
  </div>
</template>

<script setup lang="ts">
usePageTitle("Logs");

definePageMeta({
  middleware: "auth",
});

const route = useRoute();
const serverId = route.params.serverId as string;
</script>
