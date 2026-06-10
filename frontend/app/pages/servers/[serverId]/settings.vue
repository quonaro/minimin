<script setup lang="ts">
definePageMeta({ middleware: "auth" });
const route = useRoute();
await navigateTo(`/servers/${route.params.serverId}`);
</script>
