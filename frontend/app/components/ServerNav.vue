<template>
  <div class="flex flex-wrap gap-2">
    <NuxtLink
      v-for="item in items"
      :key="item.to"
      :to="item.to"
      :class="[
        'inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors',
        isActive(item.to)
          ? 'bg-primary text-white'
          : 'bg-gray-100 dark:bg-neutral-700 text-gray-700 dark:text-neutral-300 hover:bg-gray-200 dark:hover:bg-neutral-600',
      ]"
    >
      <component :is="item.icon" class="w-4 h-4" />
      {{ item.label }}
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import {
  Folder,
  LayoutDashboard,
  Puzzle,
  ScrollText,
  Settings,
  Terminal,
  Users,
} from "lucide-vue-next";

const route = useRoute();
const { agentId } = useCurrentAgent();
const serverId = route.params.serverId as string;

const base = computed(() => `/agent/${agentId.value}/servers/${serverId}`);

const items = computed(() => [
  { to: base.value, label: "Overview", icon: LayoutDashboard },
  { to: `${base.value}/console`, label: "Console", icon: Terminal },
  { to: `${base.value}/logs`, label: "Logs", icon: ScrollText },
  { to: `${base.value}/players`, label: "Players", icon: Users },
  { to: `${base.value}/files`, label: "Files", icon: Folder },
  { to: `${base.value}/mods`, label: "Mods", icon: Puzzle },
  { to: `${base.value}/plugins`, label: "Plugins", icon: Puzzle },
  { to: `${base.value}/settings`, label: "Settings", icon: Settings },
]);

function isActive(path: string): boolean {
  if (path === base.value) {
    return route.path === path || route.path === `${path}/`;
  }
  return route.path === path || route.path.startsWith(`${path}/`);
}
</script>
