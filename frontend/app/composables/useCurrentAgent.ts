interface Agent {
  id: string;
  name: string;
  host: string;
  api_key: string;
  created_at: string;
}

export function useCurrentAgent() {
  const route = useRoute();

  const agentId = computed(() => route.params.id as string);

  const url = computed(() => (agentId.value ? `/agents/${agentId.value}` : ''));

  const { data, pending, error } = useApiFetch<{ body: Agent }>(url, {
    server: false,
    lazy: false,
  });

  const agent = computed<Agent | null>(() => {
    if (data.value && typeof data.value === "object" && "body" in data.value) {
      return (data.value as any).body as Agent;
    }
    return null;
  });

  return {
    agentId,
    agent,
    pending,
    error,
  };
}
