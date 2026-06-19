import { ref, computed, watchEffect, toValue, type MaybeRefOrGetter } from "vue";

export interface ServerDisk {
  totalBytes: number;
  worldBytes: number;
  worldNetherBytes: number;
  worldEndBytes: number;
}

interface DiskState {
  disk: ServerDisk | null;
  loading: boolean;
  error: string | null;
  fetched: boolean;
}

const diskMap = ref<Record<string, DiskState>>({});

async function fetchDisk(serverId: string) {
  if (diskMap.value[serverId]?.loading) return;

  if (!diskMap.value[serverId]) {
    diskMap.value[serverId] = {
      disk: null,
      loading: true,
      error: null,
      fetched: false,
    };
  } else {
    diskMap.value[serverId].loading = true;
    diskMap.value[serverId].error = null;
  }

  try {
    const res = await $fetch<
      { totalBytes: number; worldBytes: number; worldNetherBytes: number; worldEndBytes: number } |
      { body: { totalBytes: number; worldBytes: number; worldNetherBytes: number; worldEndBytes: number } }
    >(`/servers/${serverId}/disk`, {
      baseURL: useApiBase(),
      credentials: "include",
    });

    const body = (res as any).body ?? (res as any);

    const disk: ServerDisk = {
      totalBytes: body.totalBytes ?? 0,
      worldBytes: body.worldBytes ?? 0,
      worldNetherBytes: body.worldNetherBytes ?? 0,
      worldEndBytes: body.worldEndBytes ?? 0,
    };

    diskMap.value[serverId] = {
      disk,
      loading: false,
      error: null,
      fetched: true,
    };
  } catch (err: any) {
    diskMap.value[serverId] = {
      disk: null,
      loading: false,
      error: getApiErrorMessage(err, "Failed to load disk usage"),
      fetched: true,
    };
  }
}

export function useServerDisk(serverId: MaybeRefOrGetter<string | undefined | null>) {
  const id = computed(() => toValue(serverId) ?? "");

  watchEffect(() => {
    if (id.value && !diskMap.value[id.value]?.fetched) {
      fetchDisk(id.value);
    }
  });

  const disk = computed<ServerDisk | null>(() => {
    if (!id.value) return null;
    return diskMap.value[id.value]?.disk ?? null;
  });

  const loading = computed<boolean>(() => {
    if (!id.value) return false;
    return diskMap.value[id.value]?.loading ?? false;
  });

  const error = computed<string | null>(() => {
    if (!id.value) return null;
    return diskMap.value[id.value]?.error ?? null;
  });

  async function refresh() {
    if (!id.value) return;
    await fetchDisk(id.value);
  }

  return {
    disk,
    loading,
    error,
    refresh,
  };
}
