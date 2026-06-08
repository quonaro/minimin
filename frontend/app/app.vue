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
  </div>
</template>

<script setup lang="ts">
const route = useRoute();
const { show: showToast } = useToast();

const hasSidebar = computed(() => route.path !== "/login");

const { lastEvent } = useEventSource();

watch(lastEvent, (evt) => {
  if (!evt) return;
  switch (evt.type) {
    case "server.status":
      showToast("info", `Server ${evt.newStatus}`, {
        description: `${evt.serverId} changed from ${evt.oldStatus} to ${evt.newStatus}`,
      });
      break;
    case "agent.status":
      showToast(
        "info",
        `Agent ${evt.newStatus === "online" ? "online" : "offline"}`,
        {
          description: evt.message,
        },
      );
      break;
    case "action.success":
      showToast("success", evt.message);
      break;
    case "action.error":
      showToast("error", evt.message);
      break;
    default:
      showToast("info", evt.message);
      break;
  }
});
</script>

<style>
* {
  transition-property:
    color, background-color, border-color, text-decoration-color, fill, stroke;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 200ms;
}
</style>
