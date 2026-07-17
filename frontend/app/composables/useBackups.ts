export interface Backup {
  name: string;
  serverId: string;
  sizeBytes: number;
  createdAt: string;
}

export function useBackups(serverId: string) {
  const backups = ref<Backup[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const { show } = useToast();

  async function fetchBackups() {
    loading.value = true;
    error.value = null;
    try {
      const res = await $fetch<Backup[]>(`/servers/${serverId}/backups`, {
        baseURL: useApiBase(),
        credentials: "include",
      });
      backups.value = res || [];
    } catch (e: any) {
      error.value = getApiErrorMessage(e, "Failed to load backups");
    } finally {
      loading.value = false;
    }
  }

  async function createBackup(): Promise<Backup> {
    const res = await $fetch<Backup>(`/servers/${serverId}/backups`, {
      baseURL: useApiBase(),
      method: "POST",
      credentials: "include",
    });
    show("info", "Backup started");
    await fetchBackups();
    return res;
  }

  async function restoreBackup(name: string) {
    try {
      await $fetch(`/servers/${serverId}/backups/${encodeURIComponent(name)}/restore`, {
        baseURL: useApiBase(),
        method: "POST",
        credentials: "include",
      });
      show("success", "Backup restored");
    } catch (e: any) {
      show("error", getApiErrorMessage(e, "Failed to restore backup"));
    }
  }

  async function deleteBackup(name: string) {
    try {
      await $fetch(`/servers/${serverId}/backups/${encodeURIComponent(name)}`, {
        baseURL: useApiBase(),
        method: "DELETE",
        credentials: "include",
      });
      show("success", "Backup deleted");
      await fetchBackups();
    } catch (e: any) {
      show("error", getApiErrorMessage(e, "Failed to delete backup"));
    }
  }

  function getDownloadUrl(name: string): string {
    const origin = typeof window !== "undefined" ? window.location.origin : "";
    return `${origin}${useApiBase()}/servers/${serverId}/backups/${encodeURIComponent(name)}`;
  }

  async function downloadBackup(name: string) {
    const url = getDownloadUrl(name);
    const response = await fetch(url, { credentials: "include" });
    if (!response.ok) {
      throw new Error(`Download failed: ${response.status} ${response.statusText}`);
    }
    const blob = await response.blob();
    const objectUrl = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = objectUrl;
    a.download = name;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(objectUrl);
  }

  async function copyDownloadLink(name: string) {
    const url = getDownloadUrl(name);
    try {
      await navigator.clipboard.writeText(url);
      show("info", "Direct link copied to clipboard");
    } catch {
      show("error", "Failed to copy link");
    }
  }

  return {
    backups,
    loading,
    error,
    fetchBackups,
    createBackup,
    restoreBackup,
    deleteBackup,
    getDownloadUrl,
    downloadBackup,
    copyDownloadLink,
  };
}
