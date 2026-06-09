<template>
  <div class="p-6">
    <server-nav class="mb-4" />
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Plugins</h1>
    </div>
    <div class="text-gray-500 dark:text-neutral-400">
      Plugins management coming soon.
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});
</script>
