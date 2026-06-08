<template>
  <div
    class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-5 hover:shadow-md transition-shadow"
  >
    <div class="flex items-start justify-between mb-4">
      <div class="min-w-0">
        <h3
          class="text-lg font-semibold text-gray-900 dark:text-white truncate"
        >
          {{ server.serverId }}
        </h3>
      </div>
      <span
        :class="statusClasses"
        class="px-2.5 py-1 rounded-full text-xs font-medium flex items-center gap-1.5 shrink-0"
      >
        <component :is="statusIcon" class="w-3.5 h-3.5" />
        {{ server.status }}
      </span>
    </div>

    <div class="space-y-2.5 mb-5">
      <div
        v-if="uptime"
        class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400"
      >
        <Clock class="w-4 h-4 shrink-0 text-gray-400 dark:text-gray-500" />
        <span>Running for {{ uptime }}</span>
      </div>

      <div
        class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400"
      >
        <Network class="w-4 h-4 shrink-0 text-gray-400 dark:text-gray-500" />
        <span>Port {{ server.gamePort || "-" }}</span>
      </div>

      <div
        class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400"
      >
        <Cpu class="w-4 h-4 shrink-0 text-gray-400 dark:text-gray-500" />
        <span>{{ server.engineType }}</span>
      </div>

      <div
        class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400"
      >
        <Tag class="w-4 h-4 shrink-0 text-gray-400 dark:text-gray-500" />
        <span>{{ server.gameVersion }}</span>
      </div>
    </div>

    <NuxtLink
      :to="`/agent/${agentId}/servers/${server.serverId}`"
      class="inline-flex items-center gap-1.5 text-primary hover:text-primary/80 font-medium text-sm"
    >
      View Details
      <ChevronRight class="w-4 h-4" />
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { formatDuration } from "~/composables/useDuration";
import {
  Activity,
  Clock,
  Cpu,
  Network,
  Tag,
  ChevronRight,
} from "lucide-vue-next";

interface Server {
  serverId: string;
  status: string;
  gamePort: number;
  engineType: string;
  gameVersion: string;
  startedAt?: string;
}

const props = defineProps<{
  server: Server;
  agentId: string;
}>();

const statusClasses = computed(() => {
  switch (props.server.status) {
    case "running":
      return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
    case "exited":
      return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
    default:
      return "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300";
  }
});

const statusIcon = computed(() => {
  return props.server.status === "running" ? Activity : null;
});

const uptime = computed(() => {
  if (props.server.status !== "running" || !props.server.startedAt) {
    return "";
  }
  const start = new Date(props.server.startedAt).getTime();
  return formatDuration(Date.now() - start);
});
</script>
