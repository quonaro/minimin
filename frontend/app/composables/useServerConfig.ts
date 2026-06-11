import { ref, computed } from "vue";

export interface ServerConfig {
  content: string;
  initialized: boolean;
  pendingProperties?: Record<string, string>;
}

interface ConfigState {
  config: ServerConfig | null;
  loading: boolean;
  error: string | null;
  fetched: boolean;
}

const configMap = ref<Record<string, ConfigState>>({});

async function fetchConfig(serverId: string) {
  if (configMap.value[serverId]?.loading) return;

  if (!configMap.value[serverId]) {
    configMap.value[serverId] = {
      config: null,
      loading: true,
      error: null,
      fetched: false,
    };
  } else {
    configMap.value[serverId].loading = true;
    configMap.value[serverId].error = null;
  }

  try {
    const res = await $fetch<
      { initialized?: boolean; content?: string; pendingProperties?: Record<string, string> } |
      { body?: { initialized?: boolean; content?: string; pendingProperties?: Record<string, string> } }
    >(`/servers/${serverId}/config`, {
      baseURL: useApiBase(),
      credentials: "include",
    });

    const body =
      (res as any).body ?? (res as any);

    const config: ServerConfig = {
      content: body.content ?? "",
      initialized: body.initialized !== false,
      pendingProperties: body.pendingProperties,
    };

    configMap.value[serverId] = {
      config,
      loading: false,
      error: null,
      fetched: true,
    };
  } catch (err: any) {
    configMap.value[serverId] = {
      config: null,
      loading: false,
      error: err?.message || "Failed to load config",
      fetched: true,
    };
  }
}

export function useServerConfig(serverId: string | undefined | null) {
  const id = serverId ?? "";

  if (id && !configMap.value[id]?.fetched) {
    fetchConfig(id);
  }

  const initialized = computed<boolean>(() => {
    if (!id) return true;
    return configMap.value[id]?.config?.initialized ?? true;
  });

  const config = computed<ServerConfig | null>(() => {
    if (!id) return null;
    return configMap.value[id]?.config ?? null;
  });

  const loading = computed<boolean>(() => {
    if (!id) return false;
    return configMap.value[id]?.loading ?? false;
  });

  const error = computed<string | null>(() => {
    if (!id) return null;
    return configMap.value[id]?.error ?? null;
  });

  async function refresh() {
    if (!id) return;
    await fetchConfig(id);
  }

  return {
    initialized,
    config,
    loading,
    error,
    refresh,
  };
}
