import { defineEventHandler, readBody, send, setHeader, setResponseStatus } from 'h3'

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

    // Forward auth_token cookie to backend
    const authToken = getCookie(event, 'auth_token')
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }
    if (authToken) {
      headers['Cookie'] = `auth_token=${authToken}`
    }

    // Handle different methods
    const method = event.method
    let body: any = undefined
    if (method === 'POST' || method === 'PUT' || method === 'PATCH') {
      body = await readBody(event)
    }

    const response = await fetch(target, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    })

    const data = await response.text()
    setResponseStatus(event, response.status)
    setHeader(event, 'Content-Type', 'application/json')
    return send(event, data)
  }
})
