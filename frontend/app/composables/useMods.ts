export interface ModInfo {
  filename: string;
  name: string;
  version: string;
  modid: string;
  authors: string;
  size: number;
}

export function useMods(agentId: Ref<string | undefined>, serverId: string) {
  const { show } = useToast();

  const mods = ref<ModInfo[]>([]);
  const loading = ref(false);
  const uploadLoading = ref(false);

  const url = computed(() => {
    if (!agentId.value) return "";
    return `/agent/${agentId.value}/servers/${serverId}/mods`;
  });

  async function refresh() {
    if (!url.value) return;
    loading.value = true;
    try {
      const res = await $fetch<{ body?: { mods: ModInfo[] }; mods?: ModInfo[] }>(
        url.value,
        { baseURL: useApiBase(), credentials: "include" },
      );
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
    if (!agentId.value) return;
    try {
      await $fetch(
        `/agent/${agentId.value}/servers/${serverId}/mods/${encodeURIComponent(filename)}`,
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
    if (!agentId.value) return;
    uploadLoading.value = true;
    try {
      const formData = new FormData();
      formData.append("file", file);
      await $fetch(`/agent/${agentId.value}/servers/${serverId}/mods/upload`, {
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

  return {
    mods,
    loading,
    uploadLoading,
    refresh,
    deleteMod,
    uploadFile,
  };
}
