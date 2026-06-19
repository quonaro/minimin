import { useUploadQueue } from "./useUploadQueue";

export interface ClientAsset {
  filename: string;
  size: number;
  enabled: boolean;
}

export function useClientAssets(serverId: string, type: "resourcepacks" | "shaderpacks") {
  const { show } = useToast();
  const { upload } = useUploadQueue();

  const assets = ref<ClientAsset[]>([]);
  const loading = ref(false);
  const uploadLoading = ref(false);
  const downloadLoading = ref(false);

  const url = computed(() => `/servers/${serverId}/client-assets?type=${type}`);

  async function refresh() {
    loading.value = true;
    try {
      const res = await $fetch<{ body?: { assets: ClientAsset[] }; assets?: ClientAsset[] }>(url.value, {
        baseURL: useApiBase(),
        credentials: "include",
      });
      const list = (res as any).body?.assets ?? (res as any).assets ?? [];
      assets.value = list as ClientAsset[];
    } catch (err: any) {
      show("error", `Failed to load ${type}`, {
        description: getApiErrorMessage(err, "Unknown error"),
      });
    } finally {
      loading.value = false;
    }
  }

  async function deleteAsset(filename: string) {
    try {
      await $fetch(
        `/servers/${serverId}/client-assets/${encodeURIComponent(filename)}?type=${type}`,
        {
          baseURL: useApiBase(),
          method: "DELETE",
          credentials: "include",
        },
      );
      show("success", "Deleted", { description: filename });
      await refresh();
    } catch (err: any) {
      show("error", "Failed to delete", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
    }
  }

  async function uploadFile(file: File) {
    uploadLoading.value = true;
    try {
      await upload(
        file,
        `${useApiBase()}/servers/${serverId}/client-assets/upload?type=${type}`,
        {
          method: "POST",
          withCredentials: true,
        },
      );
      show("success", "Upload complete", { description: file.name });
      await refresh();
    } catch (err: any) {
      show("error", "Upload failed", {
        description: getApiErrorMessage(err, "Unknown error"),
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
      }>(`/servers/${serverId}/client-assets/download?type=${type}`, {
        baseURL: useApiBase(),
        method: "POST",
        credentials: "include",
        body: { url, filename },
      });
      const fn =
        (result as any).body?.filename || result?.filename || filename || "file";
      show("success", "Download complete", { description: fn });
      await refresh();
    } catch (err: any) {
      show("error", "Download failed", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
    } finally {
      downloadLoading.value = false;
    }
  }

  async function toggleAsset(filename: string) {
    try {
      const res = await $fetch<{
        body?: { filename: string; enabled: boolean };
        filename?: string;
        enabled?: boolean;
      }>(`/servers/${serverId}/client-assets/${encodeURIComponent(filename)}/toggle?type=${type}`, {
        baseURL: useApiBase(),
        method: "POST",
        credentials: "include",
      });
      const data = (res as any).body ?? res;
      show("success", data.enabled ? "Enabled" : "Disabled", {
        description: data.filename,
      });
      await refresh();
    } catch (err: any) {
      show("error", "Failed to toggle", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
    }
  }

  return {
    assets,
    loading,
    uploadLoading,
    downloadLoading,
    refresh,
    deleteAsset,
    uploadFile,
    downloadFromURL,
    toggleAsset,
  };
}
