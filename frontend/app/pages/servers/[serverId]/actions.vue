<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Actions</h1>
      <button
        @click="logOpen = true"
        class="px-4 py-2 bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-neutral-700 transition-colors flex items-center gap-2 text-gray-900 dark:text-white"
      >
        <Activity class="w-4 h-4 text-indigo-500" />
        Action Log
      </button>
    </div>

    <TasksGroup :server-id="serverId" />
    <BackupsGroup :server-id="serverId" />
    <EventsGroup :server-id="serverId" />

    <Teleport to="body">
      <Transition name="slide-right">
        <div
          v-if="logOpen"
          class="fixed top-0 right-0 h-screen w-80 sm:w-96 bg-white dark:bg-neutral-800 border-l border-gray-200 dark:border-neutral-700 shadow-2xl z-50 flex flex-col"
        >
          <div
            class="p-4 border-b border-gray-200 dark:border-neutral-700 flex items-center justify-between"
          >
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              Action Log
            </h2>
            <button
              @click="logOpen = false"
              class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-neutral-700 text-gray-500 dark:text-neutral-400 transition-colors"
              aria-label="Close action log"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="flex-1 overflow-y-auto p-4 no-scrollbar">
            <ActionLogPanel :server-id="serverId" :show-title="false" />
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { Activity, X } from "lucide-vue-next";

const route = useRoute();
const serverId = computed(() => route.params.serverId as string);

const logOpen = ref(false);
</script>

<style scoped>
.slide-right-enter-active,
.slide-right-leave-active {
  transition: transform 300ms ease-in-out;
}

.slide-right-enter-from,
.slide-right-leave-to {
  transform: translateX(100%);
}

.no-scrollbar::-webkit-scrollbar {
  display: none;
}

.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
