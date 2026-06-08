<template>
  <Teleport to="body">
    <div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 pointer-events-none">
      <TransitionGroup name="toast">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="pointer-events-auto flex items-start gap-3 rounded-xl border shadow-lg px-4 py-3 min-w-[320px] max-w-[420px] backdrop-blur-sm"
          :class="toastClass(toast.type)"
        >
          <component
            :is="iconFor(toast.type)"
            class="w-5 h-5 shrink-0 mt-0.5"
          />
          <div class="flex-1 min-w-0">
            <p class="text-sm font-semibold leading-tight">{{ toast.title }}</p>
            <p
              v-if="toast.description"
              class="text-xs opacity-90 mt-1 leading-snug"
            >
              {{ toast.description }}
            </p>
          </div>
          <button
            @click="remove(toast.id)"
            class="opacity-50 hover:opacity-100 transition-opacity"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import {
  CheckCircle,
  XCircle,
  AlertTriangle,
  Info,
  X,
} from "lucide-vue-next";

const { toasts, remove } = useToastStack();

function toastClass(type: string) {
  switch (type) {
    case "success":
      return "bg-green-50/90 border-green-200 text-green-800 dark:bg-green-900/30 dark:border-green-800 dark:text-green-300";
    case "error":
      return "bg-red-50/90 border-red-200 text-red-800 dark:bg-red-900/30 dark:border-red-800 dark:text-red-300";
    case "warning":
      return "bg-yellow-50/90 border-yellow-200 text-yellow-800 dark:bg-yellow-900/30 dark:border-yellow-800 dark:text-yellow-300";
    case "info":
    default:
      return "bg-gray-50/90 border-gray-200 text-gray-800 dark:bg-gray-800/90 dark:border-gray-700 dark:text-gray-300";
  }
}

function iconFor(type: string) {
  switch (type) {
    case "success":
      return CheckCircle;
    case "error":
      return XCircle;
    case "warning":
      return AlertTriangle;
    case "info":
    default:
      return Info;
  }
}
</script>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(20px) scale(0.95);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(20px) scale(0.95);
}
</style>
