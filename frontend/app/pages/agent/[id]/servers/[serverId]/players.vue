<template>
  <div class="p-6">
    <server-nav class="mb-4" />
    <div class="mb-4 flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          Players
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
            @click="reconnectWS"
          >
            Reconnect
          </button>
        </div>
        <button
          class="text-xs leading-none rounded-full bg-gray-200 dark:bg-neutral-700 text-gray-700 dark:text-neutral-300 hover:bg-gray-300 dark:hover:bg-neutral-600 focus:outline-none px-2 py-0.5"
          :disabled="refreshing"
          @click="refreshAll"
        >
          {{ refreshing ? "Refreshing..." : "Refresh" }}
        </button>
      </div>
    </div>

    <!-- Disconnected Banner -->
    <div
      v-if="
        (wsStatus === 'Error' || wsStatus === 'Disconnected') &&
        reconnectAttempts > 0
      "
      class="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4 flex items-center gap-3"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        class="w-5 h-5 text-red-600 dark:text-red-400 shrink-0"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
        />
      </svg>
      <div class="flex-1">
        <p class="text-sm font-medium text-red-800 dark:text-red-300">
          Disconnected from server
        </p>
        <p class="text-xs text-red-600 dark:text-red-400">
          Retrying connection... (attempt {{ reconnectAttempts }})
        </p>
      </div>
    </div>

    <!-- Offline Actions -->
    <div
      class="mb-6 bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-xl p-5"
    >
      <h2 class="text-lg font-bold text-gray-900 dark:text-white mb-3">
        Offline Actions
      </h2>
      <p class="text-sm text-gray-500 dark:text-neutral-400 mb-3">
        Manage a player who has never joined.
      </p>
      <div
        class="flex flex-col sm:flex-row items-stretch sm:items-center gap-3"
      >
        <input
          v-model="offlineName"
          type="text"
          placeholder="Player name..."
          class="flex-1 px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none text-sm"
          @keydown.enter.prevent="sendRcon(`ban ${offlineName.trim()}`)"
        />
        <div class="flex items-center gap-2 shrink-0">
          <button
            class="text-sm px-3 py-2 rounded bg-red-600 text-white hover:bg-red-700"
            @click="openReasonModal('ban', offlineName.trim())"
          >
            Ban
          </button>
          <button
            class="text-sm px-3 py-2 rounded bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400 hover:bg-primary-200 dark:hover:bg-primary-900/50"
            @click="sendRcon(`op ${offlineName.trim()}`)"
          >
            Op
          </button>
          <button
            class="text-sm px-3 py-2 rounded bg-gray-100 text-gray-700 dark:bg-neutral-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-600"
            @click="sendRcon(`whitelist add ${offlineName.trim()}`)"
          >
            Whitelist
          </button>
          <button
            class="text-sm px-3 py-2 rounded bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 hover:bg-green-200 dark:hover:bg-green-900/50"
            @click="sendRcon(`pardon ${offlineName.trim()}`)"
          >
            Unban
          </button>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Online Players -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-xl p-5"
      >
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white">
            Online Players
          </h2>
          <span
            class="text-xs font-medium px-2 py-1 rounded-full bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400"
          >
            {{ onlinePlayers.length }} / {{ maxPlayers }}
          </span>
        </div>
        <div
          v-if="onlineLoading"
          class="text-gray-500 dark:text-neutral-400 text-sm"
        >
          Loading...
        </div>
        <div
          v-else-if="onlineError"
          class="text-red-500 dark:text-red-400 text-sm"
        >
          {{ onlineError }}
        </div>
        <div
          v-else-if="onlinePlayers.length === 0"
          class="text-gray-500 dark:text-neutral-400 text-sm"
        >
          No players online.
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="name in onlinePlayers"
            :key="name"
            class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
          >
            <img
              :src="`https://cravatar.eu/helmavatar/${encodeURIComponent(name)}/32.png`"
              alt=""
              class="w-8 h-8 rounded"
            />
            <span
              class="flex-1 text-sm font-medium text-gray-900 dark:text-white"
            >
              {{ name }}
            </span>
            <div class="flex items-center gap-1">
              <button
                class="text-xs px-2 py-1 rounded bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400 hover:bg-red-200 dark:hover:bg-red-900/50"
                @click="openReasonModal('kick', name)"
              >
                Kick
              </button>
              <button
                class="text-xs px-2 py-1 rounded bg-red-600 text-white hover:bg-red-700"
                @click="openReasonModal('ban', name)"
              >
                Ban
              </button>
              <button
                class="text-xs px-2 py-1 rounded bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400 hover:bg-primary-200 dark:hover:bg-primary-900/50"
                @click="sendRcon(`op ${name}`)"
              >
                Op
              </button>
              <button
                class="text-xs px-2 py-1 rounded bg-gray-100 text-gray-700 dark:bg-neutral-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-600"
                @click="sendRcon(`whitelist add ${name}`)"
              >
                WL
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Operators -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-xl p-5"
      >
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white">
            Operators
          </h2>
        </div>
        <div
          v-if="opsLoading"
          class="text-gray-500 dark:text-neutral-400 text-sm"
        >
          Loading...
        </div>
        <div
          v-else-if="opsError"
          class="text-red-500 dark:text-red-400 text-sm"
        >
          {{ opsError }}
        </div>
        <div
          v-else-if="opsList.length === 0"
          class="text-gray-500 dark:text-neutral-400 text-sm"
        >
          No operators.
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="op in opsList"
            :key="op.uuid || op.name"
            class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
          >
            <img
              v-if="op.name"
              :src="`https://cravatar.eu/helmavatar/${encodeURIComponent(op.name)}/32.png`"
              alt=""
              class="w-8 h-8 rounded"
            />
            <span
              class="flex-1 text-sm font-medium text-gray-900 dark:text-white"
            >
              {{ op.name }}
            </span>
            <button
              class="text-xs px-2 py-1 rounded bg-gray-100 text-gray-700 dark:bg-neutral-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-600"
              @click="sendRcon(`deop ${op.name}`)"
            >
              Deop
            </button>
          </div>
        </div>
      </div>

      <!-- Whitelist -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-xl p-5"
      >
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white">
            Whitelist
          </h2>
        </div>
        <div
          v-if="wlLoading"
          class="text-gray-500 dark:text-neutral-400 text-sm"
        >
          Loading...
        </div>
        <div v-else-if="wlError" class="text-red-500 dark:text-red-400 text-sm">
          {{ wlError }}
        </div>
        <div
          v-else-if="wlList.length === 0"
          class="text-gray-500 dark:text-neutral-400 text-sm"
        >
          Whitelist is empty.
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="entry in wlList"
            :key="entry.uuid || entry.name"
            class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
          >
            <img
              v-if="entry.name"
              :src="`https://cravatar.eu/helmavatar/${encodeURIComponent(entry.name)}/32.png`"
              alt=""
              class="w-8 h-8 rounded"
            />
            <span
              class="flex-1 text-sm font-medium text-gray-900 dark:text-white"
            >
              {{ entry.name }}
            </span>
            <button
              class="text-xs px-2 py-1 rounded bg-gray-100 text-gray-700 dark:bg-neutral-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-600"
              @click="sendRcon(`whitelist remove ${entry.name}`)"
            >
              Remove
            </button>
          </div>
        </div>
      </div>

      <!-- Banned Players -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-xl p-5"
      >
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white">
            Banned Players
          </h2>
        </div>
        <div
          v-if="bansLoading"
          class="text-gray-500 dark:text-neutral-400 text-sm"
        >
          Loading...
        </div>
        <div
          v-else-if="bansError"
          class="text-red-500 dark:text-red-400 text-sm"
        >
          {{ bansError }}
        </div>
        <div
          v-else-if="bansList.length === 0"
          class="text-gray-500 dark:text-neutral-400 text-sm"
        >
          No banned players.
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="ban in bansList"
            :key="ban.uuid || ban.name"
            class="flex items-start gap-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
          >
            <img
              v-if="ban.name"
              :src="`https://cravatar.eu/helmavatar/${encodeURIComponent(ban.name)}/32.png`"
              alt=""
              class="w-8 h-8 rounded mt-0.5"
            />
            <div class="flex-1 min-w-0">
              <div class="text-sm font-medium text-gray-900 dark:text-white">
                {{ ban.name }}
              </div>
              <div
                v-if="ban.reason"
                class="text-xs text-gray-500 dark:text-neutral-400 truncate"
              >
                Reason: {{ ban.reason }}
              </div>
              <div
                v-if="ban.created"
                class="text-xs text-gray-500 dark:text-neutral-400"
              >
                {{ ban.created }}
              </div>
            </div>
            <button
              class="text-xs px-2 py-1 rounded bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 hover:bg-green-200 dark:hover:bg-green-900/50 shrink-0"
              @click="sendRcon(`pardon ${ban.name}`)"
            >
              Unban
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Event History -->
    <div
      class="mt-6 bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-xl p-5"
    >
      <h2 class="text-lg font-bold text-gray-900 dark:text-white mb-3">
        Event History
      </h2>
      <PlayerEventLog :events="eventLog" />
    </div>

    <!-- All-Time Players -->
    <div
      class="mt-6 bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-xl p-5"
    >
      <div class="flex items-center justify-between mb-3 gap-3">
        <div class="flex items-center gap-2 shrink-0">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white">
            All Players
          </h2>
          <span class="text-xs text-gray-500 dark:text-neutral-400">
            {{ allPlayers.length }}
          </span>
        </div>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search players..."
          class="max-w-[16rem] px-3 py-1.5 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none text-sm"
        />
      </div>
      <PlayerAllTimeList
        :players="allPlayers"
        :filter="searchQuery"
        :online="onlinePlayers"
        @kick="openReasonModal('kick', $event)"
        @ban="openReasonModal('ban', $event)"
        @op="sendRcon(`op ${$event}`)"
        @wladd="sendRcon(`whitelist add ${$event}`)"
      />
    </div>

    <!-- Reason Modal -->
    <div
      v-if="modalOpen"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      @click.self="closeModal"
    >
      <div
        class="bg-white dark:bg-neutral-800 rounded-xl p-6 w-full max-w-sm shadow-xl"
      >
        <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-2">
          {{ modalTitle }}
        </h3>
        <p class="text-sm text-gray-500 dark:text-neutral-400 mb-4">
          Target:
          <span class="font-medium text-gray-900 dark:text-white">{{
            modalTarget
          }}</span>
        </p>
        <input
          v-model="modalReason"
          type="text"
          placeholder="Reason (optional)"
          class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none mb-4"
          @keydown.enter.prevent="confirmModal"
        />
        <div class="flex justify-end gap-2">
          <button
            class="px-3 py-2 rounded-lg text-sm text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700"
            @click="closeModal"
          >
            Cancel
          </button>
          <button
            class="px-3 py-2 rounded-lg text-sm bg-red-600 text-white hover:bg-red-700"
            @click="confirmModal"
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import PlayerEventLog, {
  type PlayerEvent,
} from "~/components/PlayerEventLog.vue";
import PlayerAllTimeList, {
  type AllTimePlayer,
} from "~/components/PlayerAllTimeList.vue";

definePageMeta({
  middleware: "auth",
});

const route = useRoute();
const { agentId } = useCurrentAgent();
const { show: showToast } = useToast();
const serverId = route.params.serverId as string;

const apiBase = useApiBase();

interface PlayerEntry {
  name: string;
  uuid?: string;
  reason?: string;
  created?: string;
  level?: number;
}

const onlineLoading = ref(true);
const onlineError = ref<string | null>(null);
const onlinePlayers = ref<string[]>([]);
const maxPlayers = ref(0);

const opsLoading = ref(true);
const opsError = ref<string | null>(null);
const opsList = ref<PlayerEntry[]>([]);

const wlLoading = ref(true);
const wlError = ref<string | null>(null);
const wlList = ref<PlayerEntry[]>([]);

const bansLoading = ref(true);
const bansError = ref<string | null>(null);
const bansList = ref<PlayerEntry[]>([]);

const eventLog = ref<PlayerEvent[]>([]);
const allPlayers = ref<AllTimePlayer[]>([]);
const searchQuery = ref("");
const offlineName = ref("");

const refreshing = ref(false);
const wsStatus = ref("Connecting...");

let ws: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let unmounted = false;
let reconnectAttempts = 0;
let socketCounter = 0;
let eventCounter = 0;

const RECONNECT_BASE_MS = 5000;
const RECONNECT_MAX_MS = 60000;
const FETCH_RETRY_BASE_MS = 1000;
const FETCH_RETRY_MAX_MS = 30000;
const FETCH_MAX_ATTEMPTS = 5;

const { lastEvent } = useEventSource();

function lsKey(): string {
  return `mc-players-${agentId.value}-${serverId}`;
}

function loadAllPlayers() {
  try {
    const raw = localStorage.getItem(lsKey());
    if (raw) {
      const parsed = JSON.parse(raw) as Record<string, number>;
      allPlayers.value = Object.entries(parsed).map(([name, lastSeen]) => ({
        name,
        lastSeen,
      }));
    }
  } catch {
    allPlayers.value = [];
  }
}

function saveAllPlayers() {
  const map: Record<string, number> = {};
  for (const p of allPlayers.value) {
    map[p.name] = p.lastSeen;
  }
  localStorage.setItem(lsKey(), JSON.stringify(map));
}

function upsertPlayer(name: string, ts: number) {
  const idx = allPlayers.value.findIndex((p) => p.name === name);
  if (idx !== -1) {
    allPlayers.value[idx]!.lastSeen = ts;
  } else {
    allPlayers.value.push({ name, lastSeen: ts });
  }
  saveAllPlayers();
}

function addEvent(ev: Omit<PlayerEvent, "id">) {
  eventLog.value.unshift({ ...ev, id: `${ev.ts}-${++eventCounter}` });
  if (eventLog.value.length > 200) {
    eventLog.value = eventLog.value.slice(0, 200);
  }
}

function stripMC(s: string): string {
  return s.replace(/§./g, "").trim();
}

function parseLogTimestamp(line: string): number {
  const m = line.match(/^\[(\d{2}):(\d{2}):(\d{2})\]/);
  if (!m) return Date.now();

  const now = new Date();
  let ts = Date.UTC(
    now.getUTCFullYear(),
    now.getUTCMonth(),
    now.getUTCDate(),
    parseInt(m[1] as string, 10),
    parseInt(m[2] as string, 10),
    parseInt(m[3] as string, 10),
  );

  if (ts > Date.now()) {
    ts -= 24 * 60 * 60 * 1000;
  }
  return ts;
}

function handleLogLine(line: string) {
  const ts = parseLogTimestamp(line);

  const joinMatch = line.match(/]:\s*(\S+) joined the game/);
  if (joinMatch && joinMatch[1]) {
    const name = stripMC(joinMatch[1] as string);
    if (!onlinePlayers.value.includes(name)) {
      onlinePlayers.value.push(name);
    }
    upsertPlayer(name, ts);
    addEvent({ ts, type: "join", player: name });
    return;
  }

  const leftMatch = line.match(/]:\s*(\S+) left the game/);
  if (leftMatch && leftMatch[1]) {
    const name = stripMC(leftMatch[1] as string);
    const wasOnline = onlinePlayers.value.includes(name);
    onlinePlayers.value = onlinePlayers.value.filter((n) => n !== name);
    upsertPlayer(name, ts);
    if (wasOnline) {
      addEvent({ ts, type: "leave", player: name });
    }
    return;
  }

  const lostMatch = line.match(/]:\s*(\S+) lost connection:/);
  if (lostMatch && lostMatch[1]) {
    const name = stripMC(lostMatch[1] as string);
    const wasOnline = onlinePlayers.value.includes(name);
    onlinePlayers.value = onlinePlayers.value.filter((n) => n !== name);
    upsertPlayer(name, ts);
    if (wasOnline) {
      addEvent({ ts, type: "leave", player: name });
    }
    return;
  }

  const opMatch = line.match(/Made (\S+) a server operator/);
  if (opMatch && opMatch[1]) {
    const name = stripMC(opMatch[1] as string);
    addEvent({ ts, type: "op", player: name });
    if (!opsList.value.some((op) => op.name === name)) {
      opsList.value.push({ name });
    }
    return;
  }

  const deopMatch = line.match(/Made (\S+) no longer a server operator/);
  if (deopMatch && deopMatch[1]) {
    const name = stripMC(deopMatch[1] as string);
    addEvent({ ts, type: "deop", player: name });
    opsList.value = opsList.value.filter((op) => op.name !== name);
    return;
  }

  const banMatch = line.match(/Banned player ([^:\s]+)(?::\s*(.*))?/);
  if (banMatch && banMatch[1]) {
    const name = stripMC(banMatch[1] as string);
    const reason = banMatch[2]?.trim() || undefined;
    addEvent({ ts, type: "ban", player: name, reason });
    if (!bansList.value.some((ban) => ban.name === name)) {
      bansList.value.push({
        name,
        reason,
        created: new Date(ts).toISOString(),
      });
    }
    return;
  }

  const unbanMatch = line.match(/Unbanned player (\S+)/);
  if (unbanMatch && unbanMatch[1]) {
    const name = stripMC(unbanMatch[1] as string);
    addEvent({ ts, type: "unban", player: name });
    bansList.value = bansList.value.filter((ban) => ban.name !== name);
    return;
  }

  const wlAddMatch = line.match(/Added (\S+) to the whitelist/);
  if (wlAddMatch && wlAddMatch[1]) {
    const name = stripMC(wlAddMatch[1] as string);
    addEvent({ ts, type: "wladd", player: name });
    if (!wlList.value.some((entry) => entry.name === name)) {
      wlList.value.push({ name });
    }
    return;
  }

  const wlRemMatch = line.match(/Removed (\S+) from the whitelist/);
  if (wlRemMatch && wlRemMatch[1]) {
    const name = stripMC(wlRemMatch[1] as string);
    addEvent({ ts, type: "wlremove", player: name });
    wlList.value = wlList.value.filter((entry) => entry.name !== name);
    return;
  }

  const kickMatch = line.match(/Kicked ([^:\s]+)(?::\s*(.*))?/);
  if (kickMatch && kickMatch[1]) {
    const name = stripMC(kickMatch[1] as string);
    onlinePlayers.value = onlinePlayers.value.filter((n) => n !== name);
    addEvent({
      ts,
      type: "kick",
      player: name,
      reason: kickMatch[2]?.trim() || undefined,
    });
    return;
  }

  const deathMatch = line.match(
    /]:\s*(\S+) (was .+|died|drowned|fell .+|burned .+|suffocated .+|tried to swim .+|froze to death|starved to death|withered away|hit the ground too hard|experienced kinetic energy|went up in flames|discovered .+)/,
  );
  if (deathMatch && deathMatch[1]) {
    addEvent({
      ts,
      type: "death",
      player: stripMC(deathMatch[1] as string),
      reason: stripMC(deathMatch[2] as string),
    });
    return;
  }

  const advMatch = line.match(
    /]:\s*(\S+) has (made the advancement|reached the goal|completed the challenge) \[(.+)\]/,
  );
  if (advMatch && advMatch[1]) {
    addEvent({
      ts,
      type: "advancement",
      player: stripMC(advMatch[1] as string),
      reason: stripMC(advMatch[3] as string),
    });
    return;
  }

  const obtMatch = line.match(/]:\s*(\S+) has obtained \[(.+)\]/);
  if (obtMatch && obtMatch[1]) {
    addEvent({
      ts,
      type: "obtained",
      player: stripMC(obtMatch[1] as string),
      reason: stripMC(obtMatch[2] as string),
    });
  }
}

function connectLogsWS() {
  if (ws) {
    const oldWs = ws;
    ws = null;
    oldWs.onopen = null;
    oldWs.onmessage = null;
    oldWs.onerror = null;
    oldWs.onclose = null;
    oldWs.close();
  }
  if (!agentId.value) return;

  const config = useRuntimeConfig();
  const base = (config.public.apiBase as string) || "http://localhost:8081";
  const wsBase = base.replace(/^http/, "ws");
  const url = `${wsBase}/ws/agent/${agentId.value}/servers/${serverId}/logs?tail=50`;

  ++socketCounter;
  const socket = new WebSocket(url);
  ws = socket;

  socket.onopen = () => {
    if (ws !== socket) return;
    reconnectAttempts = 0;
    wsStatus.value = "Connected";
  };

  socket.onmessage = (e) => {
    if (ws !== socket) return;
    const text = String(e.data);
    const lines = text.split("\n");
    for (const line of lines) {
      handleLogLine(line);
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
      const delay = Math.min(
        RECONNECT_BASE_MS * 2 ** reconnectAttempts,
        RECONNECT_MAX_MS,
      );
      reconnectAttempts++;
      reconnectTimer = setTimeout(connectLogsWS, delay);
    }
  };
}

function reconnectWS() {
  reconnectAttempts = 0;
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  connectLogsWS();
}

const modalOpen = ref(false);
const modalAction = ref<"kick" | "ban" | null>(null);
const modalTarget = ref("");
const modalReason = ref("");

const modalTitle = computed(() =>
  modalAction.value === "kick" ? "Kick Player" : "Ban Player",
);

let fetchingOnline = false;

async function fetchOnlineWithRetry(attempt = 0) {
  if (fetchingOnline) return;
  fetchingOnline = true;
  onlineLoading.value = true;
  onlineError.value = null;
  try {
    const res = await $fetch<{
      online?: number;
      max?: number;
      players?: string[];
    }>(`/agent/${agentId.value}/servers/${serverId}/players`, {
      baseURL: apiBase,
      credentials: "include",
    });
    onlinePlayers.value = res.players || [];
    maxPlayers.value = res.max || 0;
  } catch (err: any) {
    const status = err?.response?.status;
    if (status === 503) {
      onlinePlayers.value = [];
      maxPlayers.value = 0;
      onlineError.value = "Server is offline";
    } else if (attempt < FETCH_MAX_ATTEMPTS) {
      const delay = Math.min(
        FETCH_RETRY_BASE_MS * 2 ** attempt,
        FETCH_RETRY_MAX_MS,
      );
      await new Promise((resolve) => setTimeout(resolve, delay));
      fetchingOnline = false;
      return fetchOnlineWithRetry(attempt + 1);
    } else {
      onlineError.value = err?.message || "Failed to load players";
    }
  } finally {
    onlineLoading.value = false;
    fetchingOnline = false;
  }
}

async function fetchOnline() {
  return fetchOnlineWithRetry();
}

let fetchingOps = false;

async function fetchOps() {
  if (fetchingOps) return;
  fetchingOps = true;
  opsLoading.value = true;
  opsError.value = null;
  try {
    const res = await $fetch<{ ops?: PlayerEntry[] }>(
      `/agent/${agentId.value}/servers/${serverId}/ops`,
      { baseURL: apiBase, credentials: "include" },
    );
    opsList.value = res.ops || [];
  } catch (err: any) {
    opsError.value = err?.message || "Failed to load operators";
  } finally {
    opsLoading.value = false;
    fetchingOps = false;
  }
}

let fetchingWl = false;

async function fetchWhitelist() {
  if (fetchingWl) return;
  fetchingWl = true;
  wlLoading.value = true;
  wlError.value = null;
  try {
    const res = await $fetch<{ whitelist?: PlayerEntry[] }>(
      `/agent/${agentId.value}/servers/${serverId}/whitelist`,
      { baseURL: apiBase, credentials: "include" },
    );
    wlList.value = res.whitelist || [];
  } catch (err: any) {
    wlError.value = err?.message || "Failed to load whitelist";
  } finally {
    wlLoading.value = false;
    fetchingWl = false;
  }
}

let fetchingBans = false;

async function fetchBans() {
  if (fetchingBans) return;
  fetchingBans = true;
  bansLoading.value = true;
  bansError.value = null;
  try {
    const res = await $fetch<{ bans?: PlayerEntry[] }>(
      `/agent/${agentId.value}/servers/${serverId}/bans`,
      { baseURL: apiBase, credentials: "include" },
    );
    bansList.value = res.bans || [];
  } catch (err: any) {
    bansError.value = err?.message || "Failed to load bans";
  } finally {
    bansLoading.value = false;
    fetchingBans = false;
  }
}

async function refreshAll() {
  refreshing.value = true;
  await Promise.all([fetchOnline(), fetchOps(), fetchWhitelist(), fetchBans()]);
  refreshing.value = false;
}

async function sendRcon(command: string) {
  try {
    await $fetch(`/agent/${agentId.value}/servers/${serverId}/rcon`, {
      baseURL: apiBase,
      method: "POST",
      credentials: "include",
      body: { command },
    });
    showToast("success", "Command sent", { description: command });
    setTimeout(() => refreshAll(), 500);
  } catch (err: any) {
    const msg = err?.data?.detail || err?.message || "Command failed";
    showToast("error", "RCON failed", { description: msg });
  }
}

function openReasonModal(action: "kick" | "ban", target: string) {
  modalAction.value = action;
  modalTarget.value = target;
  modalReason.value = "";
  modalOpen.value = true;
}

function closeModal() {
  modalOpen.value = false;
  modalAction.value = null;
  modalTarget.value = "";
  modalReason.value = "";
}

function confirmModal() {
  const action = modalAction.value;
  const target = modalTarget.value;
  const reason = modalReason.value.trim();
  if (!action || !target) {
    closeModal();
    return;
  }
  const cmd =
    action === "kick"
      ? reason
        ? `kick ${target} ${reason}`
        : `kick ${target}`
      : reason
        ? `ban ${target} ${reason}`
        : `ban ${target}`;
  closeModal();
  sendRcon(cmd);
}

onMounted(() => {
  loadAllPlayers();
  refreshAll();
  connectLogsWS();

  watch(lastEvent, (evt) => {
    if (!evt || evt.type !== "server.status") return;
    if (evt.agentId === agentId.value && evt.serverId === serverId) {
      refreshAll();
    }
  });
});

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
