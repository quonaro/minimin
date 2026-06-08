<template>
  <div class="p-6 h-[calc(100vh-4rem)] flex flex-col">
    <div class="mb-4 flex items-center justify-between flex-wrap gap-3">
      <div class="flex items-center gap-4">
        <NuxtLink
          :to="`/agent/${agentId}/servers/${serverId}`"
          class="inline-flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-4 h-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M10 19l-7-7m0 0l7-7m-7 7h18"
            />
          </svg>
          <span>Back to server</span>
        </NuxtLink>
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
              'text-gray-500 dark:text-gray-400': !wsStatus,
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
              'text-gray-500 dark:text-gray-400': !wsStatus,
            }"
          >
            {{ wsStatus || "Connecting..." }}
          </span>
          <button
            class="text-xs leading-none rounded-full bg-blue-600 text-white hover:bg-blue-700 focus:outline-none px-2 py-0.5"
            @click="reconnect"
          >
            Reconnect
          </button>
        </div>

        <button
          class="text-xs leading-none rounded-full bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600 focus:outline-none px-2 py-0.5"
          @click="clearMessages"
        >
          Clear
        </button>
      </div>
    </div>

    <div
      ref="msgContainer"
      :class="[
        'flex-1 min-h-0 bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 font-mono rounded-xl p-4 overflow-y-auto shadow-inner',
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
          class="flex gap-2 py-0.5 border-b border-gray-200 dark:border-gray-800/40 hover:bg-gray-100 dark:hover:bg-gray-800/60 transition-colors"
        >
          <span
            class="text-gray-400 dark:text-gray-600 select-none tabular-nums shrink-0"
          >
            {{ formatTime(msg.timestamp) }}
          </span>
          <span
            v-if="msg.type === 'command'"
            class="whitespace-pre-wrap break-words text-blue-600 dark:text-blue-400"
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
            class="whitespace-pre-wrap break-words text-gray-700 dark:text-gray-300"
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
          class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-900 dark:text-white font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
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
          class="absolute bottom-full left-0 right-0 mb-1 max-h-60 overflow-y-auto rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-lg z-50"
        >
          <div
            v-for="(cmd, i) in filteredCommands"
            :key="cmd.name"
            class="px-3 py-2 cursor-pointer text-sm font-mono flex items-center justify-between"
            :class="{
              'bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300':
                selectedIndex === i,
              'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700/50':
                selectedIndex !== i,
            }"
            @click="acceptSuggestion(i)"
            @mouseenter="selectedIndex = i"
          >
            <span>{{ cmd.name }}</span>
            <span
              class="text-xs text-gray-400 dark:text-gray-500 truncate ml-2 max-w-[60%]"
              >{{ cmd.desc }}</span
            >
          </div>
        </div>
      </div>
      <button
        class="bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg px-4 py-2 focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
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
const RECONNECT_BASE_MS = 1000;
const RECONNECT_MAX_MS = 30000;
const MAX_RECONNECT_ATTEMPTS = 5;

const MINECRAFT_COMMANDS: { name: string; desc: string }[] = [
  { name: "advancement", desc: "Give, remove, or check player advancements" },
  { name: "ban", desc: "Add player to banlist" },
  { name: "ban-ip", desc: "Add IP address to banlist" },
  { name: "banlist", desc: "Display banlist" },
  { name: "bossbar", desc: "Create and modify boss bars" },
  { name: "clear", desc: "Clear items from player inventory" },
  { name: "clone", desc: "Copy blocks from one area to another" },
  { name: "data", desc: "Get, merge, remove, or modify NBT data" },
  { name: "datapack", desc: "Enable, disable, or list data packs" },
  { name: "debug", desc: "Start or stop debug profiling" },
  { name: "defaultgamemode", desc: "Set default game mode" },
  { name: "deop", desc: "Revoke operator status" },
  { name: "difficulty", desc: "Set difficulty level" },
  { name: "effect", desc: "Add or remove status effects" },
  { name: "enchant", desc: "Enchant a player item" },
  { name: "execute", desc: "Execute another command" },
  { name: "experience", desc: "Add or set player experience" },
  { name: "fill", desc: "Fill a region with a specific block" },
  { name: "forceload", desc: "Force chunks to load" },
  { name: "function", desc: "Run a function" },
  { name: "gamemode", desc: "Set player game mode" },
  { name: "gamerule", desc: "Set or query a game rule value" },
  { name: "give", desc: "Give an item to a player" },
  { name: "help", desc: "Show list of commands or help for a command" },
  { name: "kick", desc: "Kick a player from the server" },
  { name: "kill", desc: "Kill entities" },
  { name: "list", desc: "List players on the server" },
  { name: "locate", desc: "Locate the closest structure" },
  { name: "loot", desc: "Drop items from an inventory or spawn loot" },
  { name: "me", desc: "Display a message about yourself" },
  { name: "msg", desc: "Send a private message" },
  { name: "op", desc: "Grant operator status" },
  { name: "pardon", desc: "Remove player from banlist" },
  { name: "pardon-ip", desc: "Remove IP from banlist" },
  { name: "particle", desc: "Create particles" },
  { name: "playsound", desc: "Play a sound" },
  { name: "publish", desc: "Open server to LAN" },
  { name: "recipe", desc: "Give or take recipes" },
  { name: "reload", desc: "Reload data packs and functions" },
  { name: "save-all", desc: "Save the server" },
  { name: "save-off", desc: "Disable automatic saving" },
  { name: "save-on", desc: "Enable automatic saving" },
  { name: "say", desc: "Broadcast a message" },
  { name: "schedule", desc: "Schedule a function" },
  { name: "scoreboard", desc: "Manage scoreboard objectives and players" },
  { name: "seed", desc: "Display the world seed" },
  { name: "setblock", desc: "Set a block" },
  { name: "setidletimeout", desc: "Set idle kick timer" },
  { name: "setworldspawn", desc: "Set world spawn point" },
  { name: "spawnpoint", desc: "Set player spawn point" },
  { name: "spectate", desc: "Spectate an entity" },
  { name: "spreadplayers", desc: "Spread players around a point" },
  { name: "stop", desc: "Stop the server" },
  { name: "stopsound", desc: "Stop a sound" },
  { name: "summon", desc: "Summon an entity" },
  { name: "tag", desc: "Manage entity tags" },
  { name: "team", desc: "Manage teams" },
  { name: "teleport", desc: "Teleport entities" },
  { name: "tell", desc: "Send a private message" },
  { name: "tellraw", desc: "Send a JSON message to players" },
  { name: "time", desc: "Change or query the world time" },
  { name: "title", desc: "Manage screen titles" },
  { name: "tp", desc: "Teleport entities (alias)" },
  { name: "transfer", desc: "Transfer a player to another server" },
  { name: "trigger", desc: "Set a trigger objective" },
  { name: "w", desc: "Send a private message (alias)" },
  { name: "weather", desc: "Set the weather" },
  { name: "whitelist", desc: "Manage the server whitelist" },
  { name: "worldborder", desc: "Manage the world border" },
  { name: "xp", desc: "Add or set experience (alias)" },
];

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
