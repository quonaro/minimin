<script setup lang="ts">
import { Archive, Copy, ExternalLink } from "lucide-vue-next";
import type { ArchiveInfo } from "~/composables/useClientMods";

const props = defineProps<{
  show: boolean;
  archiveLoading?: boolean;
  archiveResult?: ArchiveInfo | null;
}>();

const emit = defineEmits<{
  (e: "update:show", val: boolean): void;
  (e: "generate", formats: string[], ttl: number, include: string[]): void;
}>();

const archiveFormats = ref<string[]>(["zip"]);
const archiveInclude = ref<string[]>(["mods"]);
const archiveTTL = ref(24);

const archiveLink = computed(() => {
  if (!props.archiveResult) return "";
  const base = window.location.origin;
  return `${base}/client-archive/${props.archiveResult.token}`;
});

function generate() {
  emit("generate", archiveFormats.value, archiveTTL.value, archiveInclude.value);
}

function copyLink() {
  if (!archiveLink.value) return;
  navigator.clipboard.writeText(archiveLink.value);
  useToast().show("success", "Link copied");
}

watch(
  () => props.show,
  (val) => {
    if (!val) {
      archiveFormats.value = ["zip"];
      archiveInclude.value = ["mods"];
      archiveTTL.value = 24;
    }
  },
);
</script>

<template>
  <div
    v-if="show"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    @click.self="emit('update:show', false)"
  >
    <div
      class="bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 w-full max-w-md"
    >
      <div class="p-6 border-b border-gray-200 dark:border-neutral-700">
        <h2 class="text-xl font-bold text-gray-900 dark:text-white">
          Export Client Archive
        </h2>
      </div>
      <div class="p-6 space-y-4">
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
            Formats
          </label>
          <div class="flex flex-wrap gap-3">
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                v-model="archiveFormats"
                type="checkbox"
                value="zip"
                class="rounded border-gray-300 text-primary focus:ring-primary"
              />
              <span class="text-sm text-gray-700 dark:text-neutral-300"
                >ZIP</span
              >
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                v-model="archiveFormats"
                type="checkbox"
                value="mrpack"
                class="rounded border-gray-300 text-primary focus:ring-primary"
              />
              <span class="text-sm text-gray-700 dark:text-neutral-300"
                >Mrpack</span
              >
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                v-model="archiveFormats"
                type="checkbox"
                value="curseforge"
                class="rounded border-gray-300 text-primary focus:ring-primary"
              />
              <span class="text-sm text-gray-700 dark:text-neutral-300"
                >CurseForge</span
              >
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                v-model="archiveFormats"
                type="checkbox"
                value="prism"
                class="rounded border-gray-300 text-primary focus:ring-primary"
              />
              <span class="text-sm text-gray-700 dark:text-neutral-300"
                >Prism / MultiMC</span
              >
            </label>
          </div>
          <p class="text-xs text-gray-400 dark:text-neutral-500 mt-1">
            All formats include the selected asset types (mods, resource packs,
            shader packs) in their respective folders.
          </p>
        </div>
        <div>
          <label
            class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-2"
          >
            Link Expiration
          </label>
          <select
            v-model="archiveTTL"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-primary outline-none"
          >
            <option :value="1">1 hour</option>
            <option :value="6">6 hours</option>
            <option :value="24">24 hours</option>
            <option :value="168">7 days</option>
          </select>
        </div>
        <div v-if="archiveResult" class="flex items-center gap-3">
          <a
            :href="archiveLink"
            target="_blank"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-white text-sm font-medium hover:bg-primary/90 transition-colors"
          >
            <ExternalLink class="w-4 h-4" />
            Open
          </a>
          <button
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
            @click="copyLink"
          >
            <Copy class="w-4 h-4" />
            Copy
          </button>
        </div>
      </div>
      <div
        class="p-6 border-t border-gray-200 dark:border-neutral-700 flex items-center justify-between"
      >
        <p
          v-if="archiveResult"
          class="text-xs text-gray-500 dark:text-neutral-400"
        >
          Expires: {{ new Date(archiveResult.expiresAt).toLocaleString() }}
        </p>
        <div v-else />
        <div class="flex gap-3">
          <button
            class="px-4 py-2 rounded-lg text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors font-medium text-sm"
            @click="emit('update:show', false)"
          >
            Close
          </button>
          <button
            class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-primary hover:bg-primary/90 text-white text-sm font-medium transition-colors disabled:opacity-50"
            :disabled="
              archiveLoading ||
              archiveFormats.length === 0 ||
              archiveInclude.length === 0
            "
            @click="generate"
          >
            <Archive class="w-4 h-4" />
            {{ archiveLoading ? "Generating..." : "Generate" }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
