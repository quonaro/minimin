import { ref, onUnmounted } from "vue";

export interface RconMessage {
  type: "command" | "response" | "error" | "system";
  text: string;
  timestamp: Date;
}

const RECONNECT_BASE_MS = 5000;
const RECONNECT_MAX_MS = 60000;
const MAX_RECONNECT_ATTEMPTS = 5;

export function useRconConsole(serverId: string) {
  const wsStatus = ref("Connecting...");
  const messages = ref<RconMessage[]>([]);
  let ws: WebSocket | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let unmounted = false;
  let reconnectAttempts = 0;
  let receivedAgentError = false;

  function addMessage(type: RconMessage["type"], text: string) {
    messages.value.push({ type, text, timestamp: new Date() });
  }

  function connect() {
    if (ws) {
      const oldWs = ws;
      ws = null;
      oldWs.onopen = null;
      oldWs.onmessage = null;
      oldWs.onerror = null;
      oldWs.onclose = null;
      oldWs.close();
    }
    const config = useRuntimeConfig();
    const wsBaseRaw = (config.public.wsBase as string) || "";
    const apiBase =
      (config.public.apiBase as string) || "http://localhost:8081";
    let wsBase: string;
    if (wsBaseRaw) {
      wsBase = wsBaseRaw;
    } else if (apiBase.startsWith("/")) {
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      wsBase = `${proto}//${window.location.host}${apiBase === "/" ? "" : apiBase}`;
    } else {
      wsBase = apiBase.replace(/^http/, "ws");
    }
    const token = useCookie("auth_token").value || "";
    const url = `${wsBase}/ws/servers/${serverId}/rcon?token=${encodeURIComponent(token)}`;

    const socket = new WebSocket(url);
    ws = socket;

    socket.onopen = () => {
      if (ws !== socket) return;
      reconnectAttempts = 0;
      receivedAgentError = false;
      wsStatus.value = "Connected";
      addMessage("system", "Connected to RCON console.");
    };

    socket.onmessage = (e) => {
      if (ws !== socket) return;
      try {
        const data = JSON.parse(String(e.data));
        if (data.response !== undefined) {
          addMessage("response", data.response);
        } else if (data.error !== undefined) {
          receivedAgentError = true;
          addMessage("error", data.error);
        } else {
          addMessage("response", String(e.data));
        }
      } catch {
        addMessage("response", String(e.data));
      }
    };

    socket.onerror = () => {
      if (ws !== socket) return;
      wsStatus.value = "Error";
    };

    socket.onclose = () => {
      if (ws !== socket) return;
      ws = null;
      if (!unmounted) {
        wsStatus.value = "Disconnected";
        if (!receivedAgentError) {
          addMessage("system", "Disconnected from RCON console.");
        }
        if (receivedAgentError || reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
          return;
        }
        const delay = Math.min(
          RECONNECT_BASE_MS * 2 ** reconnectAttempts,
          RECONNECT_MAX_MS,
        );
        reconnectAttempts++;
        reconnectTimer = setTimeout(connect, delay);
      }
    };
  }

  function sendCommand(cmd: string) {
    let clean = cmd.trim();
    if (!clean || !ws || ws.readyState !== WebSocket.OPEN) return;
    if (clean.startsWith("/")) {
      clean = clean.slice(1);
    }
    ws.send(JSON.stringify({ command: clean }));
    addMessage("command", cmd.trim());
  }

  function reconnect() {
    messages.value = [];
    reconnectAttempts = 0;
    receivedAgentError = false;
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    connect();
  }

  function clearMessages() {
    messages.value = [];
  }

  onUnmounted(() => {
    unmounted = true;
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (ws) {
      const oldWs = ws;
      ws = null;
      oldWs.onclose = null;
      oldWs.close();
    }
  });

  return {
    wsStatus,
    messages,
    connect,
    sendCommand,
    reconnect,
    clearMessages,
  };
}
