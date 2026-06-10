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
        <select
          :value="projectType"
          class="px-2 py-2 rounded-lg bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:outline-none"
          @change="onProjectTypeChange"
        >
          <option value="mod">Mods</option>
          <option value="resourcepack">Resource Packs</option>
          <option value="shader">Shaders</option>
        </select>
        <button
          class="px-3 py-2 rounded-lg bg-primary hover:bg-primary/90 text-white text-sm font-medium transition-colors disabled:opacity-50"
          :disabled="searchLoading"
          @click="onSearch"
        >
          <Search class="w-4 h-4" />
        </button>
      </div>
      <div class="flex items-center gap-2">
        <div
          class="flex rounded-lg border border-gray-300 dark:border-neutral-700 overflow-hidden"
        >
          <button
            v-for="opt in sideOptions"
            :key="opt.value"
            class="px-2 py-1 text-xs font-medium transition-colors"
            :class="
              sideFilter === opt.value
                ? 'bg-primary text-white'
                : 'bg-white dark:bg-neutral-800 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700'
            "
            @click="emit('update:sideFilter', opt.value)"
          >
            {{ opt.label }}
          </button>
        </div>
        <div
          class="flex items-center gap-2 text-xs text-gray-500 dark:text-neutral-400 ml-auto"
        >
          <span>Loader: {{ loader }}</span>
          <span>&middot;</span>
          <span>Version: {{ gameVersion }}</span>
        </div>
      </div>
    </div>

    <div
      v-if="showConfirm"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
      @click.self="showConfirm = false"
    >
      <div
        class="bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 w-full max-w-md max-h-[80vh] flex flex-col"
      >
        <div class="p-6 border-b border-gray-200 dark:border-neutral-700">
          <div class="flex items-start gap-3">
            <img
              v-if="confirmProject?.icon_url"
              :src="confirmProject.icon_url"
              class="w-12 h-12 rounded-xl object-cover shrink-0 bg-gray-200 dark:bg-neutral-600"
              alt=""
            />
            <div
              v-else
              class="w-12 h-12 rounded-xl bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400 shrink-0"
            >
              <Box class="w-6 h-6" />
            </div>
            <div class="flex-1 min-w-0">
              <h2
                class="text-lg font-bold text-gray-900 dark:text-white truncate"
              >
                {{ confirmProject?.title || "Install Mod" }}
              </h2>
              <p class="text-xs text-gray-500 dark:text-neutral-400">
                {{ confirmProject?.author }} &middot; Review dependencies before
                installing.
              </p>
            </div>
          </div>
        </div>

        <div class="p-6 overflow-y-auto space-y-3">
          <div
            v-for="dep in confirmDeps"
            :key="dep.project_id"
            class="flex items-start gap-3"
          >
            <input
              :id="dep.project_id"
              type="checkbox"
              :checked="selectedDepIds.has(dep.project_id)"
              :disabled="dep.dependency_type === 'required'"
              class="mt-2 w-4 h-4 rounded border-gray-300 dark:border-neutral-600 text-primary focus:ring-primary"
              @change="toggleDep(dep.project_id)"
            />
            <label :for="dep.project_id" class="flex-1 flex items-start gap-2">
              <img
                v-if="props.depProjects?.[dep.project_id]?.icon_url"
                :src="props.depProjects?.[dep.project_id]?.icon_url"
                class="w-8 h-8 rounded-lg object-cover shrink-0 bg-gray-200 dark:bg-neutral-600"
                alt=""
              />
              <div
                v-else
                class="w-8 h-8 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400 shrink-0"
              >
                <Box class="w-4 h-4" />
              </div>
              <div class="flex-1 min-w-0">
                <span
                  class="font-medium text-sm text-gray-900 dark:text-white block truncate"
                >
                  {{
                    props.depProjects?.[dep.project_id]?.title || dep.project_id
                  }}
                </span>
                <span class="text-[10px] text-gray-400 dark:text-neutral-500">
                  {{ props.depProjects?.[dep.project_id]?.author || "" }}
                </span>
                <span
                  class="ml-1 text-[10px] px-1 py-0.5 rounded capitalize"
                  :class="
                    dep.dependency_type === 'required'
                      ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                      : 'bg-gray-100 text-gray-600 dark:bg-neutral-700 dark:text-neutral-300'
                  "
                >
                  {{ dep.dependency_type }}
                </span>
              </div>
            </label>
          </div>
        </div>

        <div class="px-6 pb-3 space-y-2">
          <label
            v-if="projectType === 'mod'"
            class="flex items-center gap-2 cursor-pointer"
          >
            <input
              v-model="installToServer"
              type="checkbox"
              class="rounded border-gray-300 dark:border-neutral-600 text-primary focus:ring-primary"
            />
            <span class="text-sm text-gray-700 dark:text-neutral-300"
              >Install to server mods</span
            >
          </label>
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              v-model="installToClient"
              type="checkbox"
              class="rounded border-gray-300 dark:border-neutral-600 text-primary focus:ring-primary"
            />
            <span class="text-sm text-gray-700 dark:text-neutral-300"
              >Install to client mods</span
            >
          </label>
        </div>

        <div
          class="p-6 border-t border-gray-200 dark:border-neutral-700 flex gap-3 justify-end"
        >
          <button
            class="px-4 py-2 rounded-lg text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors font-medium text-sm"
            @click="showConfirm = false"
          >
            Cancel
          </button>
          <button
            class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white text-sm font-medium transition-colors disabled:opacity-50"
            :disabled="isInstalling(confirmProjectId, confirmVersionId)"
            @click="confirmInstall"
          >
            <span v-if="isInstalling(confirmProjectId, confirmVersionId)"
              >Installing...</span
            >
            <span v-else>Install</span>
          </button>
        </div>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto space-y-3 pr-1" ref="scrollContainer">
      <div
        v-if="filteredResults.length === 0 && !searchLoading"
        class="text-center text-gray-500 dark:text-neutral-400 py-12 text-sm"
      >
        No results.
      </div>

      <div
        v-for="project in filteredResults"
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
            @change="onVersionChange(project.project_id)"
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
            @click="install(project, selectedVersion[project.project_id] || '')"
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
        <div
          v-if="activeDeps(project.project_id).length > 0"
          class="mt-2 space-y-1"
        >
          <p class="text-[10px] text-gray-500 dark:text-neutral-400">
            Dependencies:
          </p>
          <div
            v-for="dep in activeDeps(project.project_id)"
            :key="dep.project_id"
            class="flex items-center gap-1 text-[10px]"
            :class="
              dep.dependency_type === 'required'
                ? 'text-amber-600 dark:text-amber-400'
                : 'text-gray-400 dark:text-neutral-500'
            "
          >
            <span class="font-medium capitalize">{{
              dep.dependency_type
            }}</span>
            <span class="truncate">{{ dep.project_id }}</span>
          </div>
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
  sideFilter?: "all" | "server" | "client";
  versionDetails?: Record<string, ModrinthVersion>;
  depProjects?: Record<string, ModrinthProject>;
  projectType?: "mod" | "resourcepack" | "shader";
}

const props = withDefaults(defineProps<Props>(), {
  versions: () => ({}),
  sideFilter: "all",
  projectType: "mod",
});
const emit = defineEmits<{
  search: [];
  "update:searchQuery": [value: string];
  "update:projectType": [value: "mod" | "resourcepack" | "shader"];
  install: [
    projectId: string,
    versionId: string,
    depProjectIds: string[],
    target: "server" | "client" | "both",
  ];
  "load-versions": [projectId: string];
  "load-more": [];
  "update:sideFilter": [value: "all" | "server" | "client"];
  "load-version-details": [versionId: string];
  "load-project": [projectId: string];
}>();

function activeDeps(projectId: string) {
  const vid = selectedVersion.value[projectId];
  if (!vid) return [];
  const details = props.versionDetails?.[vid];
  return (
    details?.dependencies?.filter(
      (d) => d.dependency_type !== "incompatible",
    ) || []
  );
}

function onVersionChange(projectId: string) {
  const vid = selectedVersion.value[projectId];
  if (vid && !props.versionDetails?.[vid]) {
    emit("load-version-details", vid);
  }
}

const sideOptions = [
  { label: "All", value: "all" as const },
  { label: "Server", value: "server" as const },
  { label: "Client", value: "client" as const },
];

const filteredResults = computed(() => {
  if (props.sideFilter === "all") return props.searchResults;
  return props.searchResults.filter((p) => {
    if (props.sideFilter === "server") {
      // Works on server and is not a client-only mod
      return p.server_side !== "unsupported" && p.client_side !== "required";
    }
    if (props.sideFilter === "client") {
      // Works on client and is not a server-only mod
      return p.client_side !== "unsupported" && p.server_side !== "required";
    }
    return true;
  });
});

const selectedVersion = ref<Record<string, string>>({});
const scrollContainer = ref<HTMLDivElement | null>(null);
const sentinel = ref<HTMLDivElement | null>(null);

const showConfirm = ref(false);
const confirmProject = ref<ModrinthProject | null>(null);
const confirmProjectId = ref("");
const confirmVersionId = ref("");
const selectedDepIds = ref<Set<string>>(new Set());
const installToServer = ref(true);
const installToClient = ref(false);

const confirmDeps = computed(() => {
  const details = props.versionDetails?.[confirmVersionId.value];
  return (
    details?.dependencies?.filter(
      (d) => d.dependency_type !== "incompatible",
    ) || []
  );
});

function openConfirm(project: ModrinthProject, versionId: string) {
  confirmProject.value = project;
  confirmProjectId.value = project.project_id;
  confirmVersionId.value = versionId;
  const deps =
    props.versionDetails?.[versionId]?.dependencies?.filter(
      (d) => d.dependency_type !== "incompatible",
    ) || [];
  const ids = new Set<string>();
  for (const d of deps) {
    if (d.dependency_type === "required" || d.dependency_type === "embedded") {
      ids.add(d.project_id);
    }
    if (!props.depProjects?.[d.project_id]) {
      emit("load-project", d.project_id);
    }
  }
  selectedDepIds.value = ids;
  installToServer.value = true;
  installToClient.value = false;
  showConfirm.value = true;
}

function toggleDep(projectId: string) {
  const next = new Set(selectedDepIds.value);
  if (next.has(projectId)) {
    next.delete(projectId);
  } else {
    next.add(projectId);
  }
  selectedDepIds.value = next;
}

function getInstallTarget(): "server" | "client" | "both" {
  if (installToServer.value && installToClient.value) return "both";
  if (installToClient.value) return "client";
  return "server";
}

function confirmInstall() {
  if (!confirmVersionId.value) return;
  emit(
    "install",
    confirmProjectId.value,
    confirmVersionId.value,
    Array.from(selectedDepIds.value),
    getInstallTarget(),
  );
  showConfirm.value = false;
}

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

function onProjectTypeChange(e: Event) {
  emit(
    "update:projectType",
    (e.target as HTMLSelectElement).value as "mod" | "resourcepack" | "shader",
  );
}

async function loadVersions(projectId: string) {
  if (props.versions[projectId]) return;
  emit("load-versions", projectId);
}

function isInstalling(projectId: string, versionId: string): boolean {
  if (!versionId) return false;
  return !!props.installLoading[`${projectId}:${versionId}`];
}

function install(project: ModrinthProject, versionId: string) {
  if (!versionId) return;
  openConfirm(project, versionId);
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
