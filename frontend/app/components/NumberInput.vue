<script setup lang="ts">
import { ChevronUp, ChevronDown } from "lucide-vue-next";
import { useAttrs } from "vue";

defineOptions({ inheritAttrs: false });

interface Props {
  modelValue?: number | string | null;
  min?: number | string;
  max?: number | string;
  step?: number | string;
  disabled?: boolean;
  size?: "md" | "sm";
  class?: string;
}

const props = withDefaults(defineProps<Props>(), {
  step: 1,
  size: "md",
});

const emit = defineEmits<{
  (e: "update:modelValue", v: number): void;
}>();

const attrs = useAttrs();

function toNum(v: unknown): number {
  const n = Number(v);
  return Number.isFinite(n) ? n : 0;
}

function clampAndEmit(raw: number) {
  let val = raw;
  if (!Number.isFinite(val)) {
    val = toNum(props.min);
  }
  const min = toNum(props.min);
  const max = toNum(props.max);
  if (props.min !== undefined && val < min) val = min;
  if (props.max !== undefined && val > max) val = max;
  emit("update:modelValue", val);
}

function onInput(e: Event) {
  const target = e.target as HTMLInputElement;
  clampAndEmit(target.valueAsNumber);
}

function increment() {
  const step = toNum(props.step) || 1;
  clampAndEmit(toNum(props.modelValue) + step);
}

function decrement() {
  const step = toNum(props.step) || 1;
  clampAndEmit(toNum(props.modelValue) - step);
}
</script>

<template>
  <div
    class="relative inline-flex items-stretch w-full rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 focus-within:ring-2 focus-within:ring-primary focus-within:border-transparent overflow-hidden transition-shadow"
    :class="props.class"
  >
    <input
      type="number"
      :value="modelValue ?? ''"
      :min="min"
      :max="max"
      :step="step"
      :disabled="disabled"
      class="flex-1 min-w-0 bg-transparent text-gray-900 dark:text-white outline-none disabled:opacity-50 disabled:cursor-not-allowed"
      :class="props.size === 'sm' ? 'px-2 text-sm' : 'px-3 py-2'"
      v-bind="attrs"
      @input="onInput"
    />
    <div
      class="flex flex-col shrink-0 border-l border-gray-200 dark:border-neutral-600"
      :class="props.size === 'sm' ? 'w-5' : 'w-6'"
    >
      <button
        type="button"
        class="flex items-center justify-center flex-1 hover:bg-gray-100 dark:hover:bg-neutral-600 text-gray-500 dark:text-neutral-400 transition-colors disabled:opacity-50"
        :disabled="disabled"
        @click="increment"
      >
        <ChevronUp :class="props.size === 'sm' ? 'w-3 h-3' : 'w-3.5 h-3.5'" />
      </button>
      <button
        type="button"
        class="flex items-center justify-center flex-1 border-t border-gray-200 dark:border-neutral-600 hover:bg-gray-100 dark:hover:bg-neutral-600 text-gray-500 dark:text-neutral-400 transition-colors disabled:opacity-50"
        :disabled="disabled"
        @click="decrement"
      >
        <ChevronDown :class="props.size === 'sm' ? 'w-3 h-3' : 'w-3.5 h-3.5'" />
      </button>
    </div>
  </div>
</template>

<style scoped>
input[type="number"]::-webkit-inner-spin-button,
input[type="number"]::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
input[type="number"] {
  -webkit-appearance: none;
  -moz-appearance: textfield;
  appearance: none;
}
</style>
