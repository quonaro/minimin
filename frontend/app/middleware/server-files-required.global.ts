export default defineNuxtRouteMiddleware(async (to) => {
  if (import.meta.server) {
    return
  }

  const agentId = to.params.id
  const serverId = to.params.serverId

  if (typeof agentId !== 'string' || typeof serverId !== 'string') {
    return
  }

  const overviewPath = `/agent/${agentId}/servers/${serverId}`
  if (to.path === overviewPath || to.path === `${overviewPath}/`) {
    return
  }

  if (!to.path.startsWith(`${overviewPath}/`)) {
    return
  }

  try {
    const res = await $fetch<
      { initialized?: boolean } | { body?: { initialized?: boolean } }
    >(`/agent/${agentId}/servers/${serverId}/config`, {
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
