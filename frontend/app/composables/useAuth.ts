interface LoginResponse {
  success: boolean
}

export const useAuth = () => {
  const isAuthenticated = useCookie('is_authenticated', {
    default: () => false,
    maxAge: 60 * 60 * 24, // 24 hours
  })

  const login = async (key: string) => {
    try {
      const response = await $fetch<LoginResponse>('/api/auth/login', {
        method: 'POST',
        body: { api_key: key },
        credentials: 'include', // Important for httpOnly cookies
      })

      if (response.success) {
        isAuthenticated.value = true
        return true
      }
      return false
    } catch (error) {
      console.error('Login failed:', error)
      return false
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
