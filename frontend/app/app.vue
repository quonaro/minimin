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
    <ToastContainer />
    <UploadProgress v-if="!route.meta.noUploads" />
  </div>
</template>

<script setup lang="ts">
const route = useRoute();

const hasSidebar = computed(() => {
  return !route.meta.noSidebar;
});
</script>

<style>
* {
  transition-property:
    color, background-color, border-color, text-decoration-color, fill, stroke;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 200ms;
}

::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background-color: #c1c1c1;
  border-radius: 9999px;
  border: 2px solid transparent;
  background-clip: padding-box;
}

::-webkit-scrollbar-thumb:hover {
  background-color: #a8a8a8;
}

.dark ::-webkit-scrollbar-thumb {
  background-color: #4b5563;
}

.dark ::-webkit-scrollbar-thumb:hover {
  background-color: #6b7280;
}

html {
  scrollbar-width: thin;
  scrollbar-color: #c1c1c1 transparent;
}

html.dark {
  scrollbar-color: #4b5563 transparent;
}
</style>
