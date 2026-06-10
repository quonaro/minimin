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
  const searchOffset = ref(0);
  const searchLimit = 20;
  const hasMore = ref(true);

  const installLoading = ref<Record<string, boolean>>({});

  async function search(loader: string, gameVersion: string) {
    searchLoading.value = true;
    searchOffset.value = 0;
    hasMore.value = true;
    try {
      const facets = [`["categories:${loader.toLowerCase()}"]`, `["versions:${gameVersion}"]`];
      const res = await $fetch<any>(
        `https://api.modrinth.com/v2/search`,
        {
          query: {
            query: searchQuery.value,
            facets: `[${facets.join(",")}]`,
            offset: 0,
            limit: searchLimit,
          },
        },
      );
      const hits = res?.hits || [];
      searchResults.value = hits.map((h: any) => ({
        project_id: h.project_id,
        slug: h.slug,
        title: h.title,
        description: h.description,
        icon_url: h.icon_url,
        author: h.author,
        downloads: h.downloads,
      }));
      hasMore.value = hits.length === searchLimit;
    } catch (err: any) {
      show("error", "Search failed", {
        description: err?.message || "Unknown error",
      });
      searchResults.value = [];
      hasMore.value = false;
    } finally {
      searchLoading.value = false;
    }
  }

  async function searchMore(loader: string, gameVersion: string) {
    if (searchLoading.value || !hasMore.value) return;
    searchLoading.value = true;
    const nextOffset = searchOffset.value + searchLimit;
    try {
      const facets = [`["categories:${loader.toLowerCase()}"]`, `["versions:${gameVersion}"]`];
      const res = await $fetch<any>(
        `https://api.modrinth.com/v2/search`,
        {
          query: {
            query: searchQuery.value,
            facets: `[${facets.join(",")}]`,
            offset: nextOffset,
            limit: searchLimit,
          },
        },
      );
      const hits = res?.hits || [];
      searchResults.value.push(...hits.map((h: any) => ({
        project_id: h.project_id,
        slug: h.slug,
        title: h.title,
        description: h.description,
        icon_url: h.icon_url,
        author: h.author,
        downloads: h.downloads,
      })));
      searchOffset.value = nextOffset;
      hasMore.value = hits.length === searchLimit;
    } catch (err: any) {
      show("error", "Failed to load more", {
        description: err?.message || "Unknown error",
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
      const res = await $fetch<any[]>(`https://api.modrinth.com/v2/project/${projectId}/version`, {
        query: {
          loaders: `["${loader.toLowerCase()}"]`,
          game_versions: `["${gameVersion}"]`,
        },
      });
      return (res || []).map((v: any) => ({
        id: v.id,
        project_id: v.project_id,
        version_number: v.version_number,
        game_versions: v.game_versions,
        loaders: v.loaders,
        files: v.files.map((f: any) => ({
          url: f.url,
          filename: f.filename,
          primary: f.primary,
        })),
      }));
    } catch (err: any) {
      show("error", "Failed to load versions", {
        description: err?.message || "Unknown error",
      });
      return [];
    }
  }

  async function install(
    _projectId: string,
    versionId: string,
  ): Promise<{ url: string; filename: string } | null> {
    const key = `${_projectId}:${versionId}`;
    installLoading.value[key] = true;
    try {
      const version = await $fetch<any>(`https://api.modrinth.com/v2/version/${versionId}`);
      const primaryFile = version?.files?.find((f: any) => f.primary) || version?.files?.[0];
      if (!primaryFile) {
        throw new Error("No file found for version");
      }
      return { url: primaryFile.url, filename: primaryFile.filename };
    } catch (err: any) {
      show("error", "Install failed", {
        description: err?.message || "Unknown error",
      });
      return null;
    } finally {
      installLoading.value[key] = false;
    }
  }

  async function bulkInstall(
    items: Array<{ projectId: string; versionId: string }>,
  ): Promise<Array<{ url: string; filename: string }>> {
    const results: Array<{ url: string; filename: string }> = [];
    for (const item of items) {
      const res = await install(item.projectId, item.versionId);
      if (res) results.push(res);
    }
    return results;
  }

  return {
    searchQuery,
    searchResults,
    searchLoading,
    hasMore,
    installLoading,
    search,
    searchMore,
    getVersions,
    install,
    bulkInstall,
  };
}
