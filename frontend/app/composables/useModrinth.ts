export interface ModrinthProject {
  project_id: string;
  slug: string;
  title: string;
  description: string;
  icon_url: string;
  author: string;
  downloads: number;
}

export interface ModrinthVersion {
  id: string;
  project_id: string;
  version_number: string;
  game_versions: string[];
  loaders: string[];
  files: Array<{
    url: string;
    filename: string;
    primary: boolean;
  }>;
}

export interface BulkJob {
  id: string;
  status: "pending" | "running" | "done" | "failed";
  completed: number;
  total: number;
  errors: string[];
  updatedAt: string;
}

export function useModrinth() {
  const { show } = useToast();

  const searchQuery = ref("");
  const searchResults = ref<ModrinthProject[]>([]);
  const searchLoading = ref(false);

  const installLoading = ref<Record<string, boolean>>({});
  const bulkJobId = ref<string | null>(null);
  const bulkJob = ref<BulkJob | null>(null);
  const bulkPollInterval = ref<ReturnType<typeof setInterval> | null>(null);

  async function search(loader: string, gameVersion: string) {
    if (!searchQuery.value.trim()) {
      searchResults.value = [];
      return;
    }
    searchLoading.value = true;
    try {
      const res = await $fetch<{ body?: { results: ModrinthProject[] }; results?: ModrinthProject[] }>(
        "/mods/search",
        {
          baseURL: useApiBase(),
          credentials: "include",
          query: {
            q: searchQuery.value,
            loader: loader.toLowerCase(),
            game_version: gameVersion,
            limit: 20,
          },
        },
      );
      searchResults.value = (res as any).body?.results ?? (res as any).results ?? [];
    } catch (err: any) {
      show("error", "Search failed", {
        description: err?.data?.detail || err?.message || "Unknown error",
      });
    } finally {
      searchLoading.value = false;
    }
  }

  async function getVersions(
    projectId: string,
    loader: string,
    gameVersion: string,
  ): Promise<ModrinthVersion[]> {
    try {
      const res = await $fetch<{
        body?: { versions: ModrinthVersion[] };
        versions?: ModrinthVersion[];
      }>(`/mods/versions/${projectId}`, {
        baseURL: useApiBase(),
        credentials: "include",
        query: {
          loader: loader.toLowerCase(),
          game_version: gameVersion,
        },
      });
      return (
        ((res as any).body?.versions ?? (res as any).versions ?? []) as ModrinthVersion[]
      );
    } catch (err: any) {
      show("error", "Failed to load versions", {
        description: err?.data?.detail || err?.message || "Unknown error",
      });
      return [];
    }
  }

  async function install(
    agentId: string,
    serverId: string,
    projectId: string,
    versionId: string,
  ) {
    const key = `${projectId}:${versionId}`;
    installLoading.value[key] = true;
    try {
      await $fetch("/mods/install", {
        baseURL: useApiBase(),
        method: "POST",
        credentials: "include",
        body: { agentId, serverId, projectId, versionId },
      });
      show("success", "Mod installed", { description: projectId });
    } catch (err: any) {
      show("error", "Install failed", {
        description: err?.data?.detail || err?.message || "Unknown error",
      });
    } finally {
      installLoading.value[key] = false;
    }
  }

  async function bulkInstall(
    agentId: string,
    serverId: string,
    items: Array<{ projectId: string; versionId: string }>,
  ) {
    try {
      const res = await $fetch<{ body?: { jobId: string }; jobId?: string }>(
        "/mods/bulk",
        {
          baseURL: useApiBase(),
          method: "POST",
          credentials: "include",
          body: { agentId, serverId, items },
        },
      );
      const jobId = (res as any).body?.jobId ?? (res as any).jobId;
      if (jobId) {
        bulkJobId.value = jobId;
        startPolling(jobId);
        show("info", "Bulk install started", { description: `Job ${jobId}` });
      }
    } catch (err: any) {
      show("error", "Bulk install failed", {
        description: err?.data?.detail || err?.message || "Unknown error",
      });
    }
  }

  function startPolling(jobId: string) {
    if (bulkPollInterval.value) {
      clearInterval(bulkPollInterval.value);
    }
    bulkPollInterval.value = setInterval(async () => {
      try {
        const res = await $fetch<{ body?: BulkJob }>(`/mods/jobs/${jobId}`, {
          baseURL: useApiBase(),
          credentials: "include",
        });
        const job = (res as any).body ?? res;
        bulkJob.value = job as BulkJob;
        if (job.status === "done" || job.status === "failed") {
          if (bulkPollInterval.value) {
            clearInterval(bulkPollInterval.value);
            bulkPollInterval.value = null;
          }
          if (job.status === "done") {
            show("success", "Bulk install complete");
          } else {
            show("error", "Bulk install failed", {
              description: job.errors?.join(", ") || "Unknown error",
            });
          }
        }
      } catch {
        // ignore polling errors
      }
    }, 2000);
  }

  onBeforeUnmount(() => {
    if (bulkPollInterval.value) {
      clearInterval(bulkPollInterval.value);
    }
  });

  return {
    searchQuery,
    searchResults,
    searchLoading,
    installLoading,
    bulkJobId,
    bulkJob,
    search,
    getVersions,
    install,
    bulkInstall,
  };
}
