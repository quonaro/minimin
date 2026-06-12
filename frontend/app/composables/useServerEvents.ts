import { ref, computed } from "vue";
import type { Server } from "./useServers";

export interface PlayerEntry {
  name: string;
  uuid?: string;
  reason?: string;
  created?: string;
  level?: number;
}

interface PlayerListPayload {
  serverId: string;
  players: string[];
  max: number;
}

interface OpsPayload {
  serverId: string;
  ops: PlayerEntry[];
}

interface BansPayload {
  serverId: string;
  bans: PlayerEntry[];
}

interface WhitelistPayload {
  serverId: string;
  whitelist: PlayerEntry[];
}

export interface MetricsPayload {
  serverId: string;
  ramUsage: number;
  ramLimit: number;
  cpu: number;
  online: number;
  max: number;
  tps?: number;
  timestamp: string;
}

const serverMap = ref<Record<string, Server>>({});
const initialized = ref(false);
const playersMap = ref<Record<string, { players: string[]; max: number }>>({});
const opsMap = ref<Record<string, PlayerEntry[]>>({});
const bansMap = ref<Record<string, PlayerEntry[]>>({});
const whitelistMap = ref<Record<string, PlayerEntry[]>>({});
const metricsMap = ref<Record<string, MetricsPayload[]>>({});

const servers = computed<Server[]>(() =>
  Object.values(serverMap.value).sort(
    (a, b) =>
      (a.serverId ? a.serverId.localeCompare(b.serverId) : 0),
  ),
);

let es: EventSource | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let unmounted = false;
let initStarted = false;
let initialFetchDone = false;

async function doInitialFetch() {
  if (initialFetchDone) return;
  initialFetchDone = true;
  try {
    const res = await $fetch<Server[] | { body: Server[] }>("/servers", {
      baseURL: useApiBase(),
      credentials: "include",
    });
    const list = Array.isArray(res)
      ? res
      : ((res as any).body || []);
    for (const s of list) {
      if (s.serverId) {
        serverMap.value[s.serverId] = s;
      }
    }
  } catch {
    // silently fail; SSE will keep trying
  }
}

function initEventSource() {
  if (typeof window === "undefined") return;
  if (es || initStarted) return;
  initStarted = true;

  doInitialFetch().then(() => {
    initialized.value = true;
  });

  const base = useApiBase();
  const url = `${base}/events`;

  es = new EventSource(url, { withCredentials: true });

  es.addEventListener("server", (e) => {
    try {
      const s = JSON.parse(e.data) as Server;
      if (!s.serverId) return;
      const hasStatus = s.containerStatus || s.serverStatus;
      if (!hasStatus) {
        delete serverMap.value[s.serverId];
      } else {
        serverMap.value[s.serverId] = s;
      }
    } catch {
      // ignore malformed
    }
  });

  es.addEventListener("players", (e) => {
    try {
      const p = JSON.parse(e.data) as PlayerListPayload;
      playersMap.value[p.serverId] = {
        players: p.players || [],
        max: p.max || 0,
      };
    } catch {
      // ignore
    }
  });

  es.addEventListener("ops", (e) => {
    try {
      const p = JSON.parse(e.data) as OpsPayload;
      opsMap.value[p.serverId] = p.ops || [];
    } catch {
      // ignore
    }
  });

  es.addEventListener("bans", (e) => {
    try {
      const p = JSON.parse(e.data) as BansPayload;
      bansMap.value[p.serverId] = p.bans || [];
    } catch {
      // ignore
    }
  });

  es.addEventListener("whitelist", (e) => {
    try {
      const p = JSON.parse(e.data) as WhitelistPayload;
      whitelistMap.value[p.serverId] = p.whitelist || [];
    } catch {
      // ignore
    }
  });

  es.addEventListener("metrics", (e) => {
    try {
      const m = JSON.parse(e.data) as MetricsPayload;
      if (!m.serverId) return;
      const arr = metricsMap.value[m.serverId] || [];
      arr.push(m);
      // cap at ~1800 points (approx 1h @ 2s interval)
      if (arr.length > 1800) {
        arr.shift();
      }
      metricsMap.value[m.serverId] = arr;
    } catch {
      // ignore
    }
  });

  // Clear metrics for stopped servers
  es.addEventListener("server", (e) => {
    try {
      const s = JSON.parse(e.data) as Server;
      if (s.serverId && (s.serverStatus === "stopped" || s.serverStatus === "exited")) {
        delete metricsMap.value[s.serverId];
      }
    } catch {
      // ignore
    }
  });

  es.onerror = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (!unmounted) {
      reconnectTimer = setTimeout(() => {
        if (es?.readyState === EventSource.CLOSED) {
          es = null;
          initStarted = false;
          initEventSource();
        }
      }, 5000);
    }
  };
}

export function useServerEvents() {
  if (!initStarted && typeof window !== "undefined") {
    initEventSource();
  }
  return {
    servers,
    serverMap,
    initialized,
    playersMap,
    opsMap,
    bansMap,
    whitelistMap,
    metricsMap,
  };
}
