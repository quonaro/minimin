<script setup lang="ts">
const { isAuthenticated } = useAuth();

if (isAuthenticated.value) {
  await navigateTo("/agents", { replace: true });
} else {
  await navigateTo("/login", { replace: true });
}
</script>
