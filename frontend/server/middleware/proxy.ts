import { defineEventHandler, proxyRequest } from 'h3'

export default defineEventHandler(async (event) => {
  const path = event.path
  // Proxy /api/* requests to the backend (excluding Nitro's own /api routes)
  if (path.startsWith('/api/') && !path.startsWith('/api/_nitro')) {
    // Auth endpoints keep /api/ prefix on backend, others strip it
    let targetPath = path
    if (!path.startsWith('/api/auth/')) {
      targetPath = path.replace('/api/', '/')
    }
    const target = `http://localhost:8081${targetPath}`
    return proxyRequest(event, target)
  }
})
