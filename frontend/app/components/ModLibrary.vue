<template>
  <div class="flex flex-col h-full">
    <div class="mb-4 space-y-2">
      <div class="flex items-center gap-2">
        <input
          :value="searchQuery"
          type="text"
          placeholder="Search Modrinth..."
          @input="onInput"
          class="flex-1 px-3 py-2 rounded-lg bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 text-sm text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-neutral-500 focus:ring-2 focus:ring-primary focus:outline-none"
          @keyup.enter="onSearch"
        />
        <button
          class="px-3 py-2 rounded-lg bg-primary hover:bg-primary/90 text-white text-sm font-medium transition-colors disabled:opacity-50"
          :disabled="searchLoading"
          @click="onSearch"
        >
          <Search class="w-4 h-4" />
        </button>
      </div>
      <div
        class="flex items-center gap-2 text-xs text-gray-500 dark:text-neutral-400"
      >
        <span>Loader: {{ loader }}</span>
        <span>&middot;</span>
        <span>Version: {{ gameVersion }}</span>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto space-y-3 pr-1" ref="scrollContainer">
      <div
        v-if="searchResults.length === 0 && !searchLoading"
        class="text-center text-gray-500 dark:text-neutral-400 py-12 text-sm"
      >
        No results.
      </div>

      <div
        v-for="project in searchResults"
        :key="project.project_id"
        class="p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 hover:border-gray-300 dark:hover:border-neutral-600 transition-colors"
      >
        <div class="flex items-start gap-3">
          <img
            v-if="project.icon_url"
            :src="project.icon_url"
            class="w-10 h-10 rounded-lg object-cover shrink-0 bg-gray-200 dark:bg-neutral-600"
            alt=""
          />
          <div
            v-else
            class="w-10 h-10 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400 shrink-0"
          >
            <Box class="w-5 h-5" />
          </div>
          <div class="flex-1 min-w-0">
            <p
              class="text-sm font-semibold text-gray-900 dark:text-white truncate"
            >
              {{ project.title }}
            </p>
            <p class="text-xs text-gray-500 dark:text-neutral-400 line-clamp-2">
              {{ project.description }}
            </p>
            <p class="text-[10px] text-gray-400 dark:text-neutral-500 mt-0.5">
              {{ project.author }} &middot;
              {{ formatDownloads(project.downloads) }}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2 mt-2">
          <select
            v-model="selectedVersion[project.project_id]"
            class="flex-1 text-xs bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-600 rounded-lg px-2 py-1.5 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:outline-none"
            @focus="loadVersions(project.project_id)"
          >
            <option value="" disabled>Select version</option>
            <option
              v-for="v in props.versions[project.project_id] || []"
              :key="v.id"
              :value="v.id"
            >
              {{ v.version_number }}
            </option>
          </select>
          <button
            class="px-3 py-1.5 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white text-xs font-medium transition-colors disabled:opacity-50"
            :disabled="
              !selectedVersion[project.project_id] ||
              isInstalling(
                project.project_id,
                selectedVersion[project.project_id] || '',
              )
            "
            @click="
              install(
                project.project_id,
                selectedVersion[project.project_id] || '',
              )
            "
          >
            <span
              v-if="
                isInstalling(
                  project.project_id,
                  selectedVersion[project.project_id] || '',
                )
              "
              >...</span
            >
            <span v-else>Install</span>
          </button>
        </div>
      </div>

      <div
        v-if="searchLoading && searchResults.length > 0"
        class="py-4 flex justify-center"
      >
        <div
          class="w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin"
        />
      </div>

      <div ref="sentinel" class="h-1" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { Box, Search } from "lucide-vue-next";
import type {
  ModrinthProject,
  ModrinthVersion,
} from "~/composables/useModrinth";

interface Props {
  searchQuery: string;
  searchResults: ModrinthProject[];
  searchLoading: boolean;
  loader: string;
  gameVersion: string;
  installLoading: Record<string, boolean>;
  versions?: Record<string, ModrinthVersion[]>;
  hasMore?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  versions: () => ({}),
});
const emit = defineEmits<{
  search: [];
  "update:searchQuery": [value: string];
  install: [projectId: string, versionId: string];
  "load-versions": [projectId: string];
  "load-more": [];
}>();

const selectedVersion = ref<Record<string, string>>({});
const scrollContainer = ref<HTMLDivElement | null>(null);
const sentinel = ref<HTMLDivElement | null>(null);

let observer: IntersectionObserver | null = null;

onMounted(() => {
  if (!sentinel.value) return;
  observer = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];
      if (
        entry &&
        entry.isIntersecting &&
        props.hasMore &&
        !props.searchLoading
      ) {
        emit("load-more");
      }
    },
    { root: scrollContainer.value, rootMargin: "200px" },
  );
  observer.observe(sentinel.value);
});

onUnmounted(() => {
  if (observer && sentinel.value) {
    observer.unobserve(sentinel.value);
  }
  observer = null;
});

function onSearch() {
  emit("search");
}

function onInput(e: Event) {
  emit("update:searchQuery", (e.target as HTMLInputElement).value);
}

async function loadVersions(projectId: string) {
  if (props.versions[projectId]) return;
  emit("load-versions", projectId);
}

function isInstalling(projectId: string, versionId: string): boolean {
  if (!versionId) return false;
  return !!props.installLoading[`${projectId}:${versionId}`];
}

function install(projectId: string, versionId: string) {
  if (!versionId) return;
  emit("install", projectId, versionId);
}

function formatDownloads(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + "M";
  if (n >= 1_000) return (n / 1_000).toFixed(1) + "K";
  return String(n);
}

watch(
  () => props.searchResults,
  () => {
    selectedVersion.value = {};
  },
);
</script>
