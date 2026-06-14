<script setup lang="ts">
import { ref, computed } from "vue";

interface CrashReport {
  name: string;
  size: number;
  modifiedAt: string;
}

usePageTitle("Crash Reports");
definePageMeta({ middleware: "auth" });

const route = useRoute();
const serverId = computed(() => route.params.serverId as string);
const { show: showToast } = useToast();

const reports = ref<CrashReport[]>([]);
const loading = ref(false);
const selected = ref<string | null>(null);
const content = ref("");
const contentLoading = ref(false);
const deleteConfirm = ref<string | null>(null);
const search = ref("");

const filteredReports = computed(() => {
  const q = search.value.trim().toLowerCase();
  if (!q) return reports.value;
  return reports.value.filter((r) => r.name.toLowerCase().includes(q));
});

async function fetchReports() {
  loading.value = true;
  try {
    const res = await $fetch<{ reports: CrashReport[] }>(
      `/api/servers/${serverId.value}/crash-reports`,
      { baseURL: useApiBase(), credentials: "include" },
    );
    reports.value = res.reports || [];
    const first = reports.value.at(0);
    if (first && !selected.value) {
      await viewReport(first.name);
    }
  } catch (err: any) {
    showToast("error", err?.message || "Failed to load crash reports");
  } finally {
    loading.value = false;
  }
}

async function viewReport(name: string) {
  selected.value = name;
  contentLoading.value = true;
  content.value = "";
  try {
    const text = await $fetch<string>(
      `/api/servers/${serverId.value}/crash-reports/${encodeURIComponent(name)}`,
      { baseURL: useApiBase(), credentials: "include" },
    );
    content.value = text;
  } catch (err: any) {
    showToast("error", err?.message || "Failed to read report");
  } finally {
    contentLoading.value = false;
  }
}

async function deleteReport(name: string) {
  deleteConfirm.value = null;
  try {
    await $fetch(
      `/api/servers/${serverId.value}/crash-reports/${encodeURIComponent(name)}`,
      {
        baseURL: useApiBase(),
        credentials: "include",
        method: "DELETE",
      },
    );
    showToast("success", "Deleted", { description: name });
    if (selected.value === name) {
      selected.value = null;
      content.value = "";
    }
    await fetchReports();
  } catch (err: any) {
    showToast("error", err?.message || "Failed to delete report");
  }
}

async function copyContent() {
  try {
    await navigator.clipboard.writeText(content.value);
    showToast("success", "Copied to clipboard");
  } catch {
    showToast("error", "Failed to copy");
  }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleString();
}

fetchReports();
</script>

<template>
  <div class="p-6">
    <div class="space-y-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
        Crash Reports
      </h1>

      <div class="flex items-center gap-3 flex-wrap">
        <input
          v-model="search"
          type="text"
          placeholder="Search reports..."
          class="w-full md:w-80 px-3 py-2 rounded-xl bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-sm text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-neutral-500 focus:ring-2 focus:ring-primary focus:outline-none"
        />
      </div>

      <div
        class="flex flex-col md:flex-row gap-6 h-[calc(100vh-14rem)] min-h-[500px]"
      >
        <!-- List -->
        <div
          class="w-full md:w-[30%] bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden flex flex-col min-h-0"
        >
          <div
            class="px-4 py-2 bg-gray-50 dark:bg-neutral-900 border-b border-gray-200 dark:border-neutral-700 text-xs text-gray-500 dark:text-neutral-400"
          >
            {{ filteredReports.length }} reports
          </div>
          <div class="flex-1 min-h-0 overflow-auto">
            <div
              v-if="loading"
              class="p-4 text-sm text-gray-500 dark:text-neutral-400"
            >
              Loading crash reports...
            </div>
            <div
              v-else-if="reports.length === 0"
              class="p-4 text-sm text-gray-500 dark:text-neutral-400"
            >
              No crash reports found.
            </div>
            <div
              v-else-if="filteredReports.length === 0"
              class="p-4 text-sm text-gray-500 dark:text-neutral-400"
            >
              No matches.
            </div>
            <table v-else class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-neutral-900 sticky top-0">
                <tr class="text-left text-gray-500 dark:text-neutral-400">
                  <th class="px-3 py-2">Name</th>
                  <th class="px-3 py-2">Size</th>
                  <th class="px-3 py-2">Modified</th>
                  <th class="px-3 py-2 w-16"></th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="report in filteredReports"
                  :key="report.name"
                  class="border-t border-gray-100 dark:border-neutral-800 hover:bg-gray-50 dark:hover:bg-neutral-900 cursor-pointer"
                  :class="
                    selected === report.name
                      ? 'bg-blue-50 dark:bg-blue-900/20'
                      : ''
                  "
                  @click="viewReport(report.name)"
                >
                  <td
                    class="px-3 py-2 text-gray-900 dark:text-white truncate max-w-xs"
                  >
                    {{ report.name }}
                  </td>
                  <td
                    class="px-3 py-2 text-gray-500 dark:text-neutral-400 whitespace-nowrap"
                  >
                    {{ formatSize(report.size) }}
                  </td>
                  <td
                    class="px-3 py-2 text-gray-500 dark:text-neutral-400 whitespace-nowrap"
                  >
                    {{ formatDate(report.modifiedAt) }}
                  </td>
                  <td class="px-3 py-2">
                    <button
                      class="text-red-500 hover:text-red-600 p-1 rounded transition-colors"
                      title="Delete"
                      @click.stop="deleteConfirm = report.name"
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        class="w-4 h-4"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                        />
                      </svg>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Viewer -->
        <div
          class="w-full md:w-[70%] bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden flex flex-col min-h-0"
        >
          <div
            v-if="contentLoading"
            class="p-4 text-sm text-gray-500 dark:text-neutral-400"
          >
            Loading report...
          </div>
          <div
            v-else-if="!selected"
            class="p-4 text-sm text-gray-500 dark:text-neutral-400"
          >
            Select a report to view its contents.
          </div>
          <template v-else>
            <div
              class="flex items-center justify-between px-4 py-2 bg-gray-50 dark:bg-neutral-900 border-b border-gray-200 dark:border-neutral-700"
            >
              <span
                class="font-mono text-sm text-gray-900 dark:text-white truncate"
              >
                {{ selected }}
              </span>
              <div class="flex items-center gap-2">
                <button
                  class="text-gray-600 dark:text-neutral-300 hover:text-gray-900 dark:hover:text-white text-xs px-2 py-1 rounded-lg hover:bg-gray-100 dark:hover:bg-neutral-800 transition-colors"
                  @click="copyContent"
                >
                  Copy
                </button>
                <button
                  class="text-gray-600 dark:text-neutral-300 hover:text-gray-900 dark:hover:text-white text-xs px-2 py-1 rounded-lg hover:bg-gray-100 dark:hover:bg-neutral-800 transition-colors"
                  @click="
                    selected = null;
                    content = '';
                  "
                >
                  Close
                </button>
              </div>
            </div>
            <pre
              class="flex-1 min-h-0 overflow-auto p-4 text-xs font-mono whitespace-pre-wrap bg-white dark:bg-neutral-800 text-gray-900 dark:text-neutral-200"
              >{{ content }}</pre
            >
          </template>
        </div>
      </div>
    </div>

    <!-- Delete confirmation modal -->
    <div
      v-if="deleteConfirm"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      @click.self="deleteConfirm = null"
    >
      <div
        class="bg-white dark:bg-neutral-800 rounded-2xl p-6 max-w-sm w-full mx-4 shadow-lg border border-gray-200 dark:border-neutral-700"
      >
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
          Delete crash report?
        </h3>
        <p class="text-gray-600 dark:text-neutral-400 mb-4">
          Are you sure you want to delete
          <code
            class="font-mono bg-gray-100 dark:bg-neutral-700 px-1 rounded text-gray-900 dark:text-white"
            >{{ deleteConfirm }}</code
          >?
        </p>
        <div class="flex justify-end gap-3">
          <button
            class="px-4 py-2 rounded-xl text-sm text-gray-600 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors"
            @click="deleteConfirm = null"
          >
            Cancel
          </button>
          <button
            class="px-4 py-2 rounded-xl text-sm bg-red-500 text-white hover:bg-red-600 shadow-sm hover:shadow-md active:scale-95 transition-all"
            @click="deleteReport(deleteConfirm)"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
