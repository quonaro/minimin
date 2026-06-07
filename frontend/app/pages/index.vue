<script setup lang="ts">
const { isAuthenticated } = useAuth();

if (isAuthenticated.value) {
  await navigateTo("/servers", { replace: true });
} else {
  await navigateTo("/login", { replace: true });
}
</script>
