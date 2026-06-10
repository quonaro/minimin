export function useApiFetch<T = unknown>(url: string | Ref<string> | (() => string), options: any = {}) {
  const authCookie = useCookie('is_authenticated', { default: () => false })

  return useFetch<T>(url, {
    baseURL: useApiBase(),
    credentials: 'include',
    onResponseError({ response }) {
      if (response.status === 401) {
        authCookie.value = false
        navigateTo('/login')
      }
      if (options.onResponseError) {
        options.onResponseError({ response })
      }
    },
    ...options,
  })
}
