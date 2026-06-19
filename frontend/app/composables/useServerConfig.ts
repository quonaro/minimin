import { ref, computed, watch, watchEffect, toValue, type MaybeRefOrGetter } from "vue";
import { useServerEvents } from "./useServerEvents";

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

export function useServerConfig(serverId: MaybeRefOrGetter<string | undefined | null>) {
  const id = computed(() => toValue(serverId) ?? "");
  const { serverMap } = useServerEvents();
  const server = computed(() => (id.value ? serverMap.value[id.value] : undefined));

  watchEffect(() => {
    if (id.value && !configMap.value[id.value]?.fetched) {
      fetchConfig(id.value);
    }
  });

  watch(
    () => [server.value?.serverStatus, server.value?.containerStatus],
    () => {
      if (!id.value || !server.value) return;
      if (configMap.value[id.value]?.config?.initialized) return;
      if (configMap.value[id.value]?.loading) return;
      refresh();
    },
  );

  const initialized = computed<boolean>(() => {
    if (!id.value) return true;
    return configMap.value[id.value]?.config?.initialized ?? true;
  });

  const config = computed<ServerConfig | null>(() => {
    if (!id.value) return null;
    return configMap.value[id.value]?.config ?? null;
  });

  const loading = computed<boolean>(() => {
    if (!id.value) return false;
    return configMap.value[id.value]?.loading ?? false;
  });

  const error = computed<string | null>(() => {
    if (!id.value) return null;
    return configMap.value[id.value]?.error ?? null;
  });

  async function refresh() {
    if (!id.value) return;
    await fetchConfig(id.value);
  }

  return {
    initialized,
    config,
    loading,
    error,
    refresh,
  };
}
