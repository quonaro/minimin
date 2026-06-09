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
}

export function useServers(agentId: Ref<string | undefined>) {
  const id = computed(() => agentId.value);

  const serversUrl = computed(() =>
    id.value ? `/agent/${id.value}/servers` : null,
  );

  const { data: serversData, pending, error, refresh } = useApiFetch<
    Server[] | { body: Server[] }
  >(serversUrl as any, {
    key: `agent-servers-${agentId.value ?? 'none'}`,
  });

  const servers = computed<Server[]>(() => {
    const val = serversData.value;
    if (Array.isArray(val)) return val as Server[];
    if (val && typeof val === 'object' && 'body' in val) {
      return (val as any).body || [];
    }
    return [];
  });

  return {
    servers,
    pending,
    error,
    refresh,
  };
}
