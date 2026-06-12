<script setup lang="ts">
import {
  Archive,
  Copy,
  ExternalLink,
  Trash2,
  Link,
  RefreshCw,
  Download,
} from "lucide-vue-next";
import type {
  ArchiveInfo,
  ArchiveLinkEntry,
} from "~/composables/useClientMods";

const props = defineProps<{
  show: boolean;
  archiveLoading?: boolean;
  archiveResult?: ArchiveInfo | null;
  archiveLinks?: ArchiveLinkEntry[];
  linksLoading?: boolean;
}>();

const emit = defineEmits<{
  (e: "update:show", val: boolean): void;
  (e: "generate", ttl: number, include: string[]): void;
  (e: "refresh-links"): void;
  (e: "delete-link", token: string): void;
}>();

const activeTab = ref<"generate" | "links">("generate");
const archiveInclude = ref<string[]>(["mods"]);
const archiveTTLDays = ref(1);

const ttlOptions = [
  { value: 1, label: "1 day" },
  { value: 7, label: "7 days" },
  { value: 30, label: "30 days" },
  { value: 90, label: "3 months" },
  { value: 180, label: "6 months" },
  { value: 365, label: "12 months" },
];

function generate() {
  emit("generate", archiveTTLDays.value * 24, archiveInclude.value);
}

function archiveLink(token: string) {
  const base = window.location.origin;
  return `${base}/client-archive/${token}`;
}

function copyArchiveLink(token: string) {
  navigator.clipboard.writeText(archiveLink(token));
  useToast().show("success", "Link copied");
}

function openArchive(token: string) {
  window.open(archiveLink(token), "_blank");
}

function deleteLink(token: string) {
  emit("delete-link", token);
}

watch(
  () => props.show,
  (val) => {
    if (val) {
      emit("refresh-links");
    } else {
      archiveInclude.value = ["mods"];
      archiveTTLDays.value = 1;
      activeTab.value = "generate";
    }
  },
);

watch(activeTab, (tab) => {
  if (tab === "links") {
    emit("refresh-links");
  }
});
</script>

<template>
  <div
    v-if="show"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    @click.self="emit('update:show', false)"
  >
    <div
      class="bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 w-full max-w-lg flex flex-col max-h-[85vh]"
    >
      <div class="p-6 border-b border-gray-200 dark:border-neutral-700">
        <h2 class="text-xl font-bold text-gray-900 dark:text-white">
          Export Client Archive
        </h2>
      </div>

      <!-- Tabs -->
      <div class="flex border-b border-gray-200 dark:border-neutral-700">
        <button
          class="flex-1 px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[1px]"
          :class="
            activeTab === 'generate'
              ? 'border-primary text-primary'
              : 'border-transparent text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-300'
          "
          @click="activeTab = 'generate'"
        >
          Generate
        </button>
        <button
          class="flex-1 px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[1px]"
          :class="
            activeTab === 'links'
              ? 'border-primary text-primary'
              : 'border-transparent text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-300'
          "
          @click="activeTab = 'links'"
        >
          Links
        </button>
      </div>

      <div class="p-6 overflow-y-auto flex-1">
        <!-- Generate Tab -->
        <div v-if="activeTab === 'generate'" class="space-y-4">
          <div>
            <label
              class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-2"
            >
              Include
            </label>
            <div class="flex flex-wrap gap-3">
              <label class="flex items-center gap-2 cursor-pointer">
                <input
                  v-model="archiveInclude"
                  type="checkbox"
                  value="mods"
                  class="rounded border-gray-300 text-primary focus:ring-primary"
                />
                <span class="text-sm text-gray-700 dark:text-neutral-300"
                  >Mods</span
                >
              </label>
              <label class="flex items-center gap-2 cursor-pointer">
                <input
                  v-model="archiveInclude"
                  type="checkbox"
                  value="resourcepacks"
                  class="rounded border-gray-300 text-primary focus:ring-primary"
                />
                <span class="text-sm text-gray-700 dark:text-neutral-300"
                  >Resource Packs</span
                >
              </label>
              <label class="flex items-center gap-2 cursor-pointer">
                <input
                  v-model="archiveInclude"
                  type="checkbox"
                  value="shaderpacks"
                  class="rounded border-gray-300 text-primary focus:ring-primary"
                />
                <span class="text-sm text-gray-700 dark:text-neutral-300"
                  >Shader Packs</span
                >
              </label>
            </div>
          </div>
          <div>
            <label
              class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-2"
            >
              Link Expiration
            </label>
            <select
              v-model="archiveTTLDays"
              class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-primary outline-none"
            >
              <option
                v-for="opt in ttlOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </option>
            </select>
          </div>
          <div v-if="archiveResult" class="space-y-2">
            <p class="text-sm font-medium text-gray-700 dark:text-neutral-300">
              Generated link
            </p>
            <div class="flex items-center gap-2">
              <input
                :value="archiveLink(archiveResult.token)"
                readonly
                class="flex-1 px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-gray-50 dark:bg-neutral-900 text-sm text-gray-900 dark:text-white outline-none"
              />
              <button
                class="inline-flex items-center gap-1 px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
                @click="copyArchiveLink(archiveResult.token)"
              >
                <Copy class="w-4 h-4" />
              </button>
              <a
                :href="archiveLink(archiveResult.token)"
                target="_blank"
                class="inline-flex items-center gap-1 px-3 py-2 rounded-lg bg-primary text-white text-sm font-medium hover:bg-primary/90 transition-colors"
              >
                <ExternalLink class="w-4 h-4" />
              </a>
            </div>
            <p class="text-xs text-gray-500 dark:text-neutral-400">
              Expires: {{ new Date(archiveResult.expiresAt).toLocaleString() }}
            </p>
          </div>
        </div>

        <!-- Links Tab -->
        <div v-else class="space-y-4">
          <div class="flex items-center justify-between">
            <p class="text-sm font-medium text-gray-700 dark:text-neutral-300">
              Active links ({{ archiveLinks?.length ?? 0 }})
            </p>
            <button
              class="inline-flex items-center gap-1 text-xs text-primary hover:text-primary/80 font-medium"
              :disabled="linksLoading"
              @click="emit('refresh-links')"
            >
              <RefreshCw
                class="w-3.5 h-3.5"
                :class="linksLoading ? 'animate-spin' : ''"
              />
              Refresh
            </button>
          </div>

          <div
            v-if="!archiveLinks || archiveLinks.length === 0"
            class="text-center text-sm text-gray-500 dark:text-neutral-400 py-8"
          >
            No active archive links.
          </div>

          <div v-else class="space-y-3">
            <div
              v-for="link in archiveLinks"
              :key="link.token"
              class="p-3 rounded-xl border border-gray-200 dark:border-neutral-700 bg-gray-50 dark:bg-neutral-700/30 space-y-2"
            >
              <div class="flex items-center gap-2">
                <input
                  :value="archiveLink(link.token)"
                  readonly
                  class="flex-1 px-2 py-1.5 rounded border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-800 text-xs text-gray-900 dark:text-white outline-none truncate"
                />
                <button
                  class="text-gray-400 hover:text-primary transition-colors p-1"
                  title="Copy link"
                  @click="copyArchiveLink(link.token)"
                >
                  <Copy class="w-4 h-4" />
                </button>
                <button
                  class="text-gray-400 hover:text-blue-500 transition-colors p-1"
                  title="Open link"
                  @click="openArchive(link.token)"
                >
                  <ExternalLink class="w-4 h-4" />
                </button>
                <button
                  class="text-gray-400 hover:text-red-500 transition-colors p-1"
                  title="Delete link"
                  @click="deleteLink(link.token)"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
              <div
                class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-neutral-400"
              >
                <span
                  >Created:
                  {{ new Date(link.createdAt).toLocaleDateString() }}</span
                >
                <span
                  >Expires:
                  {{ new Date(link.expiresAt).toLocaleDateString() }}</span
                >
                <span class="font-medium text-gray-700 dark:text-neutral-300"
                  >Downloads: {{ link.totalDownloads }}</span
                >
              </div>
              <div
                v-if="link.downloadCounts"
                class="flex flex-wrap gap-2 text-xs text-gray-500 dark:text-neutral-400"
              >
                <span
                  v-for="fmt in link.formats"
                  :key="fmt"
                  class="px-1.5 py-0.5 rounded bg-gray-200 dark:bg-neutral-600"
                >
                  {{ fmt }}: {{ link.downloadCounts[fmt] ?? 0 }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        class="p-6 border-t border-gray-200 dark:border-neutral-700 flex items-center justify-end gap-3"
      >
        <button
          class="px-4 py-2 rounded-lg text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors font-medium text-sm"
          @click="emit('update:show', false)"
        >
          Close
        </button>
        <button
          v-if="activeTab === 'generate'"
          class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-primary hover:bg-primary/90 text-white text-sm font-medium transition-colors disabled:opacity-50"
          :disabled="archiveLoading || archiveInclude.length === 0"
          @click="generate"
        >
          <Archive class="w-4 h-4" />
          {{ archiveLoading ? "Generating..." : "Generate" }}
        </button>
      </div>
    </div>
  </div>
</template>
