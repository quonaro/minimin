<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

const route = useRoute();
await navigateTo(`/agent/${route.params.id}/servers`, { replace: true });
</script>
