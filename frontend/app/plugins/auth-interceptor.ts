export default defineNuxtPlugin(() => {
  const { isAuthenticated } = useAuth()

  globalThis.$fetch = $fetch.create({
    onResponseError({ response }) {
      if (response.status === 401 && isAuthenticated.value) {
        isAuthenticated.value = false
        navigateTo('/login', { replace: true })
      }
    },
  })
})
