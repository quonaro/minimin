<template>
  <div class="p-6">
    <div class="mb-6">
      <NuxtLink
        :to="`/agent/${agentId}/servers/${serverId}`"
        class="inline-flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors mb-4"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="w-4 h-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 19l-7-7m0 0l7-7m-7 7h18"
          />
        </svg>
        <span>Back to server</span>
      </NuxtLink>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Settings</h1>
    </div>
    <div class="text-gray-500 dark:text-gray-400">
      Server settings coming soon.
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
</script>
