export type ToastType = "success" | "error" | "warning" | "info";

export interface ToastItem {
  id: string;
  type: ToastType;
  title: string;
  description?: string;
  createdAt: number;
}

function useToastsState() {
  return useState<ToastItem[]>("toasts", () => []);
}

export function useToastStack() {
  const toasts = useToastsState();

  function push(
    type: ToastType,
    title: string,
    description?: string,
    duration = 4000,
  ) {
    const id = Math.random().toString(36).slice(2);
    const item: ToastItem = {
      id,
      type,
      title,
      description,
      createdAt: Date.now(),
    };
    toasts.value.push(item);
    setTimeout(() => remove(id), duration);
  }

  function remove(id: string) {
    const idx = toasts.value.findIndex((t) => t.id === id);
    if (idx !== -1) {
      toasts.value.splice(idx, 1);
    }
  }

  return {
    toasts,
    push,
    remove,
  };
}
