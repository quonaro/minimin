<template>
  <div class="p-6 h-[calc(100vh-4rem)] flex flex-col gap-4">
    <div class="flex items-center justify-between flex-wrap gap-3">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Files</h1>
      <div class="flex items-center gap-2 flex-wrap">
        <button
          class="px-3 py-1.5 rounded-lg border border-gray-300 dark:border-neutral-700 text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-800"
          :disabled="!canGoUp"
          @click="goUp"
        >
          Up
        </button>
        <button
          class="px-3 py-1.5 rounded-lg border border-gray-300 dark:border-neutral-700 text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-800"
          @click="refreshList"
        >
          Refresh
        </button>
        <button
          class="px-3 py-1.5 rounded-lg bg-primary text-white text-sm hover:bg-primary/90"
          @click="openCreateDirModal"
        >
          New Folder
        </button>
        <button
          class="px-3 py-1.5 rounded-lg bg-primary text-white text-sm hover:bg-primary/90"
          @click="openCreateFileModal"
        >
          New File
        </button>
        <button
          class="px-3 py-1.5 rounded-lg bg-primary text-white text-sm hover:bg-primary/90"
          @click="triggerUpload"
        >
          Upload
        </button>
      </div>
    </div>

    <div class="flex items-center gap-2 flex-wrap text-sm">
      <button class="text-primary hover:underline" @click="navigateToPath('')">
        /
      </button>
      <template v-for="(crumb, idx) in breadcrumbs" :key="crumb.path || 'root'">
        <span class="text-gray-400 dark:text-neutral-600">/</span>
        <button
          class="text-primary hover:underline"
          @click="navigateToPath(crumb.path)"
        >
          {{ crumb.label }}
        </button>
        <span
          v-if="idx === breadcrumbs.length - 1"
          class="text-xs text-gray-500 dark:text-neutral-400 ml-1"
        >
          (current)
        </span>
      </template>
    </div>

    <input
      v-model="search"
      type="text"
      placeholder="Search in current folder..."
      class="w-full md:w-80 px-3 py-2 rounded-lg bg-white dark:bg-neutral-800 border border-gray-300 dark:border-neutral-700 text-sm text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-neutral-500"
    />

    <div class="grid grid-cols-1 xl:grid-cols-[1fr_440px] gap-4 flex-1 min-h-0">
      <div
        class="border border-gray-200 dark:border-neutral-800 rounded-xl overflow-hidden flex flex-col min-h-0"
      >
        <div
          class="px-4 py-2 bg-gray-50 dark:bg-neutral-900 border-b border-gray-200 dark:border-neutral-800 text-xs text-gray-500 dark:text-neutral-400"
        >
          {{ currentPath || "/" }} · {{ filteredEntries.length }} items
        </div>
        <div class="flex-1 min-h-0 overflow-auto">
          <div
            v-if="listLoading"
            class="p-4 text-sm text-gray-500 dark:text-neutral-400"
          >
            Loading files...
          </div>
          <div v-else-if="listError" class="p-4 text-sm text-red-500">
            {{ listError }}
          </div>
          <div
            v-else-if="filteredEntries.length === 0"
            class="p-4 text-sm text-gray-500 dark:text-neutral-400"
          >
            Folder is empty.
          </div>
          <table v-else class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-neutral-900 sticky top-0">
              <tr class="text-left text-gray-500 dark:text-neutral-400">
                <th class="px-3 py-2">Name</th>
                <th class="px-3 py-2">Size</th>
                <th class="px-3 py-2">Modified</th>
                <th class="px-3 py-2 w-[260px]">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="entry in filteredEntries"
                :key="entry.path"
                class="border-t border-gray-100 dark:border-neutral-800 hover:bg-gray-50 dark:hover:bg-neutral-900"
              >
                <td class="px-3 py-2">
                  <button
                    class="text-left hover:underline"
                    :class="
                      entry.isDir
                        ? 'text-primary font-medium'
                        : 'text-gray-900 dark:text-white'
                    "
                    @click="openEntry(entry)"
                  >
                    {{ entry.isDir ? "📁" : "📄" }} {{ entry.name }}
                  </button>
                </td>
                <td class="px-3 py-2 text-gray-500 dark:text-neutral-400">
                  {{ entry.isDir ? "—" : formatSize(entry.size) }}
                </td>
                <td class="px-3 py-2 text-gray-500 dark:text-neutral-400">
                  {{ formatDate(entry.modifiedAt) }}
                </td>
                <td class="px-3 py-2">
                  <div class="flex items-center gap-1 flex-wrap">
                    <button
                      v-if="!entry.isDir"
                      class="px-2 py-1 rounded border border-gray-300 dark:border-neutral-700 text-xs text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-800"
                      @click="openFile(entry.path)"
                    >
                      Open
                    </button>
                    <button
                      v-if="!entry.isDir"
                      class="px-2 py-1 rounded border border-gray-300 dark:border-neutral-700 text-xs text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-800"
                      @click="downloadFile(entry.path)"
                    >
                      Download
                    </button>
                    <button
                      class="px-2 py-1 rounded border border-gray-300 dark:border-neutral-700 text-xs text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-800"
                      @click="openRenameModal(entry.path)"
                    >
                      Rename/Move
                    </button>
                    <button
                      class="px-2 py-1 rounded border border-red-300 text-red-600 text-xs hover:bg-red-50 dark:hover:bg-red-900/20"
                      @click="openDeleteModal(entry.path)"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div
        class="border border-gray-200 dark:border-neutral-800 rounded-xl overflow-hidden flex flex-col min-h-0"
      >
        <div
          class="px-4 py-2 bg-gray-50 dark:bg-neutral-900 border-b border-gray-200 dark:border-neutral-800 text-sm font-medium text-gray-900 dark:text-white"
        >
          Editor
        </div>
        <div class="p-4 flex-1 min-h-0 flex flex-col gap-3">
          <div
            v-if="!openedFilePath"
            class="text-sm text-gray-500 dark:text-neutral-400"
          >
            Select a text file to view or edit.
          </div>
          <template v-else>
            <div class="text-xs text-gray-500 dark:text-neutral-400 break-all">
              {{ openedFilePath }}
            </div>
            <div
              v-if="editorLoading"
              class="text-sm text-gray-500 dark:text-neutral-400"
            >
              Loading file...
            </div>
            <div v-else-if="editorError" class="text-sm text-red-500">
              {{ editorError }}
            </div>
            <div
              v-else-if="openedFileState?.tooLarge"
              class="text-sm text-orange-500"
            >
              File is too large for inline editing (limit:
              {{ formatSize(openedFileState.maxEditableBytes || 0) }}).
            </div>
            <div
              v-else-if="openedFileState?.isBinary"
              class="text-sm text-orange-500"
            >
              Binary file cannot be edited inline.
            </div>
            <template v-else>
              <textarea
                v-model="editorContent"
                class="flex-1 min-h-0 w-full rounded-lg border border-gray-300 dark:border-neutral-700 bg-white dark:bg-neutral-900 p-3 font-mono text-xs text-gray-900 dark:text-neutral-100"
              />
              <div class="flex justify-end">
                <button
                  class="px-3 py-1.5 rounded-lg bg-primary text-white text-sm hover:bg-primary/90"
                  :disabled="saveLoading"
                  @click="openSaveModal"
                >
                  {{ saveLoading ? "Saving..." : "Save changes" }}
                </button>
              </div>
            </template>
          </template>
        </div>
      </div>
    </div>

    <input
      ref="uploadInput"
      type="file"
      class="hidden"
      @change="onUploadSelected"
    />

    <div
      v-if="warningModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    >
      <div
        class="bg-white dark:bg-neutral-800 rounded-2xl border border-gray-200 dark:border-neutral-700 shadow-xl w-full max-w-lg p-6"
      >
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
          Warning
        </h2>
        <p class="text-sm text-gray-700 dark:text-neutral-300 leading-relaxed">
          Любое случайное изменение файлов может повредить сохранение, мир или
          конфигурацию сервера. Продолжайте только если понимаете риск.
        </p>
        <div class="mt-5 flex justify-end">
          <button
            class="px-4 py-2 rounded-lg bg-red-600 text-white hover:bg-red-700"
            @click="warningModalOpen = false"
          >
            Я понимаю риск
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="confirmModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
      @click.self="closeConfirmModal"
    >
      <div
        class="bg-white dark:bg-neutral-800 rounded-2xl border border-gray-200 dark:border-neutral-700 shadow-xl w-full max-w-lg p-6"
      >
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
          {{ confirmTitle }}
        </h2>
        <p class="text-sm text-gray-700 dark:text-neutral-300 break-words">
          {{ confirmMessage }}
        </p>
        <div v-if="confirmInputLabel" class="mt-3">
          <label class="text-xs text-gray-500 dark:text-neutral-400">{{
            confirmInputLabel
          }}</label>
          <input
            v-model="confirmInputValue"
            type="text"
            class="mt-1 w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-700 bg-white dark:bg-neutral-900 text-sm text-gray-900 dark:text-neutral-100"
          />
        </div>
        <div class="mt-5 flex justify-end gap-2">
          <button
            class="px-3 py-1.5 rounded-lg border border-gray-300 dark:border-neutral-700 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm text-gray-700 dark:text-neutral-200"
            @click="closeConfirmModal"
          >
            Cancel
          </button>
          <button
            class="px-3 py-1.5 rounded-lg text-white text-sm"
            :class="
              confirmDanger
                ? 'bg-red-600 hover:bg-red-700'
                : 'bg-primary hover:bg-primary/90'
            "
            :disabled="confirmLoading"
            @click="confirmAction"
          >
            {{ confirmLoading ? "Please wait..." : confirmButtonText }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "auth" });

interface FileEntry {
  name: string;
  path: string;
  isDir: boolean;
  size: number;
  modifiedAt: string;
}

interface ReadFileResponse {
  path: string;
  size: number;
  content: string;
  isBinary: boolean;
  tooLarge: boolean;
  maxEditableBytes: number;
}

const route = useRoute();
const { agentId } = useCurrentAgent();
const { show: showToast } = useToast();
const serverId = route.params.serverId as string;

const warningModalOpen = ref(true);
const uploadInput = ref<HTMLInputElement | null>(null);

const currentPath = ref("");
const search = ref("");
const listLoading = ref(false);
const listError = ref("");
const entries = ref<FileEntry[]>([]);

const openedFilePath = ref("");
const openedFileState = ref<ReadFileResponse | null>(null);
const editorContent = ref("");
const editorLoading = ref(false);
const editorError = ref("");
const saveLoading = ref(false);

const confirmModalOpen = ref(false);
const confirmLoading = ref(false);
const confirmTitle = ref("");
const confirmMessage = ref("");
const confirmInputLabel = ref("");
const confirmInputValue = ref("");
const confirmButtonText = ref("Confirm");
const confirmDanger = ref(false);
let onConfirm: (() => Promise<void>) | null = null;

const canGoUp = computed(() => !!currentPath.value);

const breadcrumbs = computed(() => {
  if (!currentPath.value) return [];
  const parts = currentPath.value.split("/").filter(Boolean);
  return parts.map((label, idx) => ({
    label,
    path: parts.slice(0, idx + 1).join("/"),
  }));
});

const filteredEntries = computed(() => {
  const q = search.value.trim().toLowerCase();
  if (!q) return entries.value;
  return entries.value.filter((entry) => entry.name.toLowerCase().includes(q));
});

function normalizePath(path: string): string {
  return path.split("/").filter(Boolean).join("/");
}

function buildChildPath(basePath: string, childName: string): string {
  return normalizePath([basePath, childName].filter(Boolean).join("/"));
}

function parentPath(path: string): string {
  const parts = normalizePath(path).split("/").filter(Boolean);
  parts.pop();
  return parts.join("/");
}

function formatSize(size: number): string {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  if (size < 1024 * 1024 * 1024)
    return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  return `${(size / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleString();
}

async function refreshList() {
  listLoading.value = true;
  listError.value = "";
  try {
    const res = await $fetch<{ path: string; entries: FileEntry[] }>(
      `/agent/${agentId.value}/servers/${serverId}/files`,
      {
        baseURL: useApiBase(),
        credentials: "include",
        query: { path: currentPath.value || undefined },
      },
    );
    entries.value = res.entries || [];
  } catch (err: any) {
    listError.value =
      err?.data?.detail || err?.message || "Failed to load files";
  } finally {
    listLoading.value = false;
  }
}

async function navigateToPath(path: string) {
  currentPath.value = normalizePath(path);
  await refreshList();
}

async function goUp() {
  await navigateToPath(parentPath(currentPath.value));
}

async function openEntry(entry: FileEntry) {
  if (entry.isDir) {
    await navigateToPath(entry.path);
    return;
  }
  await openFile(entry.path);
}

async function openFile(path: string) {
  openedFilePath.value = path;
  editorLoading.value = true;
  editorError.value = "";
  openedFileState.value = null;
  editorContent.value = "";
  try {
    const res = await $fetch<ReadFileResponse>(
      `/agent/${agentId.value}/servers/${serverId}/file`,
      {
        baseURL: useApiBase(),
        credentials: "include",
        query: { path },
      },
    );
    openedFileState.value = res;
    editorContent.value = res.content || "";
  } catch (err: any) {
    editorError.value =
      err?.data?.detail || err?.message || "Failed to open file";
  } finally {
    editorLoading.value = false;
  }
}

function downloadFile(path: string) {
  const qs = new URLSearchParams({ path }).toString();
  window.open(
    `${useApiBase()}/agent/${agentId.value}/servers/${serverId}/files/download?${qs}`,
    "_blank",
  );
}

function triggerUpload() {
  uploadInput.value?.click();
}

async function onUploadSelected(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0] ?? null;
  input.value = "";
  if (!file) return;
  openConfirmModal({
    title: "Confirm upload",
    message: `Upload file ${file.name} to ${currentPath.value || "/"}?`,
    buttonText: "Upload",
    danger: false,
    action: async () => {
      const formData = new FormData();
      formData.append("file", file);
      await $fetch(`/agent/${agentId.value}/servers/${serverId}/files/upload`, {
        baseURL: useApiBase(),
        credentials: "include",
        method: "POST",
        query: { path: currentPath.value || undefined },
        body: formData,
      });
      showToast("success", "Upload completed", { description: file.name });
      await refreshList();
    },
  });
}

function openCreateDirModal() {
  openConfirmModal({
    title: "Create folder",
    message: `Create folder in ${currentPath.value || "/"}`,
    inputLabel: "Folder name",
    buttonText: "Create",
    danger: false,
    action: async (value) => {
      const name = value.trim();
      if (!name) throw new Error("Folder name is required");
      const path = buildChildPath(currentPath.value, name);
      await $fetch(`/agent/${agentId.value}/servers/${serverId}/files/mkdir`, {
        baseURL: useApiBase(),
        credentials: "include",
        method: "POST",
        body: { path },
      });
      showToast("success", "Folder created", { description: path });
      await refreshList();
    },
  });
}

function openCreateFileModal() {
  openConfirmModal({
    title: "Create file",
    message: `Create file in ${currentPath.value || "/"}`,
    inputLabel: "File name",
    buttonText: "Create",
    danger: false,
    action: async (value) => {
      const name = value.trim();
      if (!name) throw new Error("File name is required");
      const path = buildChildPath(currentPath.value, name);
      await $fetch(`/agent/${agentId.value}/servers/${serverId}/files/create`, {
        baseURL: useApiBase(),
        credentials: "include",
        method: "POST",
        body: { path, content: "" },
      });
      showToast("success", "File created", { description: path });
      await refreshList();
    },
  });
}

function openRenameModal(path: string) {
  openConfirmModal({
    title: "Rename or move",
    message: `Current path: ${path}`,
    inputLabel: "New relative path",
    initialValue: path,
    buttonText: "Apply",
    danger: false,
    action: async (value) => {
      const target = normalizePath(value);
      if (!target) throw new Error("Target path is required");
      await $fetch(`/agent/${agentId.value}/servers/${serverId}/files/move`, {
        baseURL: useApiBase(),
        credentials: "include",
        method: "POST",
        body: { fromPath: path, toPath: target },
      });
      if (openedFilePath.value === path) {
        openedFilePath.value = target;
      }
      showToast("success", "Path updated", {
        description: `${path} → ${target}`,
      });
      await refreshList();
    },
  });
}

function openDeleteModal(path: string) {
  openConfirmModal({
    title: "Delete",
    message: `Delete ${path}? This action cannot be undone.`,
    buttonText: "Delete",
    danger: true,
    action: async () => {
      await $fetch(`/agent/${agentId.value}/servers/${serverId}/files`, {
        baseURL: useApiBase(),
        credentials: "include",
        method: "DELETE",
        query: { path },
      });
      if (openedFilePath.value === path) {
        openedFilePath.value = "";
        openedFileState.value = null;
        editorContent.value = "";
      }
      showToast("success", "Deleted", { description: path });
      await refreshList();
    },
  });
}

function openSaveModal() {
  if (!openedFilePath.value) return;
  openConfirmModal({
    title: "Save file",
    message: `Save changes to ${openedFilePath.value}?`,
    buttonText: "Save",
    danger: true,
    action: async () => {
      saveLoading.value = true;
      try {
        await $fetch(`/agent/${agentId.value}/servers/${serverId}/file`, {
          baseURL: useApiBase(),
          credentials: "include",
          method: "PUT",
          body: {
            path: openedFilePath.value,
            content: editorContent.value,
          },
        });
        showToast("success", "File saved", {
          description: openedFilePath.value,
        });
        await refreshList();
      } finally {
        saveLoading.value = false;
      }
    },
  });
}

function openConfirmModal(options: {
  title: string;
  message: string;
  buttonText: string;
  danger: boolean;
  inputLabel?: string;
  initialValue?: string;
  action: (value: string) => Promise<void>;
}) {
  confirmTitle.value = options.title;
  confirmMessage.value = options.message;
  confirmButtonText.value = options.buttonText;
  confirmDanger.value = options.danger;
  confirmInputLabel.value = options.inputLabel || "";
  confirmInputValue.value = options.initialValue || "";
  onConfirm = () => options.action(confirmInputValue.value);
  confirmModalOpen.value = true;
}

function closeConfirmModal() {
  if (confirmLoading.value) return;
  forceCloseConfirmModal();
}

function forceCloseConfirmModal() {
  confirmModalOpen.value = false;
  confirmInputLabel.value = "";
  confirmInputValue.value = "";
  onConfirm = null;
}

async function confirmAction() {
  if (!onConfirm) return;
  confirmLoading.value = true;
  try {
    await onConfirm();
    forceCloseConfirmModal();
  } catch (err: any) {
    const msg = err?.data?.detail || err?.message || "Action failed";
    showToast("error", "Operation failed", { description: msg });
  } finally {
    confirmLoading.value = false;
  }
}

onMounted(async () => {
  await refreshList();
});
</script>
