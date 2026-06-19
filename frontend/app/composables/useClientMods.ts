import type { ModInfo } from "./useMods";
import { useUploadQueue } from "./useUploadQueue";

export interface ArchiveInfo {
  token: string;
  expiresAt: string;
  serverName: string;
  formats: string[];
}

export interface ArchiveLinkEntry {
  token: string;
  serverName: string;
  expiresAt: string;
  createdAt: string;
  formats: string[];
  downloadCounts: Record<string, number>;
  totalDownloads: number;
}

export function useClientMods(serverId: string) {
  const { show } = useToast();
  const { upload } = useUploadQueue();

  const mods = ref<ModInfo[]>([]);
  const loading = ref(false);
  const uploadLoading = ref(false);
  const downloadLoading = ref(false);
  const archiveLoading = ref(false);

  const url = computed(() => `/servers/${serverId}/client-mods`);

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
      show("error", "Failed to load client mods", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
    } finally {
      loading.value = false;
    }
  }

  async function deleteMod(filename: string) {
    try {
      await $fetch(
        `/servers/${serverId}/client-mods/${encodeURIComponent(filename)}`,
        {
          baseURL: useApiBase(),
          method: "DELETE",
          credentials: "include",
        },
      );
      show("success", "Client mod deleted", { description: filename });
      await refresh();
    } catch (err: any) {
      show("error", "Failed to delete client mod", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
    }
  }

  async function uploadFile(file: File) {
    uploadLoading.value = true;
    try {
      await upload(file, `${useApiBase()}/servers/${serverId}/client-mods/upload`, {
        method: "POST",
        withCredentials: true,
      });
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
      }>(`/servers/${serverId}/client-mods/download`, {
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
        description: getApiErrorMessage(err, "Unknown error"),
      });
    } finally {
      downloadLoading.value = false;
    }
  }

  async function toggleMod(filename: string) {
    try {
      const res = await $fetch<{
        body?: { filename: string; enabled: boolean };
      }>(`/servers/${serverId}/client-mods/${encodeURIComponent(filename)}/toggle`, {
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
        description: getApiErrorMessage(err, "Unknown error"),
      });
    }
  }

  async function deleteMods(filenames: string[]) {
    if (filenames.length === 0) return;
    try {
      await Promise.all(
        filenames.map((filename) =>
          $fetch(`/servers/${serverId}/client-mods/${encodeURIComponent(filename)}`, {
            baseURL: useApiBase(),
            method: "DELETE",
            credentials: "include",
          }),
        ),
      );
      show("success", `Deleted ${filenames.length} mod(s)`);
      await refresh();
    } catch (err: any) {
      show("error", "Failed to delete client mods", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
    }
  }

  async function toggleMods(filenames: string[]) {
    if (filenames.length === 0) return;
    try {
      await Promise.all(
        filenames.map((filename) =>
          $fetch(`/servers/${serverId}/client-mods/${encodeURIComponent(filename)}/toggle`, {
            baseURL: useApiBase(),
            method: "POST",
            credentials: "include",
          }),
        ),
      );
      show("success", `Toggled ${filenames.length} mod(s)`);
      await refresh();
    } catch (err: any) {
      show("error", "Failed to toggle client mods", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
    }
  }

  async function moveMod(filename: string, target: "server" | "client") {
    try {
      await $fetch(`/servers/${serverId}/client-mods/move`, {
        baseURL: useApiBase(),
        method: "POST",
        body: { filename, target },
        credentials: "include",
      });
      show("success", target === "client" ? "Moved to client" : "Moved to server", {
        description: filename,
      });
      await refresh();
    } catch (err: any) {
      show("error", "Failed to move mod", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
    }
  }

  async function copyMod(filename: string, source: "server" | "client", target: "server" | "client") {
    try {
      await $fetch(`/servers/${serverId}/mods/copy`, {
        baseURL: useApiBase(),
        method: "POST",
        body: { filename, source, target },
        credentials: "include",
      });
      show("success", target === "client" ? "Copied to client" : "Copied to server", {
        description: filename,
      });
      await refresh();
    } catch (err: any) {
      show("error", "Failed to copy mod", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
    }
  }

  async function createArchive(
    ttl: number,
    include: string[] = ["mods"],
  ): Promise<ArchiveInfo | null> {
    archiveLoading.value = true;
    try {
      const res = await $fetch<{
        body?: ArchiveInfo;
        token?: string;
        expiresAt?: string;
        serverName?: string;
        formats?: string[];
      }>(`/servers/${serverId}/client-mods/archive`, {
        baseURL: useApiBase(),
        method: "POST",
        body: { ttl, include },
        credentials: "include",
      });
      const data = (res as any).body ?? res;
      show("success", "Archive generated", {
        description: `Expires in ${ttl}h`,
      });
      return {
        token: data.token ?? "",
        expiresAt: data.expiresAt ?? "",
        serverName: data.serverName ?? "",
        formats: data.formats ?? [],
      };
    } catch (err: any) {
      show("error", "Failed to create archive", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
      return null;
    } finally {
      archiveLoading.value = false;
    }
  }

  async function listArchives(): Promise<ArchiveLinkEntry[]> {
    try {
      const res = await $fetch<ArchiveLinkEntry[]>(
        `/servers/${serverId}/client-mods/archives`,
        {
          baseURL: useApiBase(),
          credentials: "include",
        },
      );
      return (res as any) ?? [];
    } catch (err: any) {
      show("error", "Failed to load archives", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
      return [];
    }
  }

  async function deleteArchive(token: string): Promise<boolean> {
    try {
      await $fetch(`/servers/${serverId}/client-mods/archives/${token}`, {
        baseURL: useApiBase(),
        method: "DELETE",
        credentials: "include",
      });
      show("success", "Archive deleted");
      return true;
    } catch (err: any) {
      show("error", "Failed to delete archive", {
        description: getApiErrorMessage(err, "Unknown error"),
      });
      return false;
    }
  }

  return {
    mods,
    loading,
    uploadLoading,
    downloadLoading,
    archiveLoading,
    refresh,
    deleteMod,
    deleteMods,
    uploadFile,
    downloadFromURL,
    toggleMod,
    toggleMods,
    moveMod,
    copyMod,
    createArchive,
    listArchives,
    deleteArchive,
  };
}
