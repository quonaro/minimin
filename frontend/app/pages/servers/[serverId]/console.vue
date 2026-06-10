<template>
  <div class="flex flex-col min-w-0 h-[calc(100vh-4rem)]">
    <div class="flex-1 min-h-0 min-w-0 px-6 pt-6 pb-6 flex flex-col">
      <div class="mb-4 flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-4">
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            Console
          </h1>
          <button
            class="px-3 py-2 rounded-lg bg-gray-100 dark:bg-neutral-800 text-sm text-gray-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-700"
            @click="clearMessages"
          >
            Clear
          </button>
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
            @click="reconnect"
          >
            Reconnect
          </button>
        </div>
      </div>

      <div
        ref="msgContainer"
        :class="[
          'flex-1 min-h-0 min-w-0 bg-white dark:bg-neutral-900 text-gray-900 dark:text-neutral-100 font-mono rounded-xl p-4 overflow-y-auto overflow-x-hidden shadow-inner',
          fontSize,
        ]"
        @scroll="onScroll"
      >
        <div
          v-if="messages.length === 0 && !wsStatus"
          class="text-gray-500 italic"
        >
          Waiting for connection...
        </div>
        <template v-else>
          <div
            v-for="(msg, i) in messages"
            :key="i"
            class="flex gap-2 py-0.5 border-b border-gray-200 dark:border-neutral-800/40 hover:bg-gray-100 dark:hover:bg-neutral-800/60 transition-colors min-w-0"
          >
            <span
              class="text-gray-400 dark:text-neutral-600 select-none tabular-nums shrink-0"
            >
              {{ formatTime(msg.timestamp) }}
            </span>
            <span
              v-if="msg.type === 'command'"
              class="whitespace-pre-wrap break-all text-primary dark:text-primary-400"
            >
              > {{ msg.text }}
            </span>
            <span
              v-else-if="msg.type === 'error'"
              class="whitespace-pre-wrap break-all text-red-600 dark:text-red-400"
            >
              {{ msg.text }}
            </span>
            <span
              v-else
              class="whitespace-pre-wrap break-all text-gray-700 dark:text-neutral-300"
            >
              {{ msg.text }}
            </span>
          </div>
        </template>
      </div>

      <div class="mt-4 flex items-center gap-2">
        <div class="flex-1 relative">
          <input
            ref="inputRef"
            v-model="command"
            type="text"
            placeholder="Type a command and press Enter..."
            class="w-full bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 rounded-lg px-3 py-2 text-sm text-gray-900 dark:text-white font-mono focus:outline-none focus:ring-2 focus:ring-primary"
            :disabled="wsStatus !== 'Connected'"
            @input="onInput"
            @keydown.enter.prevent="onEnter"
            @keydown.up.prevent="onUp"
            @keydown.down.prevent="onDown"
            @keydown.tab="onTab"
            @keydown.esc.prevent="closeSuggestions"
          />
          <div
            v-if="showSuggestions && suggestions.length"
            class="absolute bottom-full left-0 right-0 mb-1 max-h-60 overflow-y-auto rounded-lg border border-gray-200 dark:border-neutral-700 bg-white dark:bg-neutral-800 shadow-lg z-50"
          >
            <div
              v-for="(item, i) in suggestions"
              :key="suggestionMode === 'command' ? (item as any).name : item"
              class="px-3 py-2 cursor-pointer text-sm flex items-center justify-between"
              :class="{
                'bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300':
                  selectedIndex === i,
                'text-gray-700 dark:text-neutral-300 hover:bg-gray-50 dark:hover:bg-neutral-700/50':
                  selectedIndex !== i,
              }"
              @click="acceptSuggestion(i)"
              @mouseenter="selectedIndex = i"
            >
              <template v-if="suggestionMode === 'command'">
                <span class="font-mono">{{ (item as any).name }}</span>
                <span
                  class="text-xs text-gray-400 dark:text-neutral-500 truncate ml-2 max-w-[60%]"
                  >{{ (item as any).desc }}</span
                >
              </template>
              <template v-else>
                <div class="flex items-center gap-2">
                  <img
                    :src="`https://cravatar.eu/helmavatar/${encodeURIComponent(item as string)}/24.png`"
                    alt=""
                    class="w-6 h-6 rounded"
                  />
                  <span class="font-medium">{{ item }}</span>
                </div>
                <span class="text-xs text-gray-400 dark:text-neutral-500"
                  >Player</span
                >
              </template>
            </div>
          </div>
        </div>
        <button
          class="bg-primary hover:bg-primary/90 text-white text-sm font-medium rounded-lg px-4 py-2 focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="!command.trim() || wsStatus !== 'Connected'"
          @click="sendCommand"
        >
          Send
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from "vue";
import minecraftCommandsRaw from "~/data/minecraft-commands.json";

definePageMeta({
  middleware: "auth",
});

const route = useRoute();
const serverId = route.params.serverId as string;

usePageTitle("Console");

interface ConsoleMessage {
  type: "command" | "response" | "error" | "system";
  text: string;
  timestamp: Date;
}

const messages = ref<ConsoleMessage[]>([]);
const command = ref("");
const wsStatus = ref("Connecting...");
const msgContainer = ref<HTMLElement | null>(null);
const inputRef = ref<HTMLInputElement | null>(null);
const userScrolledUp = ref(false);
const fontSize = ref<string>("text-sm");

let ws: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let unmounted = false;
let reconnectAttempts = 0;
let socketCounter = 0;
let receivedAgentError = false;

const commandHistory = ref<string[]>([]);
let historyIndex = -1;
let savedInput = "";

const MAX_HISTORY = 100;
const RECONNECT_BASE_MS = 5000;
const RECONNECT_MAX_MS = 60000;
const MAX_RECONNECT_ATTEMPTS = 5;

interface MinecraftCommand {
  name: string;
  desc: string;
}

const MINECRAFT_COMMANDS: MinecraftCommand[] = minecraftCommandsRaw;

const showSuggestions = ref(false);
const selectedIndex = ref(0);
const suggestionMode = ref<"command" | "player">("command");

const players = ref<string[]>([]);
let playersPollTimer: ReturnType<typeof setInterval> | null = null;

const PLAYER_COMMANDS = new Set([
  "kick",
  "ban",
  "op",
  "deop",
  "pardon",
  "tp",
  "teleport",
  "give",
  "gamemode",
  "clear",
  "effect",
  "enchant",
  "kill",
  "msg",
  "tell",
  "w",
  "whitelist",
]);

const suggestions = computed<(MinecraftCommand | string)[]>(() => {
  const q = command.value.trimStart();
  if (suggestionMode.value === "command") {
    if (!q) return MINECRAFT_COMMANDS.slice(0, 10);
    const prefix = q.startsWith("/") ? q.slice(1) : q;
    if (!prefix) return MINECRAFT_COMMANDS.slice(0, 10);
    return MINECRAFT_COMMANDS.filter((c) =>
      c.name.toLowerCase().startsWith(prefix.toLowerCase()),
    ).slice(0, 10);
  }
  // player mode
  const lastSpace = q.lastIndexOf(" ");
  const prefix = lastSpace >= 0 ? q.slice(lastSpace + 1) : q;
  if (!prefix) return players.value.slice(0, 10);
  return players.value
    .filter((p) => p.toLowerCase().startsWith(prefix.toLowerCase()))
    .slice(0, 10);
});

function onInput() {
  updateSuggestions();
}

function updateSuggestions() {
  const q = command.value.trimStart();
  if (!q) {
    suggestionMode.value = "command";
    selectedIndex.value = 0;
    showSuggestions.value = true;
    return;
  }

  const trimmed = q.startsWith("/") ? q.slice(1) : q;
  const spaceIdx = trimmed.indexOf(" ");

  if (spaceIdx === -1) {
    suggestionMode.value = "command";
    selectedIndex.value = 0;
    showSuggestions.value = true;
    return;
  }

  const baseCmd = trimmed.slice(0, spaceIdx).toLowerCase();
  const afterCmd = trimmed.slice(spaceIdx + 1);

  if (baseCmd === "whitelist") {
    const subArgs = afterCmd.trim().split(/\s+/);
    if (
      subArgs.length === 1 &&
      (subArgs[0] === "add" || subArgs[0] === "remove")
    ) {
      suggestionMode.value = "player";
      selectedIndex.value = 0;
      showSuggestions.value = players.value.length > 0;
      return;
    }
    showSuggestions.value = false;
    return;
  }

  if (PLAYER_COMMANDS.has(baseCmd)) {
    suggestionMode.value = "player";
    selectedIndex.value = 0;
    showSuggestions.value = players.value.length > 0;
    return;
  }

  showSuggestions.value = false;
}

function closeSuggestions() {
  showSuggestions.value = false;
}

function acceptSuggestion(index: number) {
  const items = suggestions.value;
  if (index < 0 || index >= items.length) return;
  const item = items[index];
  if (!item) return;

  if (suggestionMode.value === "command") {
    const cmd = item as MinecraftCommand;
    command.value = command.value.trimStart().startsWith("/")
      ? `/${cmd.name} `
      : `${cmd.name} `;
  } else {
    const name = item as string;
    const q = command.value;
    const lastSpace = q.lastIndexOf(" ");
    if (lastSpace >= 0) {
      command.value = q.slice(0, lastSpace + 1) + name + " ";
    } else {
      command.value = name + " ";
    }
  }
  showSuggestions.value = false;
  nextTick(() => inputRef.value?.focus());
}

function onEnter() {
  if (showSuggestions.value && suggestions.value.length) {
    acceptSuggestion(selectedIndex.value);
  } else {
    sendCommand();
  }
}

function onUp() {
  if (showSuggestions.value && suggestions.value.length) {
    selectedIndex.value =
      selectedIndex.value <= 0
        ? suggestions.value.length - 1
        : selectedIndex.value - 1;
  } else {
    historyPrev();
  }
}

function onDown() {
  if (showSuggestions.value && suggestions.value.length) {
    selectedIndex.value =
      selectedIndex.value >= suggestions.value.length - 1
        ? 0
        : selectedIndex.value + 1;
  } else {
    historyNext();
  }
}

function onTab(e: KeyboardEvent) {
  if (showSuggestions.value && suggestions.value.length) {
    e.preventDefault();
    acceptSuggestion(selectedIndex.value);
  }
}

async function fetchPlayers() {
  try {
    const res = await $fetch<{ players?: string[] }>(
      `/api/servers/${serverId}/players`,
      { credentials: "include" },
    );
    players.value = res.players || [];
  } catch {
    // silently fail
  }
}

function formatTime(d: Date): string {
  return d.toLocaleTimeString("en-US", {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function addMessage(type: ConsoleMessage["type"], text: string) {
  messages.value.push({ type, text, timestamp: new Date() });
  nextTick(() => {
    if (!userScrolledUp.value) {
      scrollToBottom();
    }
  });
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
  const base = (config.public.apiBase as string) || "http://localhost:8081";
  let wsBase: string;
  if (base.startsWith("/")) {
    const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    wsBase = `${proto}//${window.location.host}${base === "/" ? "" : base}`;
  } else {
    wsBase = base.replace(/^http/, "ws");
  }
  const token = useCookie("auth_token").value || "";
  const url = `${wsBase}/ws/servers/${serverId}/rcon?token=${encodeURIComponent(token)}`;

  const socketId = ++socketCounter;
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

function sendCommand() {
  let cmd = command.value.trim();
  if (!cmd || !ws || ws.readyState !== WebSocket.OPEN) return;
  if (cmd.startsWith("/")) {
    cmd = cmd.slice(1);
  }

  ws.send(JSON.stringify({ command: cmd }));
  addMessage("command", command.value.trim());

  // Update history
  commandHistory.value.push(command.value.trim());
  if (commandHistory.value.length > MAX_HISTORY) {
    commandHistory.value.shift();
  }
  historyIndex = -1;
  command.value = "";
  showSuggestions.value = false;
}

function historyPrev() {
  if (commandHistory.value.length === 0) return;
  if (historyIndex === -1) {
    savedInput = command.value;
  }
  historyIndex++;
  if (historyIndex >= commandHistory.value.length) {
    historyIndex = commandHistory.value.length - 1;
  }
  command.value =
    commandHistory.value[commandHistory.value.length - 1 - historyIndex] || "";
}

function historyNext() {
  if (commandHistory.value.length === 0 || historyIndex === -1) return;
  historyIndex--;
  if (historyIndex < 0) {
    historyIndex = -1;
    command.value = savedInput;
  } else {
    command.value =
      commandHistory.value[commandHistory.value.length - 1 - historyIndex] ||
      "";
  }
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

function scrollToBottom() {
  const el = msgContainer.value;
  if (el) {
    el.scrollTop = el.scrollHeight;
  }
}

function onScroll() {
  const el = msgContainer.value;
  if (!el) return;
  const threshold = 20;
  const isAtBottom =
    el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
  userScrolledUp.value = !isAtBottom;
}

onMounted(() => {
  const saved = localStorage.getItem("console-font-size");
  if (saved) {
    fontSize.value = saved;
  }
  connect();
  inputRef.value?.focus();
  fetchPlayers();
  playersPollTimer = setInterval(fetchPlayers, 30000);
});

watch(fontSize, (value) => {
  localStorage.setItem("console-font-size", value);
});

onUnmounted(() => {
  unmounted = true;
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (playersPollTimer) {
    clearInterval(playersPollTimer);
    playersPollTimer = null;
  }
  if (ws) {
    const oldWs = ws;
    ws = null;
    oldWs.onclose = null;
    oldWs.close();
  }
});
</script>
