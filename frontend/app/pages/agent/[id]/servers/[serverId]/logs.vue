<template>
  <div class="flex flex-col h-[calc(100vh-4rem)]">
    <div class="flex-1 min-h-0 px-6 pt-6 pb-6 flex flex-col">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-4">
        Logs
      </h1>
      <server-logs :server-id="serverId" class="flex-1 min-h-0" />
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

const route = useRoute();
const serverId = route.params.serverId as string;
</script>
