export interface ModInfo {
  filename: string;
  name: string;
  version: string;
  modid: string;
  authors: string;
  size: number;
  description?: string;
  icon?: string;
  enabled?: boolean;
  environment?: string;
  corrupted?: boolean;
}

export function useMods(serverId: string) {
  const { show } = useToast();

  const mods = ref<ModInfo[]>([]);
  const loading = ref(false);
  const uploadLoading = ref(false);
  const downloadLoading = ref(false);

  const url = computed(() => `/servers/${serverId}/mods`);

  async function refresh() {
    loading.value = true;
    try {
      const res = await $fetch<{
        body?: { mods: ModInfo[] };
        mods?: ModInfo[];
      }>(url.value, { baseURL: useApiBase(), credentials: "include" });
      const list = (res as any).body?.mods ?? (res as any).mods ?? [];
      mods.value = list as ModInfo[];
    } catch (err: any) {
      show("error", "Failed to load mods", {
        description: err?.data?.detail || err?.message || "Unknown error",
      });
    } finally {
      loading.value = false;
    }
  }

  async function deleteMod(filename: string) {
    try {
      await $fetch(
        `/servers/${serverId}/mods/${encodeURIComponent(filename)}`,
        {
          baseURL: useApiBase(),
          method: "DELETE",
          credentials: "include",
        },
      );
      show("success", "Mod deleted", { description: filename });
      await refresh();
    } catch (err: any) {
      show("error", "Failed to delete mod", {
        description: err?.data?.detail || err?.message || "Unknown error",
      });
    }
  }

  async function uploadFile(file: File) {
    uploadLoading.value = true;
    try {
      const formData = new FormData();
      formData.append("file", file);
      await $fetch(`/servers/${serverId}/mods/upload`, {
        baseURL: useApiBase(),
        method: "POST",
        body: formData,
        credentials: "include",
      });
      show("success", "Upload complete", { description: file.name });
      await refresh();
    } catch (err: any) {
      show("error", "Upload failed", {
        description: err?.data?.detail || err?.message || "Unknown error",
      });
    } finally {
      uploadLoading.value = false;
    }
  }

  async function downloadFromURL(url: string, filename?: string) {
    downloadLoading.value = true;
    try {
      const result = await $fetch<{
        body?: { success: boolean; filename?: string };
        filename?: string;
      }>(`/servers/${serverId}/mods/download`, {
        baseURL: useApiBase(),
        method: "POST",
        credentials: "include",
        body: {
          url,
          filename,
        },
      });
      const fn =
        (result as any).body?.filename ||
        result?.filename ||
        filename ||
        "file";
      show("success", "Download complete", { description: fn });
      await refresh();
    } catch (err: any) {
      show("error", "Download failed", {
        description: err?.data?.detail || err?.message || "Unknown error",
      });
    } finally {
      downloadLoading.value = false;
    }
  }

  async function toggleMod(filename: string) {
    try {
      const res = await $fetch<{
        body?: { filename: string; enabled: boolean };
      }>(`/servers/${serverId}/mods/${encodeURIComponent(filename)}/toggle`, {
        baseURL: useApiBase(),
        method: "POST",
        credentials: "include",
      });
      const data = (res as any).body ?? res;
      show("success", data.enabled ? "Mod enabled" : "Mod disabled", {
        description: data.filename,
      });
      await refresh();
    } catch (err: any) {
      show("error", "Failed to toggle mod", {
        description: err?.data?.detail || err?.message || "Unknown error",
      });
    }
  }

  return {
    mods,
    loading,
    uploadLoading,
    downloadLoading,
    refresh,
    deleteMod,
    uploadFile,
    downloadFromURL,
    toggleMod,
  };
}
