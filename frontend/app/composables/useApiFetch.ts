export function useApiFetch<T = unknown>(url: string, options: any = {}) {
  return useFetch<T>(url, {
    baseURL: useApiBase(),
    credentials: 'include',
    ...options,
  })
}
