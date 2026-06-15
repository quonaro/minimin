import {
  useServerEvents,
  type MetricsPayload,
} from "~/composables/useServerEvents";

export function useServerMetrics(serverId: string) {
  const { metricsMap } = useServerEvents();

  const serverMetrics = computed<MetricsPayload[]>(
    () => metricsMap.value[serverId] || [],
  );

  const latestMetric = computed<MetricsPayload | null>(() => {
    const arr = serverMetrics.value;
    return arr.length > 0 ? (arr[arr.length - 1] ?? null) : null;
  });

  function sparklinePoints(
    data: MetricsPayload[],
    key: "ramUsage" | "cpu" | "online" | "tps" | "rxRate" | "txRate",
  ): string {
    let values: number[];
    if (key === "online") {
      values = data.map((d) => (d.max > 0 ? (d.online / d.max) * 100 : 0));
    } else if (key === "tps") {
      values = data.map((d) => d.tps ?? 0);
    } else {
      values = data.map((d) => (d as any)[key] ?? 0);
    }
    if (values.length < 2) return "";
    const max = Math.max(...values, 0.001);
    const min = Math.min(...values);
    const range = max - min || 1;
    const w = 100;
    const h = 32;
    const step = w / (values.length - 1);
    return values
      .map((v, i) => {
        const x = i * step;
        const y = h - ((v - min) / range) * h;
        return `${x},${y}`;
      })
      .join(" ");
  }

  function formatBytes(bytes: number): string {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
  }

  function formatNetworkRate(kbps: number): string {
    if (kbps >= 1024) {
      return (kbps / 1024).toFixed(1) + " MB/s";
    }
    return kbps.toFixed(1) + " KB/s";
  }

  return {
    serverMetrics,
    latestMetric,
    sparklinePoints,
    formatBytes,
    formatNetworkRate,
  };
}
