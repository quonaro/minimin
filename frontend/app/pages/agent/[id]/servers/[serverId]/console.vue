<template>
  <div class="p-6 h-[calc(100vh-4rem)] flex flex-col">
    <div class="mb-4 flex items-center justify-between flex-wrap gap-3">
      <div class="flex items-center gap-4">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          Console
        </h1>
      </div>

      <div class="flex items-center gap-4">
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

        <button
          class="text-xs leading-none rounded-full bg-gray-200 dark:bg-neutral-700 text-gray-700 dark:text-neutral-300 hover:bg-gray-300 dark:hover:bg-neutral-600 focus:outline-none px-2 py-0.5"
          @click="clearMessages"
        >
          Clear
        </button>
      </div>
    </div>

    <div
      ref="msgContainer"
      :class="[
        'flex-1 min-h-0 bg-white dark:bg-neutral-900 text-gray-900 dark:text-neutral-100 font-mono rounded-xl p-4 overflow-y-auto shadow-inner',
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
          class="flex gap-2 py-0.5 border-b border-gray-200 dark:border-neutral-800/40 hover:bg-gray-100 dark:hover:bg-neutral-800/60 transition-colors"
        >
          <span
            class="text-gray-400 dark:text-neutral-600 select-none tabular-nums shrink-0"
          >
            {{ formatTime(msg.timestamp) }}
          </span>
          <span
            v-if="msg.type === 'command'"
            class="whitespace-pre-wrap break-words text-primary dark:text-primary-400"
          >
            > {{ msg.text }}
          </span>
          <span
            v-else-if="msg.type === 'error'"
            class="whitespace-pre-wrap break-words text-red-600 dark:text-red-400"
          >
            {{ msg.text }}
          </span>
          <span
            v-else
            class="whitespace-pre-wrap break-words text-gray-700 dark:text-neutral-300"
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
          v-if="showSuggestions && filteredCommands.length"
          class="absolute bottom-full left-0 right-0 mb-1 max-h-60 overflow-y-auto rounded-lg border border-gray-200 dark:border-neutral-700 bg-white dark:bg-neutral-800 shadow-lg z-50"
        >
          <div
            v-for="(cmd, i) in filteredCommands"
            :key="cmd.name"
            class="px-3 py-2 cursor-pointer text-sm font-mono flex items-center justify-between"
            :class="{
              'bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300':
                selectedIndex === i,
              'text-gray-700 dark:text-neutral-300 hover:bg-gray-50 dark:hover:bg-neutral-700/50':
                selectedIndex !== i,
            }"
            @click="acceptSuggestion(i)"
            @mouseenter="selectedIndex = i"
          >
            <span>{{ cmd.name }}</span>
            <span
              class="text-xs text-gray-400 dark:text-neutral-500 truncate ml-2 max-w-[60%]"
              >{{ cmd.desc }}</span
            >
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
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from "vue";
import minecraftCommandsRaw from "~/data/minecraft-commands.json";

definePageMeta({
  middleware: "auth",
});

const route = useRoute();
const { agentId } = useCurrentAgent();
const serverId = route.params.serverId as string;

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

const filteredCommands = computed(() => {
  const q = command.value.trimStart();
  if (!q) return MINECRAFT_COMMANDS.slice(0, 10);
  const prefix = q.startsWith("/") ? q.slice(1) : q;
  if (!prefix) return MINECRAFT_COMMANDS.slice(0, 10);
  return MINECRAFT_COMMANDS.filter((c) =>
    c.name.toLowerCase().startsWith(prefix.toLowerCase()),
  ).slice(0, 10);
});

function onInput() {
  updateSuggestions();
}

function updateSuggestions() {
  const q = command.value.trimStart();
  if (!q || q.includes(" ")) {
    showSuggestions.value = false;
    return;
  }
  selectedIndex.value = 0;
  showSuggestions.value = true;
}

function closeSuggestions() {
  showSuggestions.value = false;
}

function acceptSuggestion(index: number) {
  const cmds = filteredCommands.value;
  if (index < 0 || index >= cmds.length) return;
  const item = cmds[index];
  if (!item) return;
  command.value = command.value.trimStart().startsWith("/")
    ? `/${item.name} `
    : `${item.name} `;
  showSuggestions.value = false;
  nextTick(() => inputRef.value?.focus());
}

function onEnter() {
  if (showSuggestions.value && filteredCommands.value.length) {
    acceptSuggestion(selectedIndex.value);
  } else {
    sendCommand();
  }
}

function onUp() {
  if (showSuggestions.value && filteredCommands.value.length) {
    selectedIndex.value =
      selectedIndex.value <= 0
        ? filteredCommands.value.length - 1
        : selectedIndex.value - 1;
  } else {
    historyPrev();
  }
}

function onDown() {
  if (showSuggestions.value && filteredCommands.value.length) {
    selectedIndex.value =
      selectedIndex.value >= filteredCommands.value.length - 1
        ? 0
        : selectedIndex.value + 1;
  } else {
    historyNext();
  }
}

function onTab(e: KeyboardEvent) {
  if (showSuggestions.value && filteredCommands.value.length) {
    e.preventDefault();
    acceptSuggestion(selectedIndex.value);
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
  if (!agentId.value) {
    return;
  }

  const config = useRuntimeConfig();
  const base = (config.public.apiBase as string) || "http://localhost:8081";
  const wsBase = base.replace(/^http/, "ws");
  const url = `${wsBase}/ws/agent/${agentId.value}/servers/${serverId}/rcon`;

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
});

watch(fontSize, (value) => {
  localStorage.setItem("console-font-size", value);
});

watch(
  () => agentId.value,
  (newVal, oldVal) => {
    if (newVal && newVal !== oldVal) {
      reconnect();
    }
  },
);

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
</script>
