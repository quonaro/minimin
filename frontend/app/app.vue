<template>
  <div class="flex min-h-screen bg-gray-50 dark:bg-background-dark">
    <Sidebar v-if="hasSidebar" />
    <main
      :class="{
        'flex-1': !hasSidebar,
        'flex-1 ml-64': hasSidebar,
      }"
    >
      <NuxtRouteAnnouncer />
      <NuxtPage />
    </main>
  </div>
</template>

<script setup lang="ts">
const route = useRoute();

const hasSidebar = computed(() => route.path !== "/login");
</script>

<style>
* {
  transition-property:
    color, background-color, border-color, text-decoration-color, fill, stroke;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 200ms;
}
</style>
