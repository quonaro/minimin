<template>
  <div class="p-6 h-[calc(100vh-4rem)] flex flex-col">
    <div class="mb-4 flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Logs</h1>
      </div>
      <div class="flex items-center gap-4">
        <div class="flex items-center gap-2">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-4 h-4 text-gray-500 dark:text-neutral-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M4 6h16M4 12h16M4 18h16" />
          </svg>
          <span class="text-xs text-gray-500 dark:text-neutral-400"
            >{{ lines.length }} lines</span
          >
        </div>

        <div class="flex items-center gap-2">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-4 h-4"
            :class="{
              'text-green-500': wsStatus === 'Connected',
              'text-red-500':
                wsStatus === 'Error' || wsStatus === 'Disconnected',
              'text-gray-500 dark:text-neutral-400': !wsStatus,
            }"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path
              d="M9.348 14.652a3.75 3.75 0 0 1 5.304 0m-9.9-3.9a7.5 7.5 0 0 1 14.556 0M1.5 6.75a12 12 0 0 1 21 0"
            />
          </svg>
          <span
            class="text-xs leading-none"
            :class="{
              'text-green-500': wsStatus === 'Connected',
              'text-red-500':
                wsStatus === 'Error' || wsStatus === 'Disconnected',
              'text-gray-500 dark:text-neutral-400': !wsStatus,
            }"
          >
            {{ wsStatus || "Connecting..." }}
          </span>
          <button
            class="text-xs leading-none rounded-full bg-primary text-white hover:bg-primary/90 focus:outline-none px-2 py-0.5"
            @click="reconnect"
          >
            Reconnect
          </button>
        </div>

        <div class="flex items-center gap-2">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-4 h-4 text-gray-500 dark:text-neutral-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path
              d="M3 4.5h16.5M3 9h16.5m-16.5 4.5h10.5m0 0l3-3m-3 3l3 3M3 18h16.5"
            />
          </svg>
          <select
            id="tail"
            v-model="tail"
            class="text-sm bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 rounded-lg px-2 py-1 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <option :value="100">100</option>
            <option :value="500">500</option>
            <option :value="1000">1 000</option>
            <option :value="5000">5 000</option>
          </select>
        </div>

        <div class="flex items-center gap-2">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-4 h-4 text-gray-500 dark:text-neutral-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path
              d="M4.5 3.75h2.25v6.75h-2.25v-6.75zm0 8.25h2.25v6.75h-2.25v-6.75zm3.75-8.25h2.25v4.5h-2.25v-4.5zm0 6h2.25v9h-2.25v-9zm3.75-6h9v2.25h-3.375v13.5h-2.25v-13.5h-3.375v-2.25z"
            />
          </svg>
          <select
            id="font-size"
            v-model="fontSize"
            class="text-sm bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 rounded-lg px-2 py-1 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <option value="text-xs">Small</option>
            <option value="text-sm">Medium</option>
            <option value="text-base">Large</option>
            <option value="text-lg">XL</option>
          </select>
        </div>
      </div>
    </div>

    <div
      ref="logContainer"
      :class="[
        'flex-1 min-h-0 bg-white dark:bg-neutral-900 text-gray-900 dark:text-neutral-100 font-mono rounded-xl p-4 overflow-y-auto shadow-inner',
        fontSize,
      ]"
      @scroll="onScroll"
    >
      <div v-if="lines.length === 0 && !wsStatus" class="text-gray-500 italic">
        Waiting for logs...
      </div>
      <template v-else>
        <div
          v-for="(line, i) in lines"
          :key="i"
          class="flex gap-3 py-0.5 border-b border-gray-200 dark:border-neutral-800/40 hover:bg-gray-100 dark:hover:bg-neutral-800/60 transition-colors"
        >
          <span
            class="text-gray-400 dark:text-neutral-600 select-none text-right tabular-nums min-w-[3ch] shrink-0"
          >
            {{ i + 1 }}
          </span>
          <span
            class="whitespace-pre-wrap break-words text-gray-700 dark:text-neutral-300"
            >{{ line }}</span
          >
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from "vue";

definePageMeta({
  middleware: "auth",
});

const route = useRoute();
const { agentId } = useCurrentAgent();
const serverId = route.params.serverId as string;

const tail = ref(500);
const lines = ref<string[]>([]);
const buffer = ref("");
const wsStatus = ref("Connecting...");
const logContainer = ref<HTMLElement | null>(null);
const userScrolledUp = ref(false);
const fontSize = ref<string>("text-sm");

let ws: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let unmounted = false;
let reconnectAttempts = 0;
let socketCounter = 0;

const MAX_LINES = 5000;
const RECONNECT_BASE_MS = 1000;
const RECONNECT_MAX_MS = 30000;

function connect() {
  console.log("[WS] connect() called, current ws =", ws?.readyState ?? null);
  if (ws) {
    const oldWs = ws;
    ws = null;
    oldWs.onopen = null;
    oldWs.onmessage = null;
    oldWs.onerror = null;
    oldWs.onclose = null;
    console.log("[WS] closing old socket readyState =", oldWs.readyState);
    oldWs.close();
  }
  if (!agentId.value) {
    console.log("[WS] connect() aborted: no agentId");
    return;
  }

  const config = useRuntimeConfig();
  const base = (config.public.apiBase as string) || "http://localhost:8081";
  const wsBase = base.replace(/^http/, "ws");
  const url = `${wsBase}/ws/agent/${agentId.value}/servers/${serverId}/logs?tail=${tail.value}`;

  const socketId = ++socketCounter;
  const socket = new WebSocket(url);
  ws = socket;
  console.log(
    "[WS] new socket #",
    socketId,
    "created, readyState =",
    socket.readyState,
  );

  socket.onopen = () => {
    console.log(
      "[WS] socket #",
      socketId,
      "onopen, ws === socket?",
      ws === socket,
    );
    if (ws !== socket) return;
    reconnectAttempts = 0;
    wsStatus.value = "Connected";
    buffer.value = "";
  };

  socket.onmessage = (e) => {
    if (ws !== socket) return;
    appendChunk(String(e.data));
  };

  socket.onerror = () => {
    console.log(
      "[WS] socket #",
      socketId,
      "onerror, ws === socket?",
      ws === socket,
    );
    if (ws !== socket) return;
    wsStatus.value = "Error";
  };

  socket.onclose = () => {
    console.log(
      "[WS] socket #",
      socketId,
      "onclose, ws === socket?",
      ws === socket,
      "unmounted =",
      unmounted,
    );
    if (ws !== socket) return;
    ws = null;
    if (!unmounted) {
      wsStatus.value = "Disconnected";
      const delay = Math.min(
        RECONNECT_BASE_MS * 2 ** reconnectAttempts,
        RECONNECT_MAX_MS,
      );
      console.log(
        "[WS] socket #",
        socketId,
        "scheduling reconnect in",
        delay,
        "ms, attempt",
        reconnectAttempts,
      );
      reconnectAttempts++;
      reconnectTimer = setTimeout(connect, delay);
    }
  };
}

function appendChunk(chunk: string) {
  buffer.value += chunk;
  const parts = buffer.value.split("\n");
  buffer.value = parts.pop() || "";
  if (parts.length > 0) {
    lines.value.push(...parts);
    if (lines.value.length > MAX_LINES) {
      lines.value = lines.value.slice(lines.value.length - MAX_LINES);
    }
    nextTick(() => {
      if (!userScrolledUp.value) {
        scrollToBottom();
      }
    });
  }
}

function reconnect() {
  console.log("[WS] reconnect() called");
  lines.value = [];
  buffer.value = "";
  reconnectAttempts = 0;
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  connect();
}

function scrollToBottom() {
  const el = logContainer.value;
  if (el) {
    el.scrollTop = el.scrollHeight;
  }
}

function onScroll() {
  const el = logContainer.value;
  if (!el) return;
  const threshold = 20;
  const isAtBottom =
    el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
  userScrolledUp.value = !isAtBottom;
}

onMounted(() => {
  console.log("[WS] onMounted");
  const saved = localStorage.getItem("logs-font-size");
  if (saved) {
    fontSize.value = saved;
  }
  connect();
});

watch(fontSize, (value) => {
  localStorage.setItem("logs-font-size", value);
});

watch(tail, () => {
  reconnect();
});

watch(
  () => agentId.value,
  (newVal, oldVal) => {
    console.log("[WS] watch agentId", oldVal, "->", newVal);
    if (newVal && newVal !== oldVal) {
      reconnect();
    }
  },
);

onUnmounted(() => {
  console.log("[WS] onUnmounted");
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
</script>
