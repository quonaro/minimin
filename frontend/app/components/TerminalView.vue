<template>
  <div class="flex flex-col min-w-0 h-[calc(100vh-4rem)]">
    <div class="flex-1 min-h-0 min-w-0 px-6 pt-6 pb-6 flex flex-col">
      <div class="mb-4 flex items-center justify-between flex-wrap gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            Terminal
          </h1>
          <p class="text-sm text-gray-500 dark:text-neutral-400 mt-1">
            Live server logs and RCON console.
          </p>
        </div>

        <div class="flex items-center gap-2">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-4 h-4"
            :class="{
              'text-green-500': rconStatus === 'Connected',
              'text-red-500':
                rconStatus === 'Error' || rconStatus === 'Disconnected',
              'text-gray-500 dark:text-neutral-400': !rconStatus,
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
              'text-green-500': rconStatus === 'Connected',
              'text-red-500':
                rconStatus === 'Error' || rconStatus === 'Disconnected',
              'text-gray-500 dark:text-neutral-400': !rconStatus,
            }"
          >
            {{ rconStatus || "Connecting..." }}
          </span>
          <button
            class="px-3 py-2 rounded-lg bg-primary text-white text-sm hover:bg-primary/90"
            @click="reconnectRcon"
          >
            Reconnect
          </button>
        </div>
      </div>

      <server-logs
        ref="serverLogsRef"
        :server-id="serverId"
        :filter="activeFilter"
        show-filters
        hide-status
        class="flex-1 min-h-0 min-w-0"
        @update:filter="activeFilter = $event"
      />

      <terminal-input
        ref="terminalInputRef"
        :disabled="rconStatus !== 'Connected'"
        :players="players"
        @send="onSendCommand"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { useRconConsole } from "~/composables/useRconConsole";
import type ServerLogs from "~/components/ServerLogs.vue";
import type TerminalInput from "~/components/TerminalInput.vue";

const props = defineProps<{
  serverId: string;
}>();

const serverLogsRef = ref<InstanceType<typeof ServerLogs> | null>(null);
const terminalInputRef = ref<InstanceType<typeof TerminalInput> | null>(null);

const { playersMap } = useServerEvents();
const players = computed<string[]>(
  () => playersMap.value[props.serverId]?.players ?? [],
);

const activeFilter = ref("all");

const {
  wsStatus: rconStatus,
  messages: rconMessages,
  connect: connectRcon,
  sendCommand,
  reconnect: reconnectRcon,
  clearMessages: clearRconMessages,
} = useRconConsole(props.serverId);

function onSendCommand(cmd: string) {
  sendCommand(cmd);
  serverLogsRef.value?.addExternalLine(`> ${cmd}`, "command");
}

const processedCount = ref(0);

watch(
  () => rconMessages.value.length,
  (newLen) => {
    const added = rconMessages.value.slice(processedCount.value, newLen);
    for (const msg of added) {
      if (msg.type === "command") continue; // already added inline
      serverLogsRef.value?.addExternalLine(msg.text, msg.type);
    }
    processedCount.value = newLen;
  },
);

onMounted(() => {
  connectRcon();
  terminalInputRef.value?.focus();
});
</script>
