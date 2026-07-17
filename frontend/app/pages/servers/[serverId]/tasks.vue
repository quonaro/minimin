<script setup lang="ts">
definePageMeta({
  middleware: [
    (to) => {
      return navigateTo(`/servers/${to.params.serverId}/actions`, {
        replace: true,
      });
    },
  ],
});
</script>
