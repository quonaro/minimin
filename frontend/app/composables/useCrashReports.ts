export interface CrashReport {
  name: string;
  size: number;
  modifiedAt: string;
}

export function useCrashReports(serverId: string) {
  const crashReports = ref<CrashReport[]>([]);
  const crashReportsLoading = ref(false);
  const crashReportsFetched = ref(false);

  const hasCrashReports = computed(() => crashReports.value.length > 0);

  const latestCrashReportDate = computed(() => {
    if (!crashReports.value.length) return null;
    return crashReports.value[0]?.modifiedAt ?? null;
  });

  async function fetchCrashReports() {
    if (crashReportsFetched.value) return;
    crashReportsLoading.value = true;
    try {
      const res = await $fetch<
        | { reports?: CrashReport[] }
        | { body?: { reports?: CrashReport[] } }
      >(`/servers/${serverId}/crash-reports`, {
        baseURL: useApiBase(),
        credentials: "include",
      });
      const body = (res as any).body ?? (res as any);
      crashReports.value = body?.reports ?? [];
      crashReportsFetched.value = true;
    } catch {
      crashReports.value = [];
    } finally {
      crashReportsLoading.value = false;
    }
  }

  onMounted(() => {
    fetchCrashReports();
  });

  return {
    crashReports,
    crashReportsLoading,
    crashReportsFetched,
    hasCrashReports,
    latestCrashReportDate,
    fetchCrashReports,
  };
}
