export interface ModrinthProject {
  project_id: string;
  slug: string;
  title: string;
  description: string;
  icon_url: string;
  author: string;
  downloads: number;
  server_side?: string;
  client_side?: string;
}

export interface ModrinthDependency {
  version_id: string | null;
  project_id: string;
  dependency_type: "required" | "optional" | "incompatible" | "embedded";
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
  dependencies?: ModrinthDependency[];
}

export interface BulkJob {
  id: string;
  status: "pending" | "running" | "done" | "failed";
  completed: number;
  total: number;
  errors: string[];
  updatedAt: string;
}

interface SearchCacheEntry {
  results: ModrinthProject[];
  offset: number;
  hasMore: boolean;
  timestamp: number;
}

interface VersionsCacheEntry {
  versions: ModrinthVersion[];
  timestamp: number;
}

const CACHE_TTL_MS = 5 * 60 * 1000; // 5 minutes

const searchCache = new Map<string, SearchCacheEntry>();
const versionsCache = new Map<string, VersionsCacheEntry>();

function isCacheValid(timestamp: number): boolean {
  return Date.now() - timestamp < CACHE_TTL_MS;
}

function buildSearchKey(query: string, loader: string, gameVersion: string, offset: number): string {
  return `${query.toLowerCase().trim()}:${loader.toLowerCase()}:${gameVersion}:${offset}`;
}

function buildVersionsKey(projectId: string, loader: string, gameVersion: string): string {
  return `${projectId}:${loader.toLowerCase()}:${gameVersion}`;
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

    const key = buildSearchKey(searchQuery.value, loader, gameVersion, 0);
    const cached = searchCache.get(key);
    if (cached && isCacheValid(cached.timestamp)) {
      searchResults.value = cached.results;
      searchOffset.value = cached.offset;
      hasMore.value = cached.hasMore;
      searchLoading.value = false;
      return;
    }

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
      const results = hits.map((h: any) => ({
        project_id: h.project_id,
        slug: h.slug,
        title: h.title,
        description: h.description,
        icon_url: h.icon_url,
        author: h.author,
        downloads: h.downloads,
        server_side: h.server_side,
        client_side: h.client_side,
      }));
      searchResults.value = results;
      hasMore.value = hits.length === searchLimit;
      searchCache.set(key, { results, offset: 0, hasMore: hasMore.value, timestamp: Date.now() });
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

    const key = buildSearchKey(searchQuery.value, loader, gameVersion, nextOffset);
    const cached = searchCache.get(key);
    if (cached && isCacheValid(cached.timestamp)) {
      searchResults.value.push(...cached.results);
      searchOffset.value = cached.offset;
      hasMore.value = cached.hasMore;
      searchLoading.value = false;
      return;
    }

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
      const results = hits.map((h: any) => ({
        project_id: h.project_id,
        slug: h.slug,
        title: h.title,
        description: h.description,
        icon_url: h.icon_url,
        author: h.author,
        downloads: h.downloads,
        server_side: h.server_side,
        client_side: h.client_side,
      }));
      searchResults.value.push(...results);
      searchOffset.value = nextOffset;
      hasMore.value = hits.length === searchLimit;
      searchCache.set(key, { results, offset: nextOffset, hasMore: hasMore.value, timestamp: Date.now() });
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
    const key = buildVersionsKey(projectId, loader, gameVersion);
    const cached = versionsCache.get(key);
    if (cached && isCacheValid(cached.timestamp)) {
      return cached.versions;
    }

    try {
      const res = await $fetch<any[]>(`https://api.modrinth.com/v2/project/${projectId}/version`, {
        query: {
          loaders: `["${loader.toLowerCase()}"]`,
          game_versions: `["${gameVersion}"]`,
        },
      });
      const versions = (res || []).map((v: any) => ({
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
      versionsCache.set(key, { versions, timestamp: Date.now() });
      return versions;
    } catch (err: any) {
      show("error", "Failed to load versions", {
        description: err?.message || "Unknown error",
      });
      return [];
    }
  }

  async function getVersionDetails(versionId: string): Promise<ModrinthVersion | null> {
    try {
      const v = await $fetch<any>(`https://api.modrinth.com/v2/version/${versionId}`);
      return {
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
        dependencies: v.dependencies || [],
      };
    } catch (err: any) {
      show("error", "Failed to load version details", {
        description: err?.message || "Unknown error",
      });
      return null;
    }
  }

  async function install(
    _projectId: string,
    versionId: string,
  ): Promise<{ url: string; filename: string; dependencies: ModrinthDependency[] } | null> {
    const key = `${_projectId}:${versionId}`;
    installLoading.value[key] = true;
    try {
      const version = await getVersionDetails(versionId);
      if (!version) {
        throw new Error("Version not found");
      }
      const primaryFile = version.files.find((f) => f.primary) || version.files[0];
      if (!primaryFile) {
        throw new Error("No file found for version");
      }
      return { url: primaryFile.url, filename: primaryFile.filename, dependencies: version.dependencies || [] };
    } catch (err: any) {
      show("error", "Install failed", {
        description: err?.message || "Unknown error",
      });
      return null;
    } finally {
      installLoading.value[key] = false;
    }
  }

  async function resolveDependency(
    projectId: string,
    versionId: string | null,
    loader: string,
    gameVersion: string,
  ): Promise<{ url: string; filename: string } | null> {
    if (versionId) {
      return install(projectId, versionId).then((r) => (r ? { url: r.url, filename: r.filename } : null));
    }
    const versions = await getVersions(projectId, loader, gameVersion);
    if (versions.length === 0) return null;
    const first = versions[0]!;
    return install(projectId, first.id).then((r) => (r ? { url: r.url, filename: r.filename } : null));
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
    getVersionDetails,
    install,
    resolveDependency,
    bulkInstall,
  };
}
