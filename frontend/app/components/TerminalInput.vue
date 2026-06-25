<template>
  <div class="mt-4 flex items-center gap-2">
    <div class="flex-1 relative">
      <input
        ref="inputRef"
        v-model="command"
        type="text"
        placeholder="Type a command and press Enter..."
        class="w-full bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 rounded-lg px-3 py-2 text-sm text-gray-900 dark:text-white font-mono focus:outline-none focus:ring-2 focus:ring-primary"
        :disabled="disabled"
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
                @error="onAvatarError"
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
      :disabled="!command.trim() || disabled"
      @click="send"
    >
      Send
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from "vue";
import minecraftCommandsRaw from "~/data/minecraft-commands.json";

const props = defineProps<{
  disabled?: boolean;
  players?: string[];
}>();

const emit = defineEmits<{
  send: [command: string];
}>();

interface MinecraftCommand {
  name: string;
  desc: string;
}

const MINECRAFT_COMMANDS: MinecraftCommand[] = minecraftCommandsRaw;

const command = ref("");
const inputRef = ref<HTMLInputElement | null>(null);
const showSuggestions = ref(false);
const selectedIndex = ref(0);
const suggestionMode = ref<"command" | "player">("command");

const commandHistory = ref<string[]>([]);
let historyIndex = -1;
let savedInput = "";
const MAX_HISTORY = 100;

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
  const playerList = props.players || [];
  if (!prefix) return playerList.slice(0, 10);
  return playerList
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
      showSuggestions.value = (props.players || []).length > 0;
      return;
    }
    showSuggestions.value = false;
    return;
  }

  if (PLAYER_COMMANDS.has(baseCmd)) {
    suggestionMode.value = "player";
    selectedIndex.value = 0;
    showSuggestions.value = (props.players || []).length > 0;
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
    send();
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

function send() {
  const cmd = command.value.trim();
  if (!cmd || props.disabled) return;
  emit("send", cmd);
  commandHistory.value.push(cmd);
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

function onAvatarError(e: Event) {
  const img = e.target as HTMLImageElement | null;
  if (img) {
    img.src = "/img/steve-head-32.png";
  }
}

function focus() {
  inputRef.value?.focus();
}

defineExpose({ focus });
</script>
