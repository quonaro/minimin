<template>
  <div
    class="bg-white dark:bg-neutral-800 rounded-xl shadow-sm border border-gray-200 dark:border-neutral-700 p-5 hover:shadow-md transition-shadow h-full flex flex-col"
  >
    <div class="flex items-start justify-between mb-4">
      <div class="min-w-0">
        <h3
          class="text-lg font-semibold text-gray-900 dark:text-white truncate"
        >
          {{ server.serverId }}
        </h3>
      </div>
      <div class="flex items-center gap-2 shrink-0 flex-wrap justify-end">
        <span
          :class="[
            getStatusColor(server.containerStatus),
            server.containerStatus === 'running' &&
              'animate-heartbeat dark:animate-heartbeat-dark',
          ]"
          class="px-2 py-0.5 rounded-full text-[10px] font-medium flex items-center gap-1"
        >
          <Activity
            v-if="server.containerStatus === 'running'"
            class="w-3 h-3"
          />
          container: {{ server.containerStatus }}
        </span>
        <span
          :class="[
            getStatusColor(server.serverStatus),
            server.serverStatus === 'running' &&
              'animate-heartbeat dark:animate-heartbeat-dark',
          ]"
          class="px-2 py-0.5 rounded-full text-[10px] font-medium flex items-center gap-1"
        >
          <Activity v-if="server.serverStatus === 'running'" class="w-3 h-3" />
          server: {{ server.serverStatus }}
        </span>
        <span
          v-if="server.modCount !== undefined && server.modCount > 0"
          class="px-2 py-0.5 rounded-full text-[10px] font-medium flex items-center gap-1 bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400"
        >
          <Package class="w-3 h-3" />
          {{ server.modCount }} mod{{ server.modCount === 1 ? "" : "s" }}
        </span>
      </div>
    </div>

    <div class="space-y-2.5 mb-5">
      <div
        class="flex items-center gap-2 text-sm text-gray-600 dark:text-neutral-400 min-h-[1.25rem]"
      >
        <template v-if="containerUptime">
          <Clock class="w-4 h-4 shrink-0 text-gray-400 dark:text-neutral-500" />
          <span>Container: {{ containerUptime }}</span>
        </template>
      </div>
      <div
        class="flex items-center gap-2 text-sm text-gray-600 dark:text-neutral-400 min-h-[1.25rem]"
      >
        <template v-if="serverUptime">
          <Clock class="w-4 h-4 shrink-0 text-gray-400 dark:text-neutral-500" />
          <span>Server: {{ serverUptime }}</span>
        </template>
      </div>

      <div
        class="flex items-center gap-2 text-sm text-gray-600 dark:text-neutral-400"
      >
        <Network class="w-4 h-4 shrink-0 text-gray-400 dark:text-neutral-500" />
        <span>Port {{ server.gamePort || "-" }}</span>
      </div>

      <div
        class="flex items-center gap-2 text-sm text-gray-600 dark:text-neutral-400"
      >
        <div
          :class="[
            'w-6 h-6 shrink-0 rounded flex items-center justify-center text-[10px] font-bold',
            getEngineIconColor(server.engineType),
          ]"
        >
          {{ getEngineAbbreviation(server.engineType) }}
        </div>
        <span>{{ server.engineType }}</span>
      </div>

      <div
        class="flex items-center gap-2 text-sm text-gray-600 dark:text-neutral-400"
      >
        <Tag class="w-4 h-4 shrink-0 text-gray-400 dark:text-neutral-500" />
        <span>{{ server.gameVersion }}</span>
      </div>
    </div>

    <NuxtLink
      :to="`/servers/${server.serverId}`"
      class="inline-flex items-center gap-1.5 text-primary hover:text-primary/80 font-medium text-sm mt-auto"
    >
      View Details
      <ChevronRight class="w-4 h-4" />
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useUptime } from "~/composables/useDuration";
import type { Server } from "~/composables/useServers";
import {
  Activity,
  Clock,
  Network,
  Package,
  Tag,
  ChevronRight,
} from "lucide-vue-next";

const props = defineProps<{
  server: Server;
}>();

function getStatusColor(status: string) {
  switch (status) {
    case "running":
      return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
    case "starting":
      return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
    case "exited":
      return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
    default:
      return "bg-gray-100 text-gray-800 dark:bg-neutral-700 dark:text-neutral-300";
  }
}

function getEngineAbbreviation(engineType: string) {
  switch (engineType.toUpperCase()) {
    case "FORGE":
      return "FO";
    case "FABRIC":
      return "FA";
    case "PAPERMC":
      return "PA";
    case "VANILLA":
      return "MC";
    default:
      return engineType.substring(0, 2).toUpperCase();
  }
}

function getEngineIconColor(engineType: string) {
  switch (engineType.toUpperCase()) {
    case "FORGE":
      return "bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400";
    case "FABRIC":
      return "bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400";
    case "PAPERMC":
      return "bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400";
    case "VANILLA":
      return "bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400";
    default:
      return "bg-gray-100 text-gray-600 dark:bg-neutral-700 dark:text-neutral-400";
  }
}

const containerUptime = useUptime(
  computed(() =>
    props.server.containerStatus === "running"
      ? props.server.containerStartedAt
      : undefined,
  ),
);

const serverUptime = useUptime(
  computed(() =>
    props.server.serverStatus === "running"
      ? props.server.serverStartedAt
      : undefined,
  ),
);
</script>
