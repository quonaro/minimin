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
        <Zap class="w-5 h-5 text-amber-500" />
        <h2 class="text-xl font-bold text-gray-900 dark:text-white">Events</h2>
      </button>
      <div class="flex items-center gap-3">
        <button
          @click.stop="showCreate = true"
          class="px-3 py-1.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
        >
          New Event Hook
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
        Loading event hooks...
      </div>
      <div
        v-else-if="error"
        class="text-center py-12 text-red-500 dark:text-red-400"
      >
        {{ error }}
      </div>
      <div
        v-else-if="eventTasks.length === 0"
        class="text-center py-12 text-gray-500 dark:text-neutral-400"
      >
        No event hooks yet.
      </div>

      <div
        v-else
        class="max-h-[32rem] overflow-y-auto pr-1 space-y-3 no-scrollbar"
      >
        <div
          v-for="task in eventTasks"
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
              Event:
              <span class="font-mono text-xs">{{
                task.trigger.event || "—"
              }}</span>
              ·
              <span class="capitalize">{{ task.action.type }}</span>
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
      event-only
      @close="closeModal"
      @save="onSave"
    />
  </div>
</template>

<script setup lang="ts">
import { ChevronDown, Zap } from "lucide-vue-next";
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
} = useTasks(props.serverId);

const expanded = ref(true);
const showCreate = ref(false);
const editingTask = ref<Task | null>(null);

const eventTasks = computed(() =>
  tasks.value.filter((t) => t.trigger.type === "event"),
);

onMounted(() => {
  fetchTasks();
});

function toggleTask(task: Task) {
  updateTaskApi({ ...task, enabled: !task.enabled });
}

function editTask(task: Task) {
  editingTask.value = { ...task };
}

function deleteTask(id: string) {
  deleteTaskApi(id);
}

function closeModal() {
  showCreate.value = false;
  editingTask.value = null;
}

async function onSave(task: Task) {
  try {
    if (editingTask.value) {
      await updateTaskApi(task);
    } else {
      await createTask(task);
    }
    closeModal();
  } catch {
    // error already toasted by useTasks; keep modal open
  }
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
