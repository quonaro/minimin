import { ref } from "vue";

export interface GameVersion {
  id: string;
  releaseDate?: string;
  stable?: boolean;
}

export interface LoaderVersion {
  id: string;
  stable?: boolean;
}

interface AllVersionsResponse {
  vanilla: GameVersion[];
  paper: GameVersion[];
  fabricGames: GameVersion[];
  fabricLoaders: LoaderVersion[];
  forge: { id: string; stable?: boolean; loaders?: string[] }[];
}

let cachedAll: AllVersionsResponse | null = null;
let cachedPromise: Promise<AllVersionsResponse> | null = null;

export function useVersions() {
  const loading = ref(false);
  const error = ref<string | null>(null);

  const vanillaVersions = ref<GameVersion[]>([]);
  const paperVersions = ref<GameVersion[]>([]);
  const fabricGameVersions = ref<GameVersion[]>([]);
  const fabricLoaderVersions = ref<LoaderVersion[]>([]);
  const forgeGameVersions = ref<GameVersion[]>([]);
  const forgeLoaderMap = ref<Record<string, string[]>>({});

  const baseURL = useApiBase();

  async function loadAll() {
    if (cachedAll) {
      populateFromCache(cachedAll);
      return;
    }
    if (cachedPromise) {
      const data = await cachedPromise;
      populateFromCache(data);
      return;
    }

    loading.value = true;
    error.value = null;
    cachedPromise = $fetch<AllVersionsResponse>("/versions/all", {
      baseURL,
      credentials: "include",
    });

    try {
      const data = await cachedPromise;
      cachedAll = data;
      populateFromCache(data);
    } catch (err: any) {
      error.value = `Failed to load versions: ${err?.message || err}`;
    } finally {
      loading.value = false;
      cachedPromise = null;
    }
  }

  function populateFromCache(data: AllVersionsResponse) {
    vanillaVersions.value = data.vanilla ?? [];
    paperVersions.value = data.paper ?? [];
    fabricGameVersions.value = data.fabricGames ?? [];
    fabricLoaderVersions.value = data.fabricLoaders ?? [];
    forgeGameVersions.value =
      data.forge?.map((v) => ({ id: v.id, stable: v.stable })) ?? [];
    forgeLoaderMap.value = Object.fromEntries(
      data.forge?.map((v) => [v.id, v.loaders || []]) ?? []
    );
  }

  async function loadForEngine(_engine: string) {
    await loadAll();
  }

  function gameVersionsFor(engine: string): GameVersion[] {
    switch (engine) {
      case "VANILLA":
        return vanillaVersions.value;
      case "PAPER":
        return paperVersions.value;
      case "FABRIC":
        return fabricGameVersions.value;
      case "FORGE":
        return forgeGameVersions.value;
      default:
        return [];
    }
  }

  function loaderVersionsFor(engine: string, _mcVersion?: string): LoaderVersion[] {
    switch (engine) {
      case "FABRIC":
        return fabricLoaderVersions.value;
      case "FORGE": {
        if (!_mcVersion || _mcVersion === "LATEST") return [];
        const loaders = forgeLoaderMap.value[_mcVersion] || [];
        return loaders.map((v) => ({ id: v, stable: true }));
      }
      default:
        return [];
    }
  }

  return {
    loading,
    error,
    loadForEngine,
    gameVersionsFor,
    loaderVersionsFor,
  };
}
