<template>
  <terminal-view :server-id="serverId" />
</template>

<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

usePageTitle("Terminal");

const route = useRoute();
const serverId = route.params.serverId as string;
</script>
