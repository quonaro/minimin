<script setup lang="ts">
import { computed } from "vue";

function onAvatarError(e: Event) {
  const img = e.target as HTMLImageElement | null;
  if (img) {
    img.src = "/img/steve-head-32.png";
  }
}

export interface AllTimePlayer {
  name: string;
  lastSeen: number;
}

const props = defineProps<{
  players: AllTimePlayer[];
  filter?: string;
  online?: string[];
  ops?: string[];
  banned?: string[];
  whitelisted?: string[];
}>();

const emit = defineEmits<{
  (e: "kick", name: string): void;
  (e: "ban", name: string): void;
  (e: "unban", name: string): void;
  (e: "op", name: string): void;
  (e: "deop", name: string): void;
  (e: "wladd", name: string): void;
  (e: "wlremove", name: string): void;
}>();

const sorted = computed(() =>
  [...props.players]
    .filter((p) => {
      if (!props.filter) return true;
      return p.name.toLowerCase().includes(props.filter.toLowerCase());
    })
    .sort((a, b) => b.lastSeen - a.lastSeen),
);

function formatLastSeen(ts: number): string {
  const now = Date.now();
  const diff = now - ts;
  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  const d = new Date(ts);
  return d.toLocaleDateString([], { month: "short", day: "numeric" });
}
</script>

<template>
  <div class="space-y-1 max-h-80 overflow-y-auto pr-1">
    <div
      v-for="p in sorted"
      :key="p.name"
      class="flex items-center gap-3 px-2 py-1.5 rounded-lg hover:bg-gray-50 dark:hover:bg-neutral-700/30 transition-colors"
    >
      <img
        :src="`https://cravatar.eu/helmavatar/${encodeURIComponent(p.name)}/32.png`"
        alt=""
        class="w-8 h-8 rounded"
        @error="onAvatarError"
      />
      <span class="flex-1 text-sm font-medium text-gray-900 dark:text-white">
        {{ p.name }}
      </span>
      <span class="text-xs text-gray-500 dark:text-neutral-400 tabular-nums">
        {{ formatLastSeen(p.lastSeen) }}
      </span>
      <div class="flex items-center gap-1 shrink-0">
        <button
          v-if="online?.includes(p.name)"
          class="text-xs px-2 py-1 rounded bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400 hover:bg-red-200 dark:hover:bg-red-900/50"
          @click="emit('kick', p.name)"
        >
          Kick
        </button>
        <button
          v-if="banned?.includes(p.name)"
          class="text-xs px-2 py-1 rounded bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 hover:bg-green-200 dark:hover:bg-green-900/50"
          @click="emit('unban', p.name)"
        >
          Unban
        </button>
        <button
          v-else
          class="text-xs px-2 py-1 rounded bg-red-600 text-white hover:bg-red-700"
          @click="emit('ban', p.name)"
        >
          Ban
        </button>
        <button
          v-if="ops?.includes(p.name)"
          class="text-xs px-2 py-1 rounded bg-gray-100 text-gray-700 dark:bg-neutral-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-600"
          @click="emit('deop', p.name)"
        >
          Deop
        </button>
        <button
          v-else
          class="text-xs px-2 py-1 rounded bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400 hover:bg-primary-200 dark:hover:bg-primary-900/50"
          @click="emit('op', p.name)"
        >
          Op
        </button>
        <button
          v-if="whitelisted?.includes(p.name)"
          class="text-xs px-2 py-1 rounded bg-gray-100 text-gray-700 dark:bg-neutral-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-600"
          @click="emit('wlremove', p.name)"
        >
          Remove
        </button>
        <button
          v-else
          class="text-xs px-2 py-1 rounded bg-gray-100 text-gray-700 dark:bg-neutral-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-600"
          @click="emit('wladd', p.name)"
        >
          WL
        </button>
      </div>
    </div>
    <div
      v-if="players.length === 0"
      class="text-sm text-gray-500 dark:text-neutral-400 italic px-2"
    >
      No players have joined yet.
    </div>
  </div>
</template>
