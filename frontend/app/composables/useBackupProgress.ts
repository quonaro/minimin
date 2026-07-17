import { watch } from "vue";
import type { Backup } from "~/composables/useBackups";

export interface BackupTask {
  id: string;
  serverId: string;
  name?: string;
  status: "running" | "done" | "error";
  startedAt: number;
  finishedAt?: number;
  error?: string;
  elapsedMs: number;
}

const tasks = ref<BackupTask[]>([]);
let timer: ReturnType<typeof setInterval> | null = null;

function ensureTimer() {
  if (timer) return;
  timer = setInterval(() => {
    const now = Date.now();
    for (const t of tasks.value) {
      if (t.status === "running") {
        t.elapsedMs = now - t.startedAt;
      }
    }
  }, 200);
}

watch(
  tasks,
  (list) => {
    if (list.some((t) => t.status === "running")) {
      ensureTimer();
    } else if (timer) {
      clearInterval(timer);
      timer = null;
    }
  },
  { deep: true, immediate: true },
);

export function addBackupTask(
  serverId: string,
  runFn: () => Promise<Backup>,
) {
  const id = Math.random().toString(36).slice(2);
  const task: BackupTask = {
    id,
    serverId,
    status: "running",
    startedAt: Date.now(),
    elapsedMs: 0,
  };
  tasks.value.unshift(task);
  ensureTimer();

  runFn()
    .then((result) => {
      const t = tasks.value.find((x) => x.id === id);
      if (!t) return;
      t.status = "done";
      t.finishedAt = Date.now();
      t.elapsedMs = t.finishedAt - t.startedAt;
      if (result?.name) t.name = result.name;
    })
    .catch((err: any) => {
      const t = tasks.value.find((x) => x.id === id);
      if (!t) return;
      t.status = "error";
      t.finishedAt = Date.now();
      t.elapsedMs = t.finishedAt - t.startedAt;
      t.error = err?.message || "Failed";
    });

  return id;
}

export function removeBackupTask(id: string) {
  tasks.value = tasks.value.filter((t) => t.id !== id);
}

export function useBackupProgress() {
  return { tasks, removeBackupTask };
}
