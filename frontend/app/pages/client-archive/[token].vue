<template>
  <div
    class="min-h-screen flex items-center justify-center p-4 bg-gray-50 dark:bg-neutral-900"
  >
    <div v-if="loading" class="text-center text-gray-500 dark:text-neutral-400">
      <Loader2 class="w-8 h-8 animate-spin mx-auto mb-2" />
      <p>Loading archive...</p>
    </div>

    <div v-else-if="error" class="w-full max-w-md">
      <div
        class="bg-white dark:bg-neutral-800 rounded-2xl shadow-xl border border-gray-200 dark:border-neutral-700 p-8 text-center"
      >
        <div
          class="w-16 h-16 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center mx-auto mb-4"
        >
          <XCircle class="w-8 h-8 text-red-500 dark:text-red-400" />
        </div>
        <h2 class="text-xl font-bold text-gray-900 dark:text-white mb-2">
          Archive not found
        </h2>
        <p class="text-sm text-gray-500 dark:text-neutral-400 mb-6">
          {{ error }}
        </p>
        <div class="flex justify-center gap-3">
          <button
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
            @click="goBack()"
          >
            <ArrowLeft class="w-4 h-4" />
            Go Back
          </button>
          <a
            href="/"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-white text-sm font-medium hover:bg-primary/90 transition-colors"
          >
            <Home class="w-4 h-4" />
            Go Home
          </a>
        </div>
      </div>
    </div>

    <div
      v-else
      class="w-full max-w-3xl bg-white dark:bg-neutral-800 rounded-2xl shadow-xl border border-gray-200 dark:border-neutral-700 p-8"
    >
      <div class="flex items-center justify-center gap-3 mb-4">
        <img src="/img/MiniMin_L.avif" alt="MiniMin" class="h-10 w-auto" />
        <img
          :src="
            $colorMode.value === 'dark'
              ? '/img/MiniMin_T_light.avif'
              : '/img/MiniMin_T.avif'
          "
          alt="MiniMin"
          class="h-9 w-auto"
        />
      </div>
      <div class="text-center mb-6">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ info?.serverName || "Modpack" }}
        </h1>
        <p class="text-sm text-gray-500 dark:text-neutral-400 mt-1">
          Client modpack export
        </p>
      </div>

      <div class="space-y-3 mb-6">
        <div
          v-if="hasFormat('zip')"
          class="flex items-center justify-between p-4 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-200 dark:border-neutral-600"
        >
          <div class="flex items-center gap-3">
            <FileArchive class="w-8 h-8 text-amber-600 dark:text-amber-400" />
            <div>
              <p class="font-medium text-gray-900 dark:text-white">
                ZIP Archive
              </p>
              <p class="text-xs text-gray-500 dark:text-neutral-400">
                All client mods as .jar files
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button
              class="inline-flex items-center gap-1 px-2 py-1.5 rounded-lg border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
              @click="copyUrl('zip')"
            >
              <Link class="w-4 h-4" />
              {{ copiedFormat === "zip" ? "Copied!" : "Copy link" }}
            </button>
            <a
              :href="downloadUrl('zip')"
              class="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-primary text-white text-sm font-medium hover:bg-primary/90 transition-colors"
            >
              <Download class="w-4 h-4" />
              Download
            </a>
          </div>
        </div>

        <div
          v-if="hasFormat('mrpack')"
          class="flex items-center justify-between p-4 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-200 dark:border-neutral-600"
        >
          <div class="flex items-center gap-3">
            <FileBox class="w-8 h-8 text-emerald-600 dark:text-emerald-400" />
            <div>
              <p class="font-medium text-gray-900 dark:text-white">
                Modrinth Pack (.mrpack)
              </p>
              <p class="text-xs text-gray-500 dark:text-neutral-400">
                Import into Prism Launcher or Modrinth App
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button
              class="inline-flex items-center gap-1 px-2 py-1.5 rounded-lg border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
              @click="copyUrl('mrpack')"
            >
              <Link class="w-4 h-4" />
              {{ copiedFormat === "mrpack" ? "Copied!" : "Copy link" }}
            </button>
            <a
              :href="downloadUrl('mrpack')"
              class="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-emerald-600 text-white text-sm font-medium hover:bg-emerald-600/90 transition-colors"
            >
              <Download class="w-4 h-4" />
              Download
            </a>
          </div>
        </div>

        <div
          v-if="hasFormat('curseforge')"
          class="flex items-center justify-between p-4 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-200 dark:border-neutral-600"
        >
          <div class="flex items-center gap-3">
            <Flame class="w-8 h-8 text-orange-600 dark:text-orange-400" />
            <div>
              <p class="font-medium text-gray-900 dark:text-white">
                CurseForge Pack
              </p>
              <p class="text-xs text-gray-500 dark:text-neutral-400">
                Import into CurseForge or Prism Launcher
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button
              class="inline-flex items-center gap-1 px-2 py-1.5 rounded-lg border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
              @click="copyUrl('curseforge')"
            >
              <Link class="w-4 h-4" />
              {{ copiedFormat === "curseforge" ? "Copied!" : "Copy link" }}
            </button>
            <a
              :href="downloadUrl('curseforge')"
              class="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-orange-600 text-white text-sm font-medium hover:bg-orange-600/90 transition-colors"
            >
              <Download class="w-4 h-4" />
              Download
            </a>
          </div>
        </div>

        <div
          v-if="hasFormat('prism')"
          class="flex items-center justify-between p-4 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-200 dark:border-neutral-600"
        >
          <div class="flex items-center gap-3">
            <Box class="w-8 h-8 text-indigo-600 dark:text-indigo-400" />
            <div>
              <p class="font-medium text-gray-900 dark:text-white">
                Prism / MultiMC Instance
              </p>
              <p class="text-xs text-gray-500 dark:text-neutral-400">
                Import into Prism Launcher or MultiMC
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button
              class="inline-flex items-center gap-1 px-2 py-1.5 rounded-lg border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
              @click="copyUrl('prism')"
            >
              <Link class="w-4 h-4" />
              {{ copiedFormat === "prism" ? "Copied!" : "Copy link" }}
            </button>
            <a
              :href="downloadUrl('prism')"
              class="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-indigo-600 text-white text-sm font-medium hover:bg-indigo-600/90 transition-colors"
            >
              <Download class="w-4 h-4" />
              Download
            </a>
          </div>
        </div>
      </div>

      <div class="flex items-center justify-center gap-4 mt-4">
        <div class="text-center text-xs text-gray-400 dark:text-neutral-500">
          <p>Link expires: {{ expiresText }}</p>
          <p class="mt-1">
            Generated:
            {{
              info?.createdAt ? new Date(info.createdAt).toLocaleString() : ""
            }}
          </p>
        </div>
      </div>
      <div v-if="isAuthenticated" class="flex justify-center gap-3 mt-4">
        <button
          class="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
          @click="goBack()"
        >
          <ArrowLeft class="w-4 h-4" />
          Go Back
        </button>
        <a
          href="/"
          class="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium transition-colors"
        >
          <Home class="w-4 h-4" />
          Go Home
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  noSidebar: true,
});
import {
  Loader2,
  XCircle,
  FileArchive,
  FileBox,
  Download,
  Link,
  Flame,
  Box,
  Home,
  ArrowLeft,
} from "lucide-vue-next";

interface ArchiveInfoResponse {
  token: string;
  serverName: string;
  expiresAt: string;
  createdAt: string;
  formats: string[];
}

const route = useRoute();
const router = useRouter();
const token = route.params.token as string;
const { isAuthenticated } = useAuth();

const loading = ref(true);
const error = ref("");
const info = ref<ArchiveInfoResponse | null>(null);

usePageTitle(() => info.value?.serverName || "Modpack");
const copiedFormat = ref<string | null>(null);

function copyUrl(format: string) {
  const url = downloadUrl(format);
  const absoluteUrl = url.startsWith("http")
    ? url
    : `${window.location.origin}${url}`;
  navigator.clipboard.writeText(absoluteUrl).then(() => {
    copiedFormat.value = format;
    setTimeout(() => (copiedFormat.value = null), 2000);
  });
}

function downloadUrl(format: string): string {
  const base = useApiBase();
  return `${base}/client-archive/${token}?format=${format}`;
}

function hasFormat(fmt: string): boolean {
  return info.value?.formats?.includes(fmt) ?? false;
}

const now = ref(new Date());

function goBack() {
  if (typeof window !== "undefined" && window.history.length > 1) {
    window.history.back();
  } else {
    navigateTo("/");
  }
}

const expiresText = computed(() => {
  if (!info.value?.expiresAt) return "";
  const date = new Date(info.value.expiresAt);
  const diff = date.getTime() - now.value.getTime();
  if (diff <= 0) return "Expired";
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  if (hours > 0) return `${hours}h ${minutes}m remaining`;
  return `${minutes}m remaining`;
});

let timer: ReturnType<typeof setInterval> | null = null;

onMounted(async () => {
  try {
    const res = await $fetch<ArchiveInfoResponse>(
      `/client-archive/${token}/info`,
      {
        baseURL: useApiBase(),
      },
    );
    info.value = res;
  } catch (err: any) {
    const statusCode = err?.statusCode || err?.response?.status;
    if (statusCode === 404) {
      error.value = "This archive link has expired or does not exist.";
    } else if (statusCode === 500) {
      error.value = "Server error, please try again later.";
    } else {
      error.value = "Archive not found or expired.";
    }
  } finally {
    loading.value = false;
  }
  timer = setInterval(() => {
    now.value = new Date();
  }, 30000);
});

onBeforeUnmount(() => {
  if (timer) clearInterval(timer);
});
</script>
