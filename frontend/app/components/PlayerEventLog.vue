<script setup lang="ts">
import {
  LogIn,
  LogOut,
  Shield,
  ShieldOff,
  Ban,
  CircleCheck,
  FilePlus,
  FileMinus,
  Swords,
} from "lucide-vue-next";

export interface PlayerEvent {
  id: string;
  ts: number;
  type:
    | "join"
    | "leave"
    | "kick"
    | "ban"
    | "unban"
    | "op"
    | "deop"
    | "wladd"
    | "wlremove";
  player: string;
  reason?: string;
}

const props = defineProps<{
  events: PlayerEvent[];
}>();

const typeMeta: Record<
  PlayerEvent["type"],
  { label: string; icon: Component; color: string; bg: string }
> = {
  join: {
    label: "joined",
    icon: LogIn,
    color: "text-green-600 dark:text-green-400",
    bg: "bg-green-100 dark:bg-green-900/30",
  },
  leave: {
    label: "left",
    icon: LogOut,
    color: "text-gray-600 dark:text-neutral-400",
    bg: "bg-gray-100 dark:bg-neutral-700/50",
  },
  kick: {
    label: "kicked",
    icon: Swords,
    color: "text-orange-600 dark:text-orange-400",
    bg: "bg-orange-100 dark:bg-orange-900/30",
  },
  ban: {
    label: "banned",
    icon: Ban,
    color: "text-red-600 dark:text-red-400",
    bg: "bg-red-100 dark:bg-red-900/30",
  },
  unban: {
    label: "unbanned",
    icon: CircleCheck,
    color: "text-green-600 dark:text-green-400",
    bg: "bg-green-100 dark:bg-green-900/30",
  },
  op: {
    label: "opped",
    icon: Shield,
    color: "text-blue-600 dark:text-blue-400",
    bg: "bg-blue-100 dark:bg-blue-900/30",
  },
  deop: {
    label: "deopped",
    icon: ShieldOff,
    color: "text-gray-600 dark:text-neutral-400",
    bg: "bg-gray-100 dark:bg-neutral-700/50",
  },
  wladd: {
    label: "whitelisted",
    icon: FilePlus,
    color: "text-blue-600 dark:text-blue-400",
    bg: "bg-blue-100 dark:bg-blue-900/30",
  },
  wlremove: {
    label: "removed from whitelist",
    icon: FileMinus,
    color: "text-gray-600 dark:text-neutral-400",
    bg: "bg-gray-100 dark:bg-neutral-700/50",
  },
};

function formatTime(ts: number): string {
  const d = new Date(ts);
  return d.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}
</script>

<template>
  <div class="space-y-1 max-h-80 overflow-y-auto pr-1">
    <div
      v-for="ev in events"
      :key="ev.id"
      class="flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-gray-50 dark:hover:bg-neutral-700/30 transition-colors"
    >
      <div
        class="w-6 h-6 rounded-full flex items-center justify-center shrink-0"
        :class="typeMeta[ev.type].bg"
      >
        <component
          :is="typeMeta[ev.type].icon"
          class="w-3.5 h-3.5"
          :class="typeMeta[ev.type].color"
        />
      </div>
      <span
        class="text-xs text-gray-400 dark:text-neutral-500 tabular-nums shrink-0"
      >
        {{ formatTime(ev.ts) }}
      </span>
      <img
        :src="`https://cravatar.eu/helmavatar/${encodeURIComponent(ev.player)}/16.png`"
        alt=""
        class="w-4 h-4 rounded"
      />
      <span class="text-sm text-gray-900 dark:text-white font-medium">
        {{ ev.player }}
      </span>
      <span class="text-sm text-gray-500 dark:text-neutral-400">
        {{ typeMeta[ev.type].label }}
      </span>
      <span
        v-if="ev.reason"
        class="text-xs text-gray-400 dark:text-neutral-500 truncate max-w-[8rem]"
      >
        ({{ ev.reason }})
      </span>
    </div>
    <div
      v-if="events.length === 0"
      class="text-sm text-gray-500 dark:text-neutral-400 italic px-2"
    >
      No events yet.
    </div>
  </div>
</template>
