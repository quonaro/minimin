<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean;
  title: string;
  description: string;
  confirmLabel?: string;
  danger?: boolean;
  showWipe?: boolean;
  simple?: boolean;
  expectedText?: string;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: boolean];
  confirm: [wipe: boolean];
}>();

const wipeData = ref(false);

const inputValue = ref("");
const randomWord = ref("");
const inputRef = ref<HTMLInputElement | null>(null);

function generateWord() {
  const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789";
  let result = "";
  for (let i = 0; i < 6; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  randomWord.value = result;
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      generateWord();
      inputValue.value = "";
      wipeData.value = false;
    }
  },
);

const targetText = computed(() => props.expectedText ?? randomWord.value);

const isValid = computed(
  () => props.simple || inputValue.value === targetText.value,
);

function onConfirm() {
  if (!isValid.value) return;
  emit("confirm", wipeData.value);
  emit("update:modelValue", false);
}

function onCancel() {
  emit("update:modelValue", false);
}
</script>

<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
  >
    <div
      class="bg-white dark:bg-neutral-800 rounded-xl shadow-lg max-w-md w-full p-6"
    >
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
        {{ title }}
      </h3>
      <p class="text-sm text-gray-600 dark:text-neutral-400 mb-4">
        {{ description }}
      </p>
      <label
        v-if="showWipe"
        class="flex items-center gap-2 mb-4 cursor-pointer"
      >
        <input
          v-model="wipeData"
          type="checkbox"
          class="w-4 h-4 text-primary rounded border-gray-300 dark:border-neutral-600 focus:ring-primary"
        />
        <span class="text-sm text-gray-700 dark:text-neutral-300">
          Also delete all server data (world, configs, mods)
        </span>
      </label>
      <template v-if="!simple">
        <p class="text-sm text-gray-600 dark:text-neutral-400 mb-2">
          Type
          <code
            class="font-mono font-bold bg-gray-100 dark:bg-neutral-700 px-1 rounded"
          >
            {{ targetText }}
          </code>
          to confirm:
        </p>
        <input
          ref="inputRef"
          v-model="inputValue"
          type="text"
          class="w-full mb-6 px-3 py-2 border border-gray-300 dark:border-neutral-600 rounded-lg bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:outline-none"
          :placeholder="targetText"
          @keyup.enter="onConfirm"
        />
      </template>
      <div v-else class="mb-6" />
      <div class="flex justify-end gap-3">
        <button
          class="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors"
          @click="onCancel"
        >
          Cancel
        </button>
        <button
          :disabled="!isValid"
          class="px-4 py-2 rounded-lg text-sm font-medium text-white transition-colors disabled:opacity-50"
          :class="
            danger
              ? 'bg-red-500 hover:bg-red-600'
              : 'bg-primary hover:bg-primary/90'
          "
          @click="onConfirm"
        >
          {{ confirmLabel || "Confirm" }}
        </button>
      </div>
    </div>
  </div>
</template>
