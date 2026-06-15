<script setup lang="ts">
import { Copy, FolderOpen, Globe, Zap } from "lucide-vue-next";
import type { ServerDisk } from "~/composables/useServerDisk";

const props = defineProps<{
  serverId: string;
  hostPath?: string;
  serverDisk?: ServerDisk | null;
}>();

const { show: showToast } = useToast();

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text);
    showToast("info", "Path copied to clipboard");
  } catch {
    showToast("error", "Failed to copy path");
  }
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
}
</script>

<template>
  <div class="space-y-2">
    <p
      class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
    >
      Storage
    </p>
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
      <!-- Path (host) -->
      <div
        class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 col-span-full"
      >
        <div
          class="w-9 h-9 shrink-0 rounded-lg bg-gray-100 dark:bg-neutral-600/50 flex items-center justify-center text-gray-600 dark:text-gray-400"
        >
          <FolderOpen class="w-4 h-4" />
        </div>
        <div class="min-w-0 flex-1">
          <p
            class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
          >
            Path (host)
          </p>
          <div class="flex items-center gap-2">
            <p
              class="text-sm font-semibold text-gray-900 dark:text-white break-all font-mono"
              :title="hostPath"
            >
              {{ hostPath || "-" }}
            </p>
            <button
              v-if="hostPath"
              class="text-gray-400 hover:text-primary transition-colors shrink-0"
              :title="'Copy path'"
              @click="copyToClipboard(hostPath)"
            >
              <Copy class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>
      <!-- Total Volume -->
      <div
        class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
      >
        <div
          class="w-9 h-9 shrink-0 rounded-lg bg-gray-100 dark:bg-neutral-600/50 flex items-center justify-center text-gray-600 dark:text-gray-400"
        >
          <FolderOpen class="w-4 h-4" />
        </div>
        <div class="min-w-0 flex-1">
          <p
            class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
          >
            Total Volume
          </p>
          <p class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ serverDisk ? formatBytes(serverDisk.totalBytes) : "—" }}
          </p>
        </div>
      </div>
      <!-- World -->
      <div
        class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
      >
        <div
          class="w-9 h-9 shrink-0 rounded-lg bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-600 dark:text-emerald-400"
        >
          <Globe class="w-4 h-4" />
        </div>
        <div class="min-w-0 flex-1">
          <p
            class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
          >
            World
          </p>
          <p class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ serverDisk ? formatBytes(serverDisk.worldBytes) : "—" }}
          </p>
        </div>
      </div>
      <!-- Nether -->
      <div
        class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
      >
        <div
          class="w-9 h-9 shrink-0 rounded-lg bg-red-100 dark:bg-red-900/30 flex items-center justify-center text-red-600 dark:text-red-400"
        >
          <Zap class="w-4 h-4" />
        </div>
        <div class="min-w-0 flex-1">
          <p
            class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
          >
            Nether
          </p>
          <p class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ serverDisk ? formatBytes(serverDisk.worldNetherBytes) : "—" }}
          </p>
        </div>
      </div>
      <!-- The End -->
      <div
        class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
      >
        <div
          class="w-9 h-9 shrink-0 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center text-purple-600 dark:text-purple-400"
        >
          <Zap class="w-4 h-4" />
        </div>
        <div class="min-w-0 flex-1">
          <p
            class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
          >
            The End
          </p>
          <p class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ serverDisk ? formatBytes(serverDisk.worldEndBytes) : "—" }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
