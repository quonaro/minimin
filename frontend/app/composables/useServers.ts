export interface Server {
  serverId: string;
  containerStatus: string;
  containerStartedAt?: string;
  serverStatus: string;
  serverStartedAt?: string;
  desiredStatus?: string;
  name?: string;
  gamePort: number;
  publicRcon: boolean;
  rconPort?: number;
  restartPolicy?: string;
  engineType: string;
  gameVersion: string;
  loaderVersion?: string;
  ramBytes?: number;
  volumePath?: string;
  hostPath?: string;
  containerPath?: string;
  modCount?: number;
  externalJavaArgs?: string[];
  imageName?: string;
}

export function useServers() {
  const { servers, initialized, serverMap } = useServerEvents();
  const pending = computed(() => !initialized.value);
  const error = ref<Error | null>(null);

  async function refresh() {
    try {
      const res = await $fetch<Server[] | { body: Server[] }>("/servers", {
        baseURL: useApiBase(),
        credentials: "include",
      });
      const list = Array.isArray(res)
        ? res
        : ((res as any).body || []);
      const next: Record<string, Server> = {};
      for (const s of list) {
        if (s.serverId) next[s.serverId] = s;
      }
      serverMap.value = next;
    } catch (e: any) {
      error.value = e;
    }
  }

  return {
    servers,
    pending,
    error,
    refresh,
  };
}
