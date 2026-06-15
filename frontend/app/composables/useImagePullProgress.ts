export interface ImagePullProgress {
  current: number;
  total: number;
}

export function useImagePullProgress(serverId: Ref<string>, serverStatus: Ref<string>) {
  const progress = ref<ImagePullProgress | null>(null);

  let interval: ReturnType<typeof setInterval> | null = null;

  async function fetchProgress() {
    if (!serverId.value) return;
    try {
      const res = await $fetch<ImagePullProgress | null>(
        `/servers/${serverId.value}/pull-progress`,
        {
          baseURL: useApiBase(),
          credentials: "include",
        },
      );
      progress.value = res;
    } catch {
      progress.value = null;
    }
  }

  watch(
    () => serverStatus.value,
    (status) => {
      if (status === "pulling_image") {
        fetchProgress();
        if (!interval) {
          interval = setInterval(fetchProgress, 500);
        }
      } else {
        progress.value = null;
        if (interval) {
          clearInterval(interval);
          interval = null;
        }
      }
    },
    { immediate: true },
  );

  onUnmounted(() => {
    if (interval) {
      clearInterval(interval);
    }
  });

  const percentage = computed(() => {
    const p = progress.value;
    if (!p || p.total <= 0) return 0;
    return Math.min(100, Math.round((p.current / p.total) * 100));
  });

  return { progress, percentage };
}
