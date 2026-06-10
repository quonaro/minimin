export default defineNuxtRouteMiddleware(async (to) => {
  if (import.meta.server) {
    return
  }

  const serverId = to.params.serverId

  if (typeof serverId !== 'string') {
    return
  }

  const overviewPath = `/servers/${serverId}`
  const logsPath = `${overviewPath}/logs`
  if (to.path === overviewPath || to.path === `${overviewPath}/` || to.path === logsPath || to.path === `${logsPath}/`) {
    return
  }

  if (!to.path.startsWith(`${overviewPath}/`)) {
    return
  }

  try {
    const res = await $fetch<
      { initialized?: boolean } | { body?: { initialized?: boolean } }
    >(`/servers/${serverId}/config`, {
      baseURL: useApiBase(),
      credentials: 'include',
    })

    const initializedRoot = (res as { initialized?: boolean })?.initialized
    const initializedBody = (res as { body?: { initialized?: boolean } })?.body
      ?.initialized
    const initialized: boolean = initializedRoot ?? initializedBody ?? true

    if (!initialized) {
      return navigateTo(overviewPath)
    }
  } catch {
    return
  }
})
