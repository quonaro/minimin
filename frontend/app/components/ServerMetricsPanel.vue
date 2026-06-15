<script setup lang="ts">
import {
  Activity,
  ArrowDown,
  ArrowUp,
  MemoryStick,
  Users,
  Zap,
} from "lucide-vue-next";

const props = defineProps<{
  serverId: string;
  containerStatus?: string;
}>();

const { serverMetrics, latestMetric, sparklinePoints, formatBytes } =
  useServerMetrics(props.serverId);
</script>

<template>
  <div class="space-y-2">
    <p
      class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
    >
      Live Metrics
    </p>
    <div class="grid grid-cols-2 sm:grid-cols-2 gap-3">
      <!-- RAM -->
      <div
        class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
      >
        <div
          class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400"
          >
            <MemoryStick class="w-4 h-4" />
          </div>
          <div class="min-w-0">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              RAM
            </p>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{
                latestMetric
                  ? formatBytes(latestMetric.ramUsage) +
                    " \u00B7 " +
                    formatBytes(latestMetric.ramLimit)
                  : "—"
              }}
            </p>
          </div>
        </div>
        <template v-if="containerStatus === 'running'">
          <svg
            class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
            viewBox="0 0 100 32"
            preserveAspectRatio="none"
          >
            <polyline
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              class="text-indigo-500"
              :points="sparklinePoints(serverMetrics, 'ramUsage')"
            />
          </svg>
        </template>
        <template v-else>
          <div
            class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
          >
            <span class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
              >Server is down</span
            >
          </div>
        </template>
      </div>

      <!-- CPU -->
      <div
        class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
      >
        <div
          class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-green-100 dark:bg-green-900/30 flex items-center justify-center text-green-600 dark:text-green-400"
          >
            <Activity class="w-4 h-4" />
          </div>
          <div class="min-w-0">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              CPU
            </p>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ latestMetric ? latestMetric.cpu.toFixed(1) + "%" : "—" }}
            </p>
          </div>
        </div>
        <template v-if="containerStatus === 'running'">
          <svg
            class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
            viewBox="0 0 100 32"
            preserveAspectRatio="none"
          >
            <polyline
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              class="text-emerald-500"
              :points="sparklinePoints(serverMetrics, 'cpu')"
            />
          </svg>
        </template>
        <template v-else>
          <div
            class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
          >
            <span class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
              >Server is down</span
            >
          </div>
        </template>
      </div>

      <!-- Online -->
      <div
        class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
      >
        <div
          class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400"
          >
            <Users class="w-4 h-4" />
          </div>
          <div class="min-w-0">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              Online
            </p>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{
                latestMetric
                  ? latestMetric.online + "/" + latestMetric.max
                  : "—"
              }}
            </p>
          </div>
        </div>
        <template v-if="containerStatus === 'running'">
          <svg
            class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
            viewBox="0 0 100 32"
            preserveAspectRatio="none"
          >
            <polyline
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              class="text-sky-500"
              :points="sparklinePoints(serverMetrics, 'online')"
            />
          </svg>
        </template>
        <template v-else>
          <div
            class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
          >
            <span class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
              >Server is down</span
            >
          </div>
        </template>
      </div>

      <!-- TPS -->
      <div
        class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
      >
        <div
          class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center text-purple-600 dark:text-purple-400"
          >
            <Zap class="w-4 h-4" />
          </div>
          <div class="min-w-0">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              TPS
            </p>
            <p
              class="text-sm font-semibold"
              :class="
                latestMetric?.tps == null
                  ? 'text-gray-900 dark:text-white'
                  : latestMetric.tps >= 18
                    ? 'text-green-600 dark:text-green-400'
                    : latestMetric.tps >= 15
                      ? 'text-yellow-600 dark:text-yellow-400'
                      : 'text-red-600 dark:text-red-400'
              "
            >
              {{ latestMetric?.tps != null ? latestMetric.tps.toFixed(1) : "—" }}
            </p>
          </div>
        </div>
        <template v-if="containerStatus === 'running'">
          <svg
            class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
            viewBox="0 0 100 32"
            preserveAspectRatio="none"
          >
            <polyline
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              class="text-purple-500"
              :points="sparklinePoints(serverMetrics, 'tps')"
            />
          </svg>
        </template>
        <template v-else>
          <div
            class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
          >
            <span class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
              >Server is down</span
            >
          </div>
        </template>
      </div>

      <!-- Net RX -->
      <div
        class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
      >
        <div
          class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-cyan-100 dark:bg-cyan-900/30 flex items-center justify-center text-cyan-600 dark:text-cyan-400"
          >
            <ArrowDown class="w-4 h-4" />
          </div>
          <div class="min-w-0">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              Net RX
            </p>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ latestMetric ? latestMetric.rxRate.toFixed(1) + " KB/s" : "—" }}
            </p>
          </div>
        </div>
        <template v-if="containerStatus === 'running'">
          <svg
            class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
            viewBox="0 0 100 32"
            preserveAspectRatio="none"
          >
            <polyline
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              class="text-cyan-500"
              :points="sparklinePoints(serverMetrics, 'rxRate')"
            />
          </svg>
        </template>
        <template v-else>
          <div
            class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
          >
            <span class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
              >Server is down</span
            >
          </div>
        </template>
      </div>

      <!-- Net TX -->
      <div
        class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
      >
        <div
          class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
        >
          <div
            class="w-9 h-9 shrink-0 rounded-lg bg-teal-100 dark:bg-teal-900/30 flex items-center justify-center text-teal-600 dark:text-teal-400"
          >
            <ArrowUp class="w-4 h-4" />
          </div>
          <div class="min-w-0">
            <p
              class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
            >
              Net TX
            </p>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ latestMetric ? latestMetric.txRate.toFixed(1) + " KB/s" : "—" }}
            </p>
          </div>
        </div>
        <template v-if="containerStatus === 'running'">
          <svg
            class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
            viewBox="0 0 100 32"
            preserveAspectRatio="none"
          >
            <polyline
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              class="text-teal-500"
              :points="sparklinePoints(serverMetrics, 'txRate')"
            />
          </svg>
        </template>
        <template v-else>
          <div
            class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
          >
            <span class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
              >Server is down</span
            >
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
