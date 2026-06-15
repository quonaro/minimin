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

type ContextTargetType = "file" | "dir" | "empty";

interface ConfirmOptions {
  title: string;
  message: string;
  buttonText: string;
  danger: boolean;
  inputLabel?: string;
  initialValue?: string;
  action: (value: string) => Promise<void>;
}

function getErrorMessage(err: unknown, fallback: string): string {
  if (typeof err === "string") return err;
  if (err && typeof err === "object") {
    const e = err as {
      data?: { detail?: string };
      message?: string;
    };
    if (e.data?.detail) return e.data.detail;
    if (e.message) return e.message;
  }
  return fallback;
}

export function useFilesExplorer() {
  const route = useRoute();
  const { show: showToast } = useToast();
  const serverId = route.params.serverId as string;

  const uploadInput = ref<HTMLInputElement | null>(null);

  const currentPath = ref("");
  const search = ref("");
  const listLoading = ref(false);
  const listError = ref("");
  const entries = ref<FileEntry[]>([]);
  const selectedEntry = ref<FileEntry | null>(null);

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

  const contextMenuOpen = ref(false);
  const contextMenuX = ref(0);
  const contextMenuY = ref(0);
  const contextTargetType = ref<ContextTargetType>("empty");
  const contextTargetPath = ref("");

  const dragSourcePath = ref("");
  const dragSourceIsDir = ref(false);
  const dragOverDirPath = ref("");
  const dragOverBreadcrumbPath = ref<string | null>(null);

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

  const displayEntries = computed(() => {
    const hasParent = currentPath.value !== "";
    const hasSearch = search.value.trim() !== "";
    if (!hasParent || hasSearch) return filteredEntries.value;

    const parentEntry: FileEntry = {
      name: "..",
      path: parentPath(currentPath.value),
      isDir: true,
      size: 0,
      modifiedAt: "",
    };
    return [parentEntry, ...filteredEntries.value];
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

  function closeContextMenu() {
    contextMenuOpen.value = false;
  }

  async function refreshList() {
    listLoading.value = true;
    listError.value = "";
    try {
      const res = await $fetch<{ path: string; entries: FileEntry[] }>(
        `/servers/${serverId}/files`,
        {
          baseURL: useApiBase(),
          credentials: "include",
          query: { path: currentPath.value || undefined },
        },
      );
      entries.value = res.entries || [];
    } catch (err: unknown) {
      listError.value = getErrorMessage(err, "Failed to load files");
    } finally {
      listLoading.value = false;
    }
  }

  async function navigateToPath(path: string) {
    closeContextMenu();
    clearSelection();
    currentPath.value = normalizePath(path);
    await refreshList();
  }

  function selectEntry(entry: FileEntry) {
    selectedEntry.value = entry;
    if (!entry.isDir) {
      openFile(entry.path);
    }
  }

  function clearSelection() {
    selectedEntry.value = null;
  }

  async function openEntry(entry: FileEntry) {
    if (entry.isDir) {
      await navigateToPath(entry.path);
      return;
    }
    await openFile(entry.path);
  }

  async function openFile(path: string) {
    closeContextMenu();
    openedFilePath.value = path;
    editorLoading.value = true;
    editorError.value = "";
    openedFileState.value = null;
    editorContent.value = "";
    try {
      const res = await $fetch<ReadFileResponse>(
        `/servers/${serverId}/file`,
        {
          baseURL: useApiBase(),
          credentials: "include",
          query: { path },
        },
      );
      openedFileState.value = res;
      editorContent.value = res.content || "";
    } catch (err: unknown) {
      editorError.value = getErrorMessage(err, "Failed to open file");
    } finally {
      editorLoading.value = false;
    }
  }

  function downloadFile(path: string) {
    const qs = new URLSearchParams({ path }).toString();
    window.open(
      `${useApiBase()}/servers/${serverId}/files/download?${qs}`,
      "_blank",
    );
  }

  function triggerUpload() {
    closeContextMenu();
    uploadInput.value?.click();
  }

  function openContextMenu(event: MouseEvent, type: ContextTargetType, path = "") {
    event.preventDefault();
    const menuWidth = 220;
    const menuHeight = 220;
    contextMenuX.value = Math.min(event.clientX, window.innerWidth - menuWidth);
    contextMenuY.value = Math.min(event.clientY, window.innerHeight - menuHeight);
    contextTargetType.value = type;
    contextTargetPath.value = path;
    contextMenuOpen.value = true;
  }

  function openEntryContextMenu(entry: FileEntry, event: MouseEvent) {
    openContextMenu(event, entry.isDir ? "dir" : "file", entry.path);
  }

  function openEmptyAreaContextMenu(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (target.closest("[data-entry-row='true']")) {
      return;
    }
    openContextMenu(event, "empty");
  }

  async function contextOpenTarget() {
    const targetPath = contextTargetPath.value;
    if (!targetPath) return;
    closeContextMenu();
    if (contextTargetType.value === "dir") {
      await navigateToPath(targetPath);
      return;
    }
    await openFile(targetPath);
  }

  function contextDownloadTarget() {
    const targetPath = contextTargetPath.value;
    if (!targetPath) return;
    closeContextMenu();
    downloadFile(targetPath);
  }

  function contextRenameTarget() {
    const targetPath = contextTargetPath.value;
    if (!targetPath) return;
    closeContextMenu();
    openRenameModal(targetPath);
  }

  function contextDeleteTarget() {
    const targetPath = contextTargetPath.value;
    if (!targetPath) return;
    closeContextMenu();
    openDeleteModal(targetPath);
  }

  function contextCreateFile() {
    closeContextMenu();
    openCreateFileModal();
  }

  function contextCreateFolder() {
    closeContextMenu();
    openCreateDirModal();
  }

  function contextUpload() {
    closeContextMenu();
    triggerUpload();
  }

  async function contextRefresh() {
    closeContextMenu();
    await refreshList();
  }

  function clearDragIndicators() {
    dragOverDirPath.value = "";
    dragOverBreadcrumbPath.value = null;
  }

  function onEntryDragStart(entry: FileEntry, event: DragEvent) {
    dragSourcePath.value = entry.path;
    dragSourceIsDir.value = entry.isDir;
    event.dataTransfer?.setData("text/plain", entry.path);
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = "move";
    }
    closeContextMenu();
  }

  function onEntryDragEnd() {
    dragSourcePath.value = "";
    dragSourceIsDir.value = false;
    clearDragIndicators();
  }

  function canDropToDirectory(destinationDir: string): boolean {
    const source = normalizePath(dragSourcePath.value);
    const destination = normalizePath(destinationDir);
    if (!source) return false;
    if (destination === parentPath(source)) return false;
    if (dragSourceIsDir.value) {
      if (destination === source) return false;
      if (destination.startsWith(`${source}/`)) return false;
    }
    return true;
  }

  function onEntryDragOver(entry: FileEntry, event: DragEvent) {
    if (!entry.isDir || !canDropToDirectory(entry.path)) return;
    event.preventDefault();
    dragOverDirPath.value = entry.path;
    dragOverBreadcrumbPath.value = null;
  }

  function onEntryDragLeave(entry: FileEntry) {
    if (dragOverDirPath.value === entry.path) {
      dragOverDirPath.value = "";
    }
  }

  async function onEntryDrop(entry: FileEntry, event: DragEvent) {
    if (!entry.isDir) return;
    event.preventDefault();
    await moveByDnD(entry.path);
  }

  function onBreadcrumbDragOver(path: string, event: DragEvent) {
    if (!canDropToDirectory(path)) return;
    event.preventDefault();
    dragOverBreadcrumbPath.value = path;
    dragOverDirPath.value = "";
  }

  function onBreadcrumbDragLeave(path: string) {
    if (dragOverBreadcrumbPath.value === path) {
      dragOverBreadcrumbPath.value = null;
    }
  }

  async function onBreadcrumbDrop(path: string, event: DragEvent) {
    event.preventDefault();
    await moveByDnD(path);
  }

  function updateOpenedPathAfterMove(
    sourcePath: string,
    targetPath: string,
    sourceIsDir: boolean,
  ) {
    if (!openedFilePath.value) return;
    const source = normalizePath(sourcePath);
    const target = normalizePath(targetPath);
    const current = normalizePath(openedFilePath.value);

    if (current === source) {
      openedFilePath.value = target;
      return;
    }

    if (sourceIsDir && current.startsWith(`${source}/`)) {
      openedFilePath.value = `${target}${current.slice(source.length)}`;
    }
  }

  async function hasNameConflict(destinationDir: string, destinationName: string) {
    const res = await $fetch<{ entries: FileEntry[] }>(
      `/servers/${serverId}/files`,
      {
        baseURL: useApiBase(),
        credentials: "include",
        query: { path: destinationDir || undefined },
      },
    );

    const sourcePath = normalizePath(dragSourcePath.value);
    const sourceParent = normalizePath(parentPath(sourcePath));
    if (sourceParent === normalizePath(destinationDir)) {
      return false;
    }

    return (res.entries || []).some((entry) => entry.name === destinationName);
  }

  async function moveByDnD(destinationDir: string) {
    clearDragIndicators();
    const sourcePath = normalizePath(dragSourcePath.value);
    const sourceIsDir = dragSourceIsDir.value;
    const destination = normalizePath(destinationDir);
    if (!sourcePath || !canDropToDirectory(destination)) {
      return;
    }

    const itemName = sourcePath.split("/").filter(Boolean).pop();
    if (!itemName) {
      return;
    }
    const targetPath = buildChildPath(destination, itemName);
    if (!targetPath || targetPath === sourcePath) {
      return;
    }

    try {
      const conflict = await hasNameConflict(destination, itemName);
      if (conflict) {
        throw new Error("Target directory already contains item with same name");
      }
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Failed to validate target");
      showToast("error", "Move blocked", { description: msg });
      return;
    }

    openConfirmModal({
      title: "Move",
      message: `Move ${sourcePath} to ${destination || "/"}?`,
      buttonText: "Move",
      danger: false,
      action: async () => {
        await $fetch(`/servers/${serverId}/files/move`, {
          baseURL: useApiBase(),
          credentials: "include",
          method: "POST",
          body: { fromPath: sourcePath, toPath: targetPath },
        });

        updateOpenedPathAfterMove(sourcePath, targetPath, sourceIsDir);
        showToast("success", "Moved", {
          description: `${sourcePath} → ${targetPath}`,
        });
        await refreshList();
      },
    });
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
        await $fetch(`/servers/${serverId}/files/upload`, {
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
        await $fetch(`/servers/${serverId}/files/mkdir`, {
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
        await $fetch(`/servers/${serverId}/files/create`, {
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
        await $fetch(`/servers/${serverId}/files/move`, {
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
        await $fetch(`/servers/${serverId}/files`, {
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
          await $fetch(`/servers/${serverId}/file`, {
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

  function openConfirmModal(options: ConfirmOptions) {
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
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Action failed");
      showToast("error", "Operation failed", { description: msg });
    } finally {
      confirmLoading.value = false;
    }
  }

  function onDocumentKeyDown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      closeContextMenu();
    }
  }

  onMounted(async () => {
    document.addEventListener("keydown", onDocumentKeyDown);
    const rawPath = Array.isArray(route.query.path)
      ? route.query.path[0]
      : route.query.path;
    if (typeof rawPath === "string" && rawPath) {
      currentPath.value = parentPath(normalizePath(rawPath));
    }
    await refreshList();
  });

  onBeforeUnmount(() => {
    document.removeEventListener("keydown", onDocumentKeyDown);
  });

  return {
    uploadInput,
    currentPath,
    search,
    listLoading,
    listError,
    filteredEntries,
    displayEntries,
    selectedEntry,
    selectEntry,
    clearSelection,
    breadcrumbs,
    openedFilePath,
    openedFileState,
    editorContent,
    editorLoading,
    editorError,
    saveLoading,
    confirmModalOpen,
    confirmLoading,
    confirmTitle,
    confirmMessage,
    confirmInputLabel,
    confirmInputValue,
    confirmButtonText,
    confirmDanger,
    contextMenuOpen,
    contextMenuX,
    contextMenuY,
    contextTargetType,
    dragOverDirPath,
    dragOverBreadcrumbPath,
    formatSize,
    formatDate,
    navigateToPath,
    openEntry,
    openFile,
    downloadFile,
    openRenameModal,
    openDeleteModal,
    openSaveModal,
    closeContextMenu,
    openEmptyAreaContextMenu,
    openEntryContextMenu,
    onEntryDragStart,
    onEntryDragEnd,
    onEntryDragOver,
    onEntryDragLeave,
    onEntryDrop,
    onBreadcrumbDragOver,
    onBreadcrumbDragLeave,
    onBreadcrumbDrop,
    onUploadSelected,
    contextOpenTarget,
    contextDownloadTarget,
    contextRenameTarget,
    contextDeleteTarget,
    contextCreateFile,
    contextCreateFolder,
    contextUpload,
    contextRefresh,
    closeConfirmModal,
    confirmAction,
  };
}
