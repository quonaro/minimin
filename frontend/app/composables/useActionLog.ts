export interface ActionLogEntry {
  id: string;
  timestamp: string;
  serverId: string;
  taskId?: string;
  source: string;
  action: string;
  detail?: string;
  status: string;
  message?: string;
}

export function useActionLog(serverId: string) {
  const entries = ref<ActionLogEntry[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchLog() {
    loading.value = true;
    error.value = null;
    try {
      const res = await $fetch<ActionLogEntry[]>(`/servers/${serverId}/actions-log`, {
        baseURL: useApiBase(),
        credentials: "include",
      });
      entries.value = res || [];
    } catch (e: any) {
      error.value = getApiErrorMessage(e, "Failed to load action log");
    } finally {
      loading.value = false;
    }
  }

  return { entries, loading, error, fetchLog };
}
