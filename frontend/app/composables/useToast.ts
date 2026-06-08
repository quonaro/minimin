export interface ToastOptions {
  description?: string;
  duration?: number;
}

export function useToast() {
  const { push } = useToastStack();

  const show = (
    type: "success" | "error" | "warning" | "info",
    title: string,
    opts?: ToastOptions,
  ) => {
    push(type, title, opts?.description, opts?.duration ?? 4000);
  };

  return { show };
}
