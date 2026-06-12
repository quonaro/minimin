<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 min-h-0 flex flex-col">
      <div class="mb-4 flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2 self-stretch">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="w-4 h-4 text-blue-500"
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
              >{{ logLines.length }} lines</span
            >
          </div>

          <div class="flex items-center gap-2 self-stretch">
            <button
              class="px-3 py-2 rounded-lg bg-gray-100 dark:bg-neutral-800 text-sm text-gray-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-700"
              @click="clearLogs"
            >
              Clear
            </button>
            <button
              class="px-3 py-2 rounded-lg bg-gray-100 dark:bg-neutral-800 text-sm text-gray-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-700"
              @click="copyLogs"
            >
              Copy
            </button>
          </div>

          <div class="flex items-center gap-2 self-stretch">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="w-4 h-4 text-gray-500"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z"
              />
            </svg>
            <input
              :value="searchQueryRaw"
              type="text"
              placeholder="Search logs..."
              @input="onSearchInput"
              class="text-sm bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 rounded-lg px-2 py-1 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary w-40"
            />
            <span
              v-if="searchQuery"
              class="text-xs text-gray-500 dark:text-neutral-400"
            >
              {{ filteredLines.length }} matches
            </span>
          </div>

          <div class="flex items-center gap-2 self-stretch">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="w-4 h-4 text-purple-500"
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
              @change="reconnect()"
            >
              <option :value="100">100</option>
              <option :value="500">500</option>
              <option :value="1000">1 000</option>
              <option :value="5000">5 000</option>
              <option :value="10000">10 000</option>
              <option :value="25000">25 000</option>
              <option :value="50000">50 000</option>
            </select>
          </div>

          <div class="flex items-center gap-2 self-stretch">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="w-4 h-4 text-orange-500"
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

          <div class="flex items-center gap-2 self-stretch">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="w-4 h-4 text-pink-500"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path
                d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z"
              />
              <path d="M2 12h20" />
              <path
                d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"
              />
            </svg>
            <button
              class="px-3 py-2 rounded-lg text-sm border border-gray-300 dark:border-neutral-700 bg-white dark:bg-neutral-800 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors"
              @click="colored = !colored"
            >
              {{ colored ? "Colored" : "Monochrome" }}
            </button>
          </div>
        </div>

        <div class="flex items-center gap-2 self-stretch">
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
            class="text-xs"
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
            class="px-3 py-2 rounded-lg bg-primary text-white text-sm hover:bg-primary/90"
            @click="reconnect()"
          >
            Reconnect
          </button>
        </div>
      </div>

      <div
        ref="logContainer"
        :class="[
          'flex-1 min-h-0 min-w-0 bg-white dark:bg-neutral-900 text-gray-900 dark:text-neutral-100 font-mono rounded-xl p-4 overflow-y-auto overflow-x-hidden shadow-inner',
          fontSize,
        ]"
        @scroll="onScroll"
      >
        <div
          v-if="logLines.length === 0"
          class="text-gray-500 dark:text-neutral-400 italic"
        >
          Waiting for logs...
        </div>
        <template v-else>
          <div
            v-for="(line, i) in filteredLines"
            :key="line.id"
            class="flex gap-3 py-0.5 border-b border-gray-200 dark:border-neutral-800/40 hover:bg-gray-100 dark:hover:bg-neutral-800/60 transition-colors min-w-0"
          >
            <span
              class="text-gray-400 dark:text-neutral-600 select-none text-right tabular-nums min-w-[3ch] shrink-0"
            >
              {{ i + 1 }}
            </span>
            <span
              :class="['whitespace-pre-wrap break-all', line.levelClass]"
              v-html="line.html"
            />
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  onMounted,
  onUnmounted,
  nextTick,
  watch,
  computed,
  type Ref,
} from "vue";
import { parseLogLevel, getLogLevelClass } from "~/utils/logLevel";

const props = defineProps<{
  serverId: string;
}>();

const colored = ref(true);

const route = useRoute();
const serverId = props.serverId || (route.params.serverId as string);

const tail = ref(500);
const searchQueryRaw = ref("");
const searchQuery = ref("");
const buffer = ref("");
const wsStatus = ref("Connecting...");
const logContainer = ref<HTMLElement | null>(null);
const userScrolledUp = ref(false);
const fontSize = ref<string>("text-sm");
const skipIncoming = ref(false);
const { show: showToast } = useToast();

interface LogLine {
  id: number;
  text: string;
  level: string;
  levelClass: string;
  html: string;
}

let lineIdCounter = 0;
const logLines = ref<LogLine[]>([]);

const filteredLines = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  if (!q) return logLines.value;
  return logLines.value.filter((line) => line.text.toLowerCase().includes(q));
});

const estimatedLineHeight = computed(() => {
  switch (fontSize.value) {
    case "text-xs":
      return 24;
    case "text-base":
      return 32;
    case "text-lg":
      return 36;
    default:
      return 28;
  }
});

let ws: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let unmounted = false;
let reconnectAttempts = 0;
let socketCounter = 0;
let nextTail: number | undefined = undefined;

const FLUSH_INTERVAL_MS = 50;
const RECONNECT_BASE_MS = 5000;
const RECONNECT_MAX_MS = 60000;

let pendingRawLines: string[] = [];
let flushTimer: ReturnType<typeof setTimeout> | null = null;
let debounceTimer: ReturnType<typeof setTimeout> | null = null;

function escapeHtml(str: string): string {
  return str.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

function buildLogLine(text: string): LogLine {
  const level = parseLogLevel(text);
  const levelClass = getLogLevelClass(level, colored.value);
  return {
    id: ++lineIdCounter,
    text,
    level: level ?? "",
    levelClass,
    html: escapeHtml(text),
  };
}

function recomputeHighlight(lines: LogLine[], query: string) {
  const q = query.trim();
  if (!q) {
    for (const l of lines) {
      l.html = escapeHtml(l.text);
    }
    return;
  }
  const escapedQuery = escapeHtml(q);
  const safeQuery = escapedQuery.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const re = new RegExp(`(${safeQuery})`, "gi");
  const markOpen =
    '<mark class="bg-yellow-300 dark:bg-yellow-700 text-gray-900 dark:text-white rounded px-0.5">';
  for (const l of lines) {
    l.html = escapeHtml(l.text).replace(re, `${markOpen}$1</mark>`);
  }
}

watch(colored, () => {
  for (const l of logLines.value) {
    l.levelClass = getLogLevelClass(
      l.level as import("~/utils/logLevel").LogLevel,
      colored.value,
    );
  }
});

watch(searchQuery, (q) => {
  recomputeHighlight(filteredLines.value, q);
});

function onSearchInput(e: Event) {
  const val = (e.target as HTMLInputElement).value;
  searchQueryRaw.value = val;
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    searchQuery.value = val;
  }, 150);
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
  const apiBase = (config.public.apiBase as string) || "http://localhost:8081";
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
  const tailVal = nextTail !== undefined ? nextTail : tail.value;
  nextTail = undefined;
  const url = `${wsBase}/ws/servers/${serverId}/logs?tail=${tailVal}&token=${encodeURIComponent(token)}`;

  const socketId = ++socketCounter;
  const socket = new WebSocket(url);
  ws = socket;

  socket.onopen = () => {
    if (ws !== socket) return;
    reconnectAttempts = 0;
    wsStatus.value = "Connected";
    buffer.value = "";
    skipIncoming.value = false;
  };

  socket.onmessage = (e) => {
    if (ws !== socket) return;
    appendChunk(String(e.data));
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
      const delay = Math.min(
        RECONNECT_BASE_MS * 2 ** reconnectAttempts,
        RECONNECT_MAX_MS,
      );
      reconnectAttempts++;
      reconnectTimer = setTimeout(connect, delay);
    }
  };
}

function scheduleFlush() {
  if (flushTimer) return;
  flushTimer = setTimeout(() => {
    flushTimer = null;
    flushPending();
  }, FLUSH_INTERVAL_MS);
}

function flushPending() {
  if (skipIncoming.value) {
    pendingRawLines = [];
    return;
  }
  if (pendingRawLines.length === 0) return;
  const built = pendingRawLines.map((t) => buildLogLine(t));
  pendingRawLines = [];
  logLines.value.push(...built);
  if (logLines.value.length > tail.value) {
    logLines.value = logLines.value.slice(logLines.value.length - tail.value);
  }
  recomputeHighlight(logLines.value, searchQuery.value);
  nextTick(() => {
    if (!userScrolledUp.value) {
      scrollToBottom();
    }
  });
}

function appendChunk(chunk: string) {
  if (skipIncoming.value) {
    buffer.value = "";
    return;
  }
  buffer.value += chunk;
  const parts = buffer.value.split("\n");
  buffer.value = parts.pop() || "";
  if (parts.length > 0) {
    pendingRawLines.push(...parts);
    if (pendingRawLines.length > tail.value) {
      pendingRawLines = pendingRawLines.slice(
        pendingRawLines.length - tail.value,
      );
    }
    scheduleFlush();
  }
}

function reconnect(withTail?: number) {
  logLines.value = [];
  lineIdCounter = 0;
  buffer.value = "";
  pendingRawLines = [];
  skipIncoming.value = false;
  if (flushTimer) {
    clearTimeout(flushTimer);
    flushTimer = null;
  }
  if (debounceTimer) {
    clearTimeout(debounceTimer);
    debounceTimer = null;
  }
  reconnectAttempts = 0;
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  nextTail = withTail;
  connect();
}

function clearLogs() {
  logLines.value = [];
  lineIdCounter = 0;
  buffer.value = "";
  pendingRawLines = [];
  searchQueryRaw.value = "";
  searchQuery.value = "";
  skipIncoming.value = true;
  if (flushTimer) {
    clearTimeout(flushTimer);
    flushTimer = null;
  }
  if (debounceTimer) {
    clearTimeout(debounceTimer);
    debounceTimer = null;
  }
  showToast("success", "Logs cleared");
}

async function copyLogs() {
  try {
    await navigator.clipboard.writeText(
      logLines.value.map((l) => l.text).join("\n"),
    );
    showToast("success", "Copied to clipboard", {
      description: `${logLines.value.length} lines`,
    });
  } catch {
    showToast("error", "Failed to copy");
  }
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
  const savedFont = localStorage.getItem("logs-font-size");
  if (savedFont) {
    fontSize.value = savedFont;
  }
  const savedColor = localStorage.getItem("logs-colored");
  if (savedColor !== null) {
    colored.value = savedColor === "true";
  }
  const savedTail = localStorage.getItem("logs-tail");
  if (savedTail) {
    const parsed = parseInt(savedTail, 10);
    if (!isNaN(parsed)) {
      tail.value = parsed;
    }
  }
  connect();
});

watch(fontSize, (value) => {
  localStorage.setItem("logs-font-size", value);
});

watch(colored, (value) => {
  localStorage.setItem("logs-colored", String(value));
});

watch(tail, (value) => {
  localStorage.setItem("logs-tail", String(value));
});

watch(filteredLines, () => {});

onUnmounted(() => {
  unmounted = true;
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (debounceTimer) {
    clearTimeout(debounceTimer);
    debounceTimer = null;
  }
  if (ws) {
    const oldWs = ws;
    ws = null;
    oldWs.onclose = null;
    oldWs.close();
  }
});

defineExpose({ scrollToBottom, reconnect });
</script>
