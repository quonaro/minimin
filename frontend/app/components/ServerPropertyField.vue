<script setup lang="ts">
import type { ServerPropType } from "~/utils/serverProperties";

const props = defineProps<{
  itemKey: string;
  label: string;
  type: ServerPropType;
  dangerous?: boolean;
  modelValue: string;
  badgeStatus?: "modified" | "restart-required" | null;
  unlocked?: boolean;
  options?: string[];
}>();

const emit = defineEmits<{
  (e: "update:modelValue", val: string): void;
  (e: "unlock"): void;
}>();

const disabled = computed(() => props.dangerous && !props.unlocked);
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-3 gap-4 items-start">
    <div class="md:col-span-1 pt-2">
      <div class="flex items-center gap-2 flex-wrap">
        <label
          class="block text-sm font-medium text-gray-700 dark:text-neutral-300"
        >
          {{ label }}
        </label>
        <span
          v-if="badgeStatus === 'modified'"
          class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400"
        >
          Modified
        </span>
        <span
          v-if="badgeStatus === 'restart-required'"
          class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400"
        >
          Restart Required
        </span>
        <button
          v-if="dangerous && !unlocked"
          class="text-xs text-primary hover:underline font-medium"
          @click="emit('unlock')"
        >
          Unlock
        </button>
      </div>
      <p
        v-if="dangerous && unlocked"
        class="mt-1 text-xs text-red-500 dark:text-red-400"
      >
        This setting can break your server. Change with caution.
      </p>
    </div>
    <div class="md:col-span-2">
      <select
        v-if="type === 'select' || type === 'boolean'"
        :value="modelValue"
        :disabled="disabled"
        class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow disabled:opacity-50 disabled:cursor-not-allowed"
        @change="
          emit('update:modelValue', ($event.target as HTMLSelectElement).value)
        "
      >
        <option v-for="opt in options" :key="opt" :value="opt">
          {{ opt }}
        </option>
      </select>
      <number-input
        v-else-if="type === 'number'"
        :model-value="modelValue"
        :min="0"
        :disabled="disabled"
        class="w-full"
        @update:model-value="emit('update:modelValue', String($event))"
      />
      <input
        v-else
        :value="modelValue"
        type="text"
        :disabled="disabled"
        class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow disabled:opacity-50 disabled:cursor-not-allowed"
        @input="
          emit('update:modelValue', ($event.target as HTMLInputElement).value)
        "
      />
    </div>
  </div>
</template>
