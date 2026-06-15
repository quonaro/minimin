<template>
  <div class="flex flex-col h-full">
    <div class="mb-4 space-y-2">
      <div class="flex items-center gap-2">
        <input
          :value="rawQuery"
          type="text"
          placeholder="Search Modrinth..."
          @input="onInput"
          class="flex-1 px-3 py-2 rounded-lg bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 text-sm text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-neutral-500 focus:ring-2 focus:ring-primary focus:outline-none"
          @keyup.enter="onSearch"
        />
        <span
          v-if="props.lockedType"
          class="px-2 py-2 rounded-lg bg-gray-100 dark:bg-neutral-700/50 border border-gray-200 dark:border-neutral-700 text-sm text-gray-700 dark:text-neutral-300 font-medium"
        >
          {{
            props.lockedType === "mod"
              ? "Mods"
              : props.lockedType === "resourcepack"
                ? "Resource Packs"
                : "Shaders"
          }}
        </span>
        <select
          v-else
          :value="projectType"
          class="px-2 py-2 rounded-lg bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:outline-none"
          @change="onProjectTypeChange"
        >
          <option value="mod">Mods</option>
          <option value="resourcepack">Resource Packs</option>
          <option value="shaderpack">Shaders</option>
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
          v-if="projectType === 'mod'"
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
          <span v-if="projectType === 'mod'">Loader: {{ loader }}</span>
          <span v-if="projectType === 'mod'">&middot;</span>
          <span>Version: {{ gameVersion }}</span>
        </div>
      </div>
    </div>

    <mod-install-modal
      v-model:show="showConfirm"
      :project="confirmProject"
      :version-id="confirmVersionId"
      :version-details="props.versionDetails"
      :dep-projects="props.depProjects"
      :project-type="props.projectType"
      :install-loading="isInstalling(confirmProjectId, confirmVersionId)"
      @install="onModalInstall"
      @load-project="emit('load-project', $event)"
    />

    <div
      class="flex-1 min-h-0 overflow-y-auto space-y-3 pr-1"
      ref="scrollContainer"
    >
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
            <div class="flex items-center gap-2">
              <p
                class="text-sm font-semibold text-gray-900 dark:text-white truncate"
              >
                {{ project.title }}
              </p>
              <span
                v-if="installStatus(project) === 'server'"
                class="shrink-0 inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300"
              >
                Server
              </span>
              <span
                v-else-if="installStatus(project) === 'client'"
                class="shrink-0 inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-300"
              >
                Client
              </span>
              <span
                v-else-if="installStatus(project) === 'both'"
                class="shrink-0 inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300"
              >
                Both
              </span>
              <span
                v-else-if="installStatus(project) === 'asset'"
                class="shrink-0 inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-gray-200 dark:bg-neutral-600 text-gray-700 dark:text-neutral-300"
              >
                Installed
              </span>
            </div>
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
              (!selectedVersion[project.project_id] &&
                !project.latest_version) ||
              isInstalling(
                project.project_id,
                selectedVersion[project.project_id] ||
                  project.latest_version ||
                  '',
              )
            "
            @click="install(project, selectedVersion[project.project_id] || '')"
          >
            <span
              v-if="
                isInstalling(
                  project.project_id,
                  selectedVersion[project.project_id] ||
                    project.latest_version ||
                    '',
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
  projectType?: "mod" | "resourcepack" | "shaderpack";
  installedModBasenames?: { server: Set<string>; client: Set<string> };
  installedAssetBasenames?: Set<string>;
  lockedType?: "mod" | "resourcepack" | "shaderpack";
}

const props = withDefaults(defineProps<Props>(), {
  versions: () => ({}),
  sideFilter: "all",
  projectType: "mod",
  installedModBasenames: () => ({
    server: new Set<string>(),
    client: new Set<string>(),
  }),
  installedAssetBasenames: () => new Set<string>(),
});
const emit = defineEmits<{
  search: [];
  "update:searchQuery": [value: string];
  "update:projectType": [value: "mod" | "resourcepack" | "shaderpack"];
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

const rawQuery = ref(props.searchQuery);
let debounceTimer: ReturnType<typeof setTimeout> | null = null;

function onInput(e: Event) {
  const val = (e.target as HTMLInputElement).value;
  rawQuery.value = val;
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    emit("update:searchQuery", val);
  }, 150);
}

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer);
});

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
function openConfirm(project: ModrinthProject, versionId: string) {
  confirmProject.value = project;
  confirmProjectId.value = project.project_id;
  confirmVersionId.value = versionId;
  showConfirm.value = true;
  if (!props.versionDetails?.[versionId]) {
    emit("load-version-details", versionId);
  }
}

function onModalInstall(
  projectId: string,
  versionId: string,
  depProjectIds: string[],
  target: "server" | "client" | "both",
) {
  emit("install", projectId, versionId, depProjectIds, target);
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

function onProjectTypeChange(e: Event) {
  if (props.lockedType) return;
  emit(
    "update:projectType",
    (e.target as HTMLSelectElement).value as
      | "mod"
      | "resourcepack"
      | "shaderpack",
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
  const vid = versionId || project.latest_version;
  if (!vid) return;
  openConfirm(project, vid);
}

function hasMatch(set: Set<string> | undefined, slug: string): boolean {
  if (!set?.size) return false;
  for (const name of set) {
    if (name.includes(slug) || slug.includes(name)) return true;
  }
  return false;
}

function installStatus(
  project: ModrinthProject,
): "server" | "client" | "both" | "asset" | null {
  const slug = project.slug.toLowerCase();
  if (props.projectType === "mod") {
    const inServer = hasMatch(props.installedModBasenames?.server, slug);
    const inClient = hasMatch(props.installedModBasenames?.client, slug);
    if (inServer && inClient) return "both";
    if (inServer) return "server";
    if (inClient) return "client";
    return null;
  }
  // resourcepack / shader
  if (hasMatch(props.installedAssetBasenames, slug)) return "asset";
  return null;
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
