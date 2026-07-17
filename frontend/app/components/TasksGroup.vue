<template>
  <div
    class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden"
  >
    <div
      class="w-full flex items-center justify-between p-4 md:p-6 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
    >
      <button
        @click="expanded = !expanded"
        class="flex items-center gap-3 flex-1 text-left"
      >
        <Clock class="w-5 h-5 text-blue-500" />
        <h2 class="text-xl font-bold text-gray-900 dark:text-white">Tasks</h2>
      </button>
      <div class="flex items-center gap-3">
        <button
          @click.stop="showCreate = true"
          class="px-3 py-1.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
        >
          New Task
        </button>
        <button
          @click="expanded = !expanded"
          class="p-1 text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-200 transition-colors"
        >
          <ChevronDown
            class="w-5 h-5 transition-transform"
            :class="{ 'rotate-180': expanded }"
          />
        </button>
      </div>
    </div>

    <div
      v-show="expanded"
      class="px-4 pt-4 pb-4 md:px-6 md:pt-6 md:pb-6 space-y-4"
    >
      <div
        v-if="loading"
        class="text-center py-12 text-gray-500 dark:text-neutral-400"
      >
        Loading tasks...
      </div>
      <div
        v-else-if="error"
        class="text-center py-12 text-red-500 dark:text-red-400"
      >
        {{ error }}
      </div>
      <div
        v-else-if="tasks.length === 0"
        class="text-center py-12 text-gray-500 dark:text-neutral-400"
      >
        No scheduled tasks yet.
      </div>

      <div
        v-else
        class="max-h-[32rem] overflow-y-auto pr-1 space-y-3 no-scrollbar"
      >
        <div
          v-for="task in tasks"
          :key="task.id"
          class="bg-gray-50 dark:bg-neutral-700/50 border border-gray-200 dark:border-neutral-700 rounded-xl p-4 flex items-center gap-4"
        >
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <h3 class="font-semibold text-gray-900 dark:text-white truncate">
                {{ task.name }}
              </h3>
              <span
                class="text-xs px-2 py-0.5 rounded-full font-medium"
                :class="
                  task.enabled
                    ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
                    : 'bg-gray-100 text-gray-600 dark:bg-neutral-700 dark:text-neutral-400'
                "
              >
                {{ task.enabled ? "Enabled" : "Disabled" }}
              </span>
            </div>
            <p class="text-sm text-gray-500 dark:text-neutral-400 mt-1">
              <template v-if="task.trigger.type === 'event'">
                Event:
                <span class="font-mono text-xs">{{ task.trigger.event }}</span>
              </template>
              <template v-else>
                <span class="capitalize">{{ task.trigger.type }}</span>
                <template v-if="task.trigger.interval">
                  · every {{ task.trigger.interval }}
                </template>
                <template v-if="task.trigger.cron">
                  · {{ task.trigger.cron }}
                </template>
              </template>
              ·
              <span class="capitalize">{{ task.action.type }}</span>
            </p>
            <p
              v-if="task.lastRun"
              class="text-xs text-gray-400 dark:text-neutral-500 mt-1"
            >
              Last run: {{ formatDate(task.lastRun) }}
            </p>
          </div>
          <div class="flex items-center gap-2 flex-shrink-0">
            <button
              @click="toggleTask(task)"
              class="h-9 px-3 py-1.5 text-sm rounded-lg text-white shadow-sm transition-colors"
              :class="
                task.enabled
                  ? 'bg-amber-600 hover:bg-amber-700'
                  : 'bg-green-600 hover:bg-green-700'
              "
            >
              {{ task.enabled ? "Disable" : "Enable" }}
            </button>
            <button
              v-if="task.trigger.type !== 'event'"
              @click="runTask(task.id)"
              class="h-9 px-3 py-1.5 text-sm rounded-lg bg-emerald-600 text-white hover:bg-emerald-700 shadow-sm transition-colors"
            >
              Run Now
            </button>
            <button
              @click="editTask(task)"
              class="h-9 px-3 py-1.5 text-sm rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 shadow-sm transition-colors"
            >
              Edit
            </button>
            <button
              @click="deleteTask(task.id)"
              class="h-9 px-3 py-1.5 text-sm rounded-lg bg-red-600 text-white hover:bg-red-700 shadow-sm transition-colors"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>

    <TaskModal
      v-if="showCreate || editingTask"
      :server-id="serverId"
      :task="editingTask"
      @close="closeModal"
      @save="onSave"
    />
  </div>
</template>

<script setup lang="ts">
import { ChevronDown, Clock } from "lucide-vue-next";
import type { Task } from "~/composables/useTasks";

const props = defineProps<{
  serverId: string;
}>();

const {
  tasks,
  loading,
  error,
  fetchTasks,
  createTask,
  updateTask: updateTaskApi,
  deleteTask: deleteTaskApi,
  runTask: runTaskApi,
} = useTasks(props.serverId);

const expanded = ref(true);
const showCreate = ref(false);
const editingTask = ref<Task | null>(null);

onMounted(() => {
  fetchTasks();
});

function toggleTask(task: Task) {
  updateTaskApi({ ...task, enabled: !task.enabled });
}

function runTask(id: string) {
  runTaskApi(id);
}

function deleteTask(id: string) {
  deleteTaskApi(id);
}

function editTask(task: Task) {
  editingTask.value = { ...task };
}

function closeModal() {
  showCreate.value = false;
  editingTask.value = null;
}

async function onSave(task: Task) {
  if (editingTask.value) {
    await updateTaskApi(task);
  } else {
    await createTask(task);
  }
  closeModal();
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString();
}
</script>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
