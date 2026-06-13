import { reactive, computed } from "vue";

const refreshCounters = reactive<Record<string, number>>({});

export function useClientAssetsRefresh(serverId: string) {
  const key = computed(() => refreshCounters[serverId] || 0);

  function trigger() {
    refreshCounters[serverId] = (refreshCounters[serverId] || 0) + 1;
  }

  return { key, trigger };
}
