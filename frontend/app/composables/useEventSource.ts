export interface SseEvent {
  type: string;
  agentId?: string;
  serverId?: string;
  message: string;
  timestamp: string;
}

export function useEventSource() {
  return {
    lastEvent: readonly(ref<SseEvent | null>(null)),
    connected: readonly(ref(false)),
    connect: () => {},
    disconnect: () => {},
  };
}
