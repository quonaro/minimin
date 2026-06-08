import type { Ref } from "vue";

export interface SseEvent {
  type: string;
  agentId?: string;
  serverId?: string;
  message: string;
  timestamp: string;
  oldStatus?: string;
  newStatus?: string;
  desiredStatus?: string;
}

// Module-level singleton state
const lastEvent: Ref<SseEvent | null> = ref(null);
const connected = ref(false);

let eventSource: EventSource | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let retryDelay = 1000;
let initialized = false;

function connect() {
  if (eventSource) {
    eventSource.close();
  }

  const base = useApiBase();
  const url = `${base}/events`;
  eventSource = new EventSource(url, { withCredentials: true });

  eventSource.onopen = () => {
    connected.value = true;
    retryDelay = 1000;
  };

  eventSource.onmessage = (e) => {
    try {
      const data = JSON.parse(e.data) as SseEvent;
      lastEvent.value = data;
    } catch (err) {
      console.error("Failed to parse SSE message:", err);
    }
  };

  eventSource.onerror = () => {
    connected.value = false;
    eventSource?.close();
    eventSource = null;
    retryDelay = Math.min(retryDelay * 2, 30000);
    reconnectTimer = setTimeout(connect, retryDelay);
  };
}

function disconnect() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  eventSource?.close();
  eventSource = null;
  connected.value = false;
}

export function useEventSource() {
  if (!initialized) {
    initialized = true;
    if (typeof window !== "undefined") {
      connect();
    }
  }

  return {
    lastEvent: readonly(lastEvent),
    connected: readonly(connected),
    connect,
    disconnect,
  };
}
