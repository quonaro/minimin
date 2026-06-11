<script setup lang="ts">
import type { Component } from "vue";
import { ChevronRight } from "lucide-vue-next";

const props = defineProps<{
  name: string;
  icon?: Component;
  collapsed: boolean;
  dangerous?: boolean;
}>();

const emit = defineEmits<{
  (e: "toggle"): void;
}>();
</script>

<template>
  <div>
    <button
      class="flex items-center gap-2 w-full text-left px-3 py-2 rounded-lg transition-colors"
      :class="
        dangerous
          ? 'bg-red-100 dark:bg-red-900/20 hover:bg-red-200 dark:hover:bg-red-900/30'
          : 'bg-gray-100 dark:bg-neutral-700/50 hover:bg-gray-200 dark:hover:bg-neutral-700'
      "
      @click="emit('toggle')"
    >
      <component
        :is="icon"
        v-if="icon"
        class="w-4 h-4 shrink-0"
        :class="
          dangerous
            ? 'text-red-600 dark:text-red-400'
            : 'text-gray-500 dark:text-neutral-400'
        "
      />
      <span
        class="text-sm font-semibold uppercase tracking-wider"
        :class="
          dangerous
            ? 'text-red-700 dark:text-red-300'
            : 'text-gray-700 dark:text-neutral-300'
        "
      >
        {{ name }}
      </span>
      <ChevronRight
        class="w-4 h-4 transition-transform shrink-0 ml-auto"
        :class="[
          dangerous
            ? 'text-red-600 dark:text-red-400'
            : 'text-gray-500 dark:text-neutral-400',
          collapsed ? '' : 'rotate-90',
        ]"
      />
    </button>
    <div v-show="!collapsed" class="mt-4 space-y-4">
      <slot />
    </div>
  </div>
</template>
