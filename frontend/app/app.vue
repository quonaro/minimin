<template>
  <div class="flex min-h-screen bg-gray-50 dark:bg-gray-900">
    <Sidebar v-if="showGlobalSidebar" />
    <AgentSidebar v-else-if="showAgentSidebar" />
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
import AgentSidebar from "./components/AgentSidebar.vue";

const route = useRoute();

const showGlobalSidebar = computed(() => {
  return (
    route.path !== "/login" &&
    !route.path.startsWith("/agent") &&
    !route.path.startsWith("/agents")
  );
});

const showAgentSidebar = computed(() => {
  return route.path.startsWith("/agent");
});

const hasSidebar = computed(
  () => showGlobalSidebar.value || showAgentSidebar.value,
);
</script>
