export default defineNuxtRouteMiddleware(async (to) => {
  if (import.meta.server) {
    return
  }

  const serverId = to.params.serverId

  if (typeof serverId !== 'string') {
    return
  }

  const overviewPath = `/servers/${serverId}`
  const terminalPath = `${overviewPath}/terminal`
  if (to.path === overviewPath || to.path === `${overviewPath}/` || to.path === terminalPath || to.path === `${terminalPath}/`) {
    return
  }

  if (!to.path.startsWith(`${overviewPath}/`)) {
    return
  }

  const { initialized, loading, refresh } = useServerConfig(serverId)

  if (loading.value) {
    await refresh()
  }

  if (!initialized.value) {
    return navigateTo(overviewPath)
  }
})
