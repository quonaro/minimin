interface LoginResponse {
  success: boolean
}

export const useAuth = () => {
  const isAuthenticated = useCookie('is_authenticated', {
    default: () => false,
    maxAge: 60 * 60 * 24, // 24 hours
  })

  const login = async (key: string): Promise<{ success: boolean; error?: string }> => {
    try {
      const response = await $fetch<LoginResponse>('/api/auth/login', {
        method: 'POST',
        body: { apiKey: key },
        credentials: 'include', // Important for httpOnly cookies
      })

      if (response.success) {
        isAuthenticated.value = true
        return { success: true }
      }
      return { success: false, error: 'Unexpected response from server' }
    } catch (error) {
      console.error('Login failed:', error)
      return { success: false, error: getApiErrorMessage(error, 'Invalid password or connection failed') }
    }
  }

  const logout = async () => {
    isAuthenticated.value = false
    // Clear httpOnly cookie by calling logout endpoint
    try {
      await $fetch('/api/auth/logout', {
        method: 'POST',
        credentials: 'include',
      })
    } catch {
      // Ignore errors, just clear local state
    }
    navigateTo('/login')
  }

  return {
    isAuthenticated,
    login,
    logout,
  }
}
