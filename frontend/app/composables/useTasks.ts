export interface Trigger {
  type: "interval" | "cron" | "event";
  interval?: string;
  cron?: string;
  event?: string;
}

export interface Action {
  type: "rcon" | "container_exec" | "lifecycle" | "backup";
  command?: string;
  lifecycle?: string;
  backupScope?: string;
  keepLast?: number;
  keepDays?: number;
}

export interface Task {
  id: string;
  serverId: string;
  name: string;
  enabled: boolean;
  trigger: Trigger;
  action: Action;
  lastRun?: string;
  nextRun?: string;
}

export function useTasks(serverId: string) {
  const tasks = ref<Task[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const { show } = useToast();

  async function fetchTasks() {
    loading.value = true;
    error.value = null;
    try {
      const res = await $fetch<Task[]>(`/servers/${serverId}/tasks`, {
        baseURL: useApiBase(),
        credentials: "include",
      });
      tasks.value = res || [];
    } catch (e: any) {
      error.value = getApiErrorMessage(e, "Failed to load tasks");
    } finally {
      loading.value = false;
    }
  }

  async function createTask(task: Omit<Task, "id" | "serverId">) {
    try {
      await $fetch(`/servers/${serverId}/tasks`, {
        baseURL: useApiBase(),
        method: "POST",
        credentials: "include",
        body: task,
      });
      show("success", "Task created");
      await fetchTasks();
    } catch (e: any) {
      show("error", getApiErrorMessage(e, "Failed to create task"));
    }
  }

  async function updateTask(task: Task) {
    try {
      await $fetch(`/servers/${serverId}/tasks/${task.id}`, {
        baseURL: useApiBase(),
        method: "PUT",
        credentials: "include",
        body: task,
      });
      show("success", "Task updated");
      await fetchTasks();
    } catch (e: any) {
      show("error", getApiErrorMessage(e, "Failed to update task"));
    }
  }

  async function deleteTask(taskId: string) {
    try {
      await $fetch(`/servers/${serverId}/tasks/${taskId}`, {
        baseURL: useApiBase(),
        method: "DELETE",
        credentials: "include",
      });
      show("success", "Task deleted");
      await fetchTasks();
    } catch (e: any) {
      show("error", getApiErrorMessage(e, "Failed to delete task"));
    }
  }

  async function runTask(taskId: string) {
    try {
      await $fetch(`/servers/${serverId}/tasks/${taskId}/run`, {
        baseURL: useApiBase(),
        method: "POST",
        credentials: "include",
      });
      show("info", "Task execution requested");
    } catch (e: any) {
      show("error", getApiErrorMessage(e, "Failed to run task"));
    }
  }

  return {
    tasks,
    loading,
    error,
    fetchTasks,
    createTask,
    updateTask,
    deleteTask,
    runTask,
  };
}
