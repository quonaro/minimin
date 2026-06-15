<script setup lang="ts">
import { Box } from "lucide-vue-next";
import type {
  ModrinthProject,
  ModrinthVersion,
} from "~/composables/useModrinth";

const props = defineProps<{
  show: boolean;
  project: ModrinthProject | null;
  versionId: string;
  versionDetails?: Record<string, ModrinthVersion>;
  depProjects?: Record<string, ModrinthProject>;
  projectType?: "mod" | "resourcepack" | "shaderpack";
  installLoading?: boolean;
}>();

const emit = defineEmits<{
  (e: "update:show", val: boolean): void;
  (
    e: "install",
    projectId: string,
    versionId: string,
    depProjectIds: string[],
    target: "server" | "client" | "both",
  ): void;
  (e: "load-project", projectId: string): void;
}>();

const selectedDepIds = ref<Set<string>>(new Set());
const installToServer = ref(true);
const installToClient = ref(false);

const confirmDeps = computed(() => {
  const details = props.versionDetails?.[props.versionId];
  return (
    details?.dependencies?.filter(
      (d) => d.dependency_type !== "incompatible",
    ) || []
  );
});

watch(
  () => props.show,
  (val) => {
    if (val) {
      const deps =
        props.versionDetails?.[props.versionId]?.dependencies?.filter(
          (d) => d.dependency_type !== "incompatible",
        ) || [];
      const ids = new Set<string>();
      for (const d of deps) {
        if (
          d.dependency_type === "required" ||
          d.dependency_type === "embedded"
        ) {
          ids.add(d.project_id);
        }
        if (!props.depProjects?.[d.project_id]) {
          emit("load-project", d.project_id);
        }
      }
      selectedDepIds.value = ids;
      if (props.projectType !== "mod") {
        installToServer.value = false;
        installToClient.value = true;
      } else if (props.project) {
        installToServer.value =
          props.project.server_side === "required" ||
          props.project.server_side !== "unsupported";
        installToClient.value =
          props.project.client_side === "required" ||
          props.project.client_side === "optional";
      } else {
        installToServer.value = true;
        installToClient.value = false;
      }
    }
  },
);

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
  if (!props.versionId || !props.project) return;
  emit(
    "install",
    props.project.project_id,
    props.versionId,
    Array.from(selectedDepIds.value),
    getInstallTarget(),
  );
  emit("update:show", false);
}
</script>

<template>
  <div
    v-if="show"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    @click.self="emit('update:show', false)"
  >
    <div
      class="bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 w-full max-w-md max-h-[80vh] flex flex-col"
    >
      <div class="p-6 border-b border-gray-200 dark:border-neutral-700">
        <div class="flex items-start gap-3">
          <img
            v-if="project?.icon_url"
            :src="project.icon_url"
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
              {{ project?.title || "Install Mod" }}
            </h2>
            <p class="text-xs text-gray-500 dark:text-neutral-400">
              {{ project?.author }} &middot; Review dependencies before
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
              v-if="depProjects?.[dep.project_id]?.icon_url"
              :src="depProjects?.[dep.project_id]?.icon_url"
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
                {{ depProjects?.[dep.project_id]?.title || dep.project_id }}
              </span>
              <span class="text-[10px] text-gray-400 dark:text-neutral-500">
                {{ depProjects?.[dep.project_id]?.author || "" }}
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
          @click="emit('update:show', false)"
        >
          Cancel
        </button>
        <button
          class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white text-sm font-medium transition-colors disabled:opacity-50"
          :disabled="installLoading"
          @click="confirmInstall"
        >
          <span v-if="installLoading">Installing...</span>
          <span v-else>Install</span>
        </button>
      </div>
    </div>
  </div>
</template>
