interface Agent {
  id: string;
  name: string;
  host: string;
  apiKey: string;
  createdAt: string;
}

export function useCurrentAgent() {
  const route = useRoute();

  const agentId = computed(() => route.params.id as string);

  const { data: agentsData, pending, error } = useFetch('/agents', {
    baseURL: useApiBase(),
    credentials: 'include',
    key: 'agents',
  });

  const agent = computed<Agent | null>(() => {
    if (!agentsData.value || !agentId.value) return null;
    let list: any[] = [];
    const val = agentsData.value;
    if (Array.isArray(val)) {
      list = val;
    } else if (typeof val === 'object' && 'body' in val) {
      list = (val as any).body || [];
    }
    return list.find((a: any) => a.id === agentId.value) || null;
  });

  return {
    agentId,
    agent,
    pending,
    error,
  };
}
