<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
    @click.self="$emit('close')"
  >
    <div
      class="bg-white dark:bg-neutral-800 rounded-2xl shadow-xl w-full max-w-lg mx-4 overflow-hidden"
    >
      <div
        class="px-6 py-4 border-b border-gray-200 dark:border-neutral-700 flex items-center justify-between"
      >
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ editing ? "Edit Task" : "New Task" }}
        </h2>
        <button
          @click="$emit('close')"
          class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
        >
          <svg
            class="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <div class="px-6 py-4 space-y-4 max-h-[70vh] overflow-y-auto">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >Name</label
          >
          <input
            v-model="form.name"
            type="text"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="e.g. Daily backup"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >Trigger Type</label
          >
          <select
            v-model="form.trigger.type"
            :disabled="props.eventOnly"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
          >
            <option value="interval">Interval</option>
            <option value="cron">Cron</option>
            <option value="event">Event</option>
          </select>
        </div>

        <div v-if="form.trigger.type === 'interval'">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >Interval</label
          >
          <input
            v-model="form.trigger.interval"
            type="text"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="e.g. 5m, 1h, 30s"
          />
        </div>

        <div v-if="form.trigger.type === 'cron'">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >Cron Expression</label
          >
          <input
            v-model="form.trigger.cron"
            type="text"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="e.g. 0 4 * * *"
          />
        </div>

        <div v-if="form.trigger.type === 'event'">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >Event</label
          >
          <select
            v-model="form.trigger.event"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="server_start">Server Start</option>
            <option value="server_stop">Server Stop</option>
            <option value="server_restart">Server Restart</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >Action Type</label
          >
          <select
            v-model="form.action.type"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="rcon">RCON Command</option>
            <option value="container_exec">Container Exec</option>
            <option value="lifecycle">Lifecycle</option>
            <option value="backup">Backup</option>
          </select>
        </div>

        <div v-if="form.action.type === 'rcon' || form.action.type === 'container_exec'">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >Command</label
          >
          <input
            v-model="form.action.command"
            type="text"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="e.g. save-all"
          />
        </div>

        <div v-if="form.action.type === 'lifecycle'">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >Lifecycle Action</label
          >
          <select
            v-model="form.action.lifecycle"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="start">Start</option>
            <option value="stop">Stop</option>
            <option value="restart">Restart</option>
          </select>
        </div>

        <div v-if="form.action.type === 'backup'">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >Keep Last</label
          >
          <input
            v-model.number="form.action.keepLast"
            type="number"
            min="1"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="7"
          />
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mt-2 mb-1"
            >Keep Days</label
          >
          <input
            v-model.number="form.action.keepDays"
            type="number"
            min="1"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="30"
          />
        </div>

        <div class="flex items-center gap-2">
          <input
            id="enabled"
            v-model="form.enabled"
            type="checkbox"
            class="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          />
          <label for="enabled" class="text-sm text-gray-700 dark:text-gray-300"
            >Enabled</label
          >
        </div>
      </div>

      <div
        class="px-6 py-4 border-t border-gray-200 dark:border-neutral-700 flex justify-end gap-3"
      >
        <button
          @click="$emit('close')"
          class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-neutral-700 rounded-lg transition-colors"
        >
          Cancel
        </button>
        <button
          @click="save"
          class="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          {{ editing ? "Update" : "Create" }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Task } from "~/composables/useTasks";

const props = defineProps<{
  serverId: string;
  task: Task | null;
  eventOnly?: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "save", task: Task): void;
}>();

const editing = computed(() => !!props.task);

const form = reactive<Task>({
  id: props.task?.id || "",
  serverId: props.serverId,
  name: props.task?.name || "",
  enabled: props.task?.enabled ?? true,
  trigger: {
    type: props.eventOnly ? "event" : props.task?.trigger?.type || "interval",
    interval: props.task?.trigger?.interval || "",
    cron: props.task?.trigger?.cron || "",
    event: props.task?.trigger?.event || "server_start",
  },
  action: {
    type: props.task?.action?.type || "rcon",
    command: props.task?.action?.command || "",
    lifecycle: props.task?.action?.lifecycle || "",
    backupScope: props.task?.action?.backupScope || "world",
    keepLast: props.task?.action?.keepLast || 7,
    keepDays: props.task?.action?.keepDays || 30,
  },
  lastRun: props.task?.lastRun || undefined,
  nextRun: props.task?.nextRun || undefined,
});

function save() {
  emit("save", { ...form });
}
</script>
