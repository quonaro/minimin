import { defineEventHandler, getCookie, getRequestHeaders, send, setHeader, setResponseStatus } from 'h3'

export default defineEventHandler(async (event) => {
  const path = event.path

  // Proxy /api/* requests to the backend (excluding Nitro's own /api routes)
  if (path.startsWith('/api/') && !path.startsWith('/api/_nitro')) {
    // Auth and events endpoints keep /api/ prefix on backend, others strip it
    let targetPath = path
    if (!path.startsWith('/api/auth/') && !path.startsWith('/api/events')) {
      targetPath = path.replace('/api/', '/')
    }
    const target = `http://localhost:8081${targetPath}`

    // Forward relevant request headers
    const headers: Record<string, string> = {}
    const reqHeaders = getRequestHeaders(event)
    for (const [key, value] of Object.entries(reqHeaders)) {
      if (key === 'host' || key === 'connection') continue
      if (value === undefined) continue
      headers[key] = Array.isArray(value) ? value.join(', ') : String(value)
    }

    const authToken = getCookie(event, 'auth_token')
    if (authToken) {
      headers['cookie'] = `auth_token=${authToken}`
    }

    const fetchOpts: RequestInit = {
      method: event.method,
      headers,
    }

    if (event.method !== 'GET' && event.method !== 'HEAD') {
      // @ts-ignore Node.js readable stream as body with duplex: 'half'
      fetchOpts.body = event.node.req
      // @ts-ignore
      fetchOpts.duplex = 'half'
    }

    const response = await fetch(target, fetchOpts)

    setResponseStatus(event, response.status)
    response.headers.forEach((value, key) => {
      if (key === 'content-encoding' || key === 'transfer-encoding') return
      setHeader(event, key, value)
    })

    const data = await response.arrayBuffer()
    return send(event, Buffer.from(data))
  }
})
