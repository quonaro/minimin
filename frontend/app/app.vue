<template>
  <div class="flex min-h-screen bg-gray-50 dark:bg-gray-900">
    <Sidebar v-if="route.path !== '/login'" />
    <main :class="route.path === '/login' ? 'flex-1' : 'flex-1 ml-64'">
      <NuxtRouteAnnouncer />
      <NuxtPage />
    </main>
  </div>
</template>

<script setup lang="ts">
const route = useRoute();
</script>
