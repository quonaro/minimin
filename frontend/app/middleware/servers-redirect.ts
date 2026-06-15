export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path !== '/') return;

  try {
    const res = await $fetch('/servers', {
      baseURL: useApiBase(),
      credentials: 'include',
    });
    const list = Array.isArray(res) ? res : ((res as any).body || []);
    if (list.length > 0) {
      const first = list[0];
      if (first?.serverId) {
        return navigateTo(`/servers/${first.serverId}`, { replace: true });
      }
    }
  } catch {
    // ignore; fall through to empty-state page
  }
});
