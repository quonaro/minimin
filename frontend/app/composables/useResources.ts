import { debounce } from "~/utils/debounce";
import {
  Upload,
  Link,
  Download,
  Copy,
  Box,
  File,
  Server,
  Monitor,
  Image,
  Sparkles,
  Search,
  X,
  Archive,
} from "lucide-vue-next";
import ConfirmDialog from "~/components/ConfirmDialog.vue";
import JSZip from "jszip";
import { useClientAssetsRefresh } from "~/composables/useClientAssetsRefresh";
import { useClientAssets } from "~/composables/useClientAssets";




const {
  mods,
  loading,
  uploadLoading,
  downloadLoading,
  refresh,
  deleteMod,
  deleteMods,
  uploadFile,
  downloadFromURL,
  toggleMod,
  toggleMods,
} = useMods(serverId);
const {
  mods: clientModList,
  loading: clientLoading,
  uploadLoading: clientUploadLoading,
  downloadLoading: clientDownloadLoading,
  archiveLoading: clientArchiveLoading,
  refresh: refreshClient,
  deleteMod: deleteClientMod,
  deleteMods: deleteClientMods,
  uploadFile: uploadClientFile,
  downloadFromURL: downloadClientFromURL,
  toggleMod: toggleClientMod,
  toggleMods: toggleClientMods,
  moveMod: moveClientMod,
  copyMod: copyClientMod,
  createArchive,
  listArchives,
  deleteArchive,
} = useClientMods(serverId);

const resourcePacks = useClientAssets(serverId, "resourcepacks");
const shaderPacks = useClientAssets(serverId, "shaderpacks");

const fileInput = ref<HTMLInputElement | null>(null);
const showUploadModal = ref(false);
const showLibraryPanel = ref(false);
const showArchiveModal = ref(false);
const modUrl = ref("");
const searchQuery = ref("");
const pendingUploadFiles = ref<File[]>([]);
const zipContentsMap = ref<Record<string, string[]>>({});
const showZipPreview = ref(false);
const zipPreviewFile = ref<File | null>(null);
const zipPreviewEntries = ref<string[]>([]);
const zipPreviewContext = ref<"server" | "client">("server");
const isDraggingOver = ref(false);
const bulkUploadLoading = ref(false);
const installToServer = ref(false);
const installToClient = ref(false);

const isModUploadContext = computed(() =>
  uploadAllowedExts.value.includes(".jar"),
);

watch(showUploadModal, (open) => {
  if (!open) return;
  if (activeMainTab.value === "server") {
    installToServer.value = true;
    installToClient.value = false;
  } else if (activeClientSubTab.value === "mods") {
    installToServer.value = false;
    installToClient.value = true;
  } else {
    installToServer.value = false;
    installToClient.value = false;
  }
});

const router = useRouter();

const validMainTabs = ["server", "client"] as const;
const validSubTabs = ["mods", "resourcepacks", "shaderpacks"] as const;

const activeMainTab = ref<"server" | "client">(
  validMainTabs.includes(route.query.tab as any)
    ? (route.query.tab as "server" | "client")
    : "server",
);
const activeClientSubTab = ref<"mods" | "resourcepacks" | "shaderpacks">(
  validSubTabs.includes(route.query.sub as any)
    ? (route.query.sub as "mods" | "resourcepacks" | "shaderpacks")
    : "mods",
);

watch(activeMainTab, (v) => {
  router.replace({ query: { ...route.query, tab: v } });
});
watch(activeClientSubTab, (v) => {
  router.replace({ query: { ...route.query, sub: v } });
});

const clientTabs = [
  { key: "mods" as const, label: "Mods", icon: Box },
  { key: "resourcepacks" as const, label: "Resource Packs", icon: Image },
  { key: "shaderpacks" as const, label: "Shader Packs", icon: Sparkles },
];

const sideOptions = [
  { label: "All", value: "all" as const },
  { label: "Server", value: "server" as const },
  { label: "Client", value: "client" as const },
];

const searchPlaceholder = computed(() => {
  if (activeMainTab.value === "server") return "Search server mods...";
  if (activeClientSubTab.value === "mods") return "Search client mods...";
  if (activeClientSubTab.value === "resourcepacks")
    return "Search resource packs...";
  return "Search shader packs...";
});

const { trigger: triggerClientAssetsRefresh } =
  useClientAssetsRefresh(serverId);

const assetUploadLoading = computed(() => {
  if (activeClientSubTab.value === "resourcepacks")
    return resourcePacks.uploadLoading.value;
  if (activeClientSubTab.value === "shaderpacks")
    return shaderPacks.uploadLoading.value;
  return false;
});

const assetDownloadLoading = computed(() => false);

const uploadButtonLabel = computed(() => {
  if (activeMainTab.value === "server") return "Upload Mods";
  if (activeClientSubTab.value === "mods") return "Upload Mods";
  if (activeClientSubTab.value === "resourcepacks")
    return "Upload Resource Packs";
  return "Upload Shader Packs";
});

const uploadAcceptTypes = computed(() => {
  if (activeMainTab.value === "server" || activeClientSubTab.value === "mods")
    return ".jar,.zip";
  return ".zip";
});

const uploadSupportText = computed(() => {
  if (activeMainTab.value === "server" || activeClientSubTab.value === "mods")
    return "Supports .jar and .zip";
  return "Supports .zip";
});

const urlPlaceholder = computed(() => {
  if (activeMainTab.value === "server" || activeClientSubTab.value === "mods")
    return "https://example.com/file.jar";
  return "https://example.com/file.zip";
});

const uploadAllowedExts = computed(() => {
  if (activeMainTab.value === "server" || activeClientSubTab.value === "mods")
    return [".jar", ".zip"];
  return [".zip"];
});

const anyUploadLoading = computed(
  () =>
    uploadLoading.value ||
    clientUploadLoading.value ||
    assetUploadLoading.value ||
    bulkUploadLoading.value,
);

const uploadTargetLabel = computed(() => {
  if (activeMainTab.value === "server") return "Server Mod";
  if (activeClientSubTab.value === "mods") return "Client Mod";
  if (activeClientSubTab.value === "resourcepacks") return "Resource Pack";
  return "Shader Pack";
});

const downloadTargetLabel = computed(() => {
  if (activeMainTab.value === "server") return "server mods";
  if (activeClientSubTab.value === "mods") return "client mods";
  if (activeClientSubTab.value === "resourcepacks") return "resource packs";
  return "shader packs";
});

const libraryProjectType = computed<"mod" | "resourcepack" | "shaderpack">(
  () => {
    if (activeMainTab.value === "server") return "mod";
    if (activeClientSubTab.value === "resourcepacks") return "resourcepack";
    if (activeClientSubTab.value === "shaderpacks") return "shaderpack";
    return "mod";
  },
);

watch(showLibraryPanel, (open) => {
  if (!open) return;
  const lt = libraryProjectType.value;
  if (lt && modrinth.projectType.value !== lt) {
    modrinth.projectType.value = lt;
    modrinth.search(serverEngine.value, serverGameVersion.value);
  }
});

const installedModBasenames = computed(() => {
  const server = new Set<string>(),
    client = new Set<string>();
  const add = (set: Set<string>, filename: string) => {
    const base = filename.replace(/\.[^.]+$/, "").toLowerCase();
    if (base) set.add(base);
  };
  mods.value.forEach((m) => add(server, m.filename));
  clientModList.value.forEach((m) => add(client, m.filename));
  return { server, client };
});

const installedAssetBasenames = computed(() => {
  const names = new Set<string>();
  const add = (filename: string) => {
    const base = filename.replace(/\.[^.]+$/, "").toLowerCase();
    if (base) names.add(base);
  };
  resourcePacks.assets.value.forEach((a) => add(a.filename));
  shaderPacks.assets.value.forEach((a) => add(a.filename));
  return names;
});

const installedSideFilter = ref<"all" | "server" | "client">("all");
const clientSideFilter = ref<"all" | "server" | "client">("all");
const librarySideFilter = ref<"all" | "server" | "client">("all");
const sortBy = ref<
  "name-asc" | "name-desc" | "size-asc" | "size-desc" | "date-asc" | "date-desc"
>("name-asc");

const showBatchDeleteModal = ref(false);
const batchDeleteTarget = ref<"server" | "client">("server");
const batchDeleteFilenames = ref<string[]>([]);

function openBatchDelete(target: "server" | "client", filenames: string[]) {
  batchDeleteTarget.value = target;
  batchDeleteFilenames.value = filenames;
  showBatchDeleteModal.value = true;
}

async function confirmBatchDelete() {
  showBatchDeleteModal.value = false;
  if (batchDeleteTarget.value === "server") {
    await handleBatchDelete(batchDeleteFilenames.value);
  } else {
    await handleClientBatchDelete(batchDeleteFilenames.value);
  }
  batchDeleteFilenames.value = [];
}

async function handleBatchDelete(filenames: string[]) {
  await deleteMods(filenames);
}

async function handleBatchToggle(filenames: string[]) {
  await toggleMods(filenames);
}

async function handleClientBatchDelete(filenames: string[]) {
  await deleteClientMods(filenames);
}

async function handleClientBatchToggle(filenames: string[]) {
  await toggleClientMods(filenames);
}

const showCopyAllModal = ref(false);
const copyAllLoading = ref(false);
const modsToCopy = ref<typeof mods.value>([]);

function openCopyAllConfirm() {
  const clientFilenames = new Set(clientModList.value.map((m) => m.filename));
  modsToCopy.value = mods.value.filter((m) => !clientFilenames.has(m.filename));
  showCopyAllModal.value = true;
}

function openFileInManager(
  filename: string,
  folder: "mods" | "mods-client" | "resourcepacks" | "shaderpacks",
) {
  navigateTo({
    path: `/servers/${serverId}/files`,
    query: { path: `${folder}/${filename}` },
  });
}

async function handleCopyAll() {
  if (modsToCopy.value.length === 0) {
    showCopyAllModal.value = false;
    return;
  }
  copyAllLoading.value = true;
  try {
    const res: any = await $fetch(`/servers/${serverId}/mods/copy-all`, {
      baseURL: useApiBase(),
      method: "POST",
      credentials: "include",
    });
    useToast().show(
      "success",
      `Copied ${res.copied?.length || 0} mod(s) to client`,
      {
        description: res.skipped?.length
          ? `${res.skipped.length} skipped (already exist)`
          : undefined,
      },
    );
    await refresh();
    await refreshClient();
    showCopyAllModal.value = false;
  } catch (err: any) {
    useToast().show("error", "Failed to copy mods", {
      description: getApiErrorMessage(err, "Unknown error"),
    });
  } finally {
    copyAllLoading.value = false;
  }
}

const versionDetailsMap = ref<
  Record<string, import("~/composables/useModrinth").ModrinthVersion>
>({});
const depProjectsMap = ref<
  Record<string, import("~/composables/useModrinth").ModrinthProject>
>({});

async function onFileSelect(e: Event) {
  const input = e.target as HTMLInputElement;
  if (!input.files) return;
  for (let i = 0; i < input.files.length; i++) {
    const file = input.files[i];
    if (!file) continue;
    if (
      !uploadAllowedExts.value.some((ext) =>
        file.name.toLowerCase().endsWith(ext),
      )
    )
      continue;
    if (pendingUploadFiles.value.some((f) => f.name === file.name)) continue;
    pendingUploadFiles.value.push(file);
    if (
      file.name.toLowerCase().endsWith(".zip") &&
      uploadAllowedExts.value.includes(".jar")
    ) {
      try {
        const zip = await JSZip.loadAsync(file);
        const jars: string[] = [];
        zip.forEach((relativePath, entry) => {
          if (!entry.dir && relativePath.toLowerCase().endsWith(".jar"))
            jars.push(relativePath);
        });
        zipContentsMap.value[file.name] = jars;
      } catch {
        /* ignore */
      }
    }
  }
  input.value = "";
}

async function handleUploadConfirm() {
  if (!modUrl.value && pendingUploadFiles.value.length === 0) return;
  bulkUploadLoading.value = true;
  try {
    if (modUrl.value) {
      await performDownloadFromURL();
      modUrl.value = "";
    }
    for (const file of pendingUploadFiles.value) {
      if (isModUploadContext.value) {
        if (installToServer.value) await uploadFile(file);
        if (installToClient.value) await uploadClientFile(file);
      } else {
        const assetType = activeClientSubTab.value;
        const formData = new FormData();
        formData.append("file", file);
        await $fetch(
          `/servers/${serverId}/client-assets/upload?type=${assetType}`,
          {
            baseURL: useApiBase(),
            method: "POST",
            body: formData,
            credentials: "include",
          },
        );
        useToast().show("success", "Upload complete", {
          description: file.name,
        });
        triggerClientAssetsRefresh();
      }
    }
    if (installToClient.value && pendingUploadFiles.value.length > 0) {
      await refreshClient();
    }
    if (installToServer.value && pendingUploadFiles.value.length > 0) {
      await refresh();
    }
  } catch (err: any) {
    useToast().show("error", "Failed to process files", {
      description: getApiErrorMessage(err, "Unknown error"),
    });
  } finally {
    bulkUploadLoading.value = false;
    closeUploadModal();
  }
}

async function downloadResourcePackFromURL(url: string, filename?: string) {
  await $fetch(
    `/servers/${serverId}/client-assets/download?type=resourcepacks`,
    {
      baseURL: useApiBase(),
      method: "POST",
      credentials: "include",
      body: { url, filename },
    },
  );
}

async function downloadShaderPackFromURL(url: string, filename?: string) {
  await $fetch(`/servers/${serverId}/client-assets/download?type=shaderpacks`, {
    baseURL: useApiBase(),
    method: "POST",
    credentials: "include",
    body: { url, filename },
  });
}

async function performDownloadFromURL() {
  if (!modUrl.value) return;
  if (isModUploadContext.value) {
    if (installToServer.value) await downloadFromURL(modUrl.value);
    if (installToClient.value) {
      await downloadClientFromURL(modUrl.value);
    }
  } else if (activeClientSubTab.value === "resourcepacks") {
    await downloadResourcePackFromURL(modUrl.value);
    triggerClientAssetsRefresh();
  } else if (activeClientSubTab.value === "shaderpacks") {
    await downloadShaderPackFromURL(modUrl.value);
    triggerClientAssetsRefresh();
  }
}

function closeUploadModal() {
  showUploadModal.value = false;
  pendingUploadFiles.value = [];
  zipContentsMap.value = {};
  modUrl.value = "";
}

function onDrop(e: DragEvent) {
  isDraggingOver.value = false;
  const files = e.dataTransfer?.files;
  if (!files) return;
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    if (!file) continue;
    if (
      !uploadAllowedExts.value.some((ext) =>
        file.name.toLowerCase().endsWith(ext),
      )
    )
      continue;
    if (pendingUploadFiles.value.some((f) => f.name === file.name)) continue;
    pendingUploadFiles.value.push(file);
    if (
      file.name.toLowerCase().endsWith(".zip") &&
      uploadAllowedExts.value.includes(".jar")
    ) {
      JSZip.loadAsync(file)
        .then((zip) => {
          const jars: string[] = [];
          zip.forEach((relativePath, entry) => {
            if (!entry.dir && relativePath.toLowerCase().endsWith(".jar"))
              jars.push(relativePath);
          });
          zipContentsMap.value[file.name] = jars;
        })
        .catch(() => {
          /* ignore */
        });
    }
  }
}

function removePendingFile(idx: number) {
  const file = pendingUploadFiles.value[idx];
  pendingUploadFiles.value.splice(idx, 1);
  if (file) delete zipContentsMap.value[file.name];
}

const modrinth = useModrinth();
const serverEngine = ref("");
const serverGameVersion = ref("");
const serverLoaderVersion = ref("");
const versionsMap = ref<
  Record<string, import("~/composables/useModrinth").ModrinthVersion[]>
>({});

const debouncedSearch = debounce(() => {
  modrinth.search(serverEngine.value, serverGameVersion.value);
}, 1300);
onBeforeUnmount(() => {
  debouncedSearch.cancel();
});

function onSearchInput(v: string) {
  modrinth.searchQuery.value = v;
  debouncedSearch();
}
function onProjectTypeChange(v: "mod" | "resourcepack" | "shaderpack") {
  modrinth.projectType.value = v;
  debouncedSearch();
}

async function fetchServerInfo() {
  try {
    const res = await $fetch<
      | {
          body: {
            engineType: string;
            gameVersion: string;
            loaderVersion?: string;
          };
        }
      | { engineType: string; gameVersion: string; loaderVersion?: string }
    >(`/servers/${serverId}`, {
      baseURL: useApiBase(),
      credentials: "include",
    });
    const data = (res as any).body ?? res;
    const engine = (data.engineType ?? "").toUpperCase();
    serverEngine.value = engine;
    serverGameVersion.value = data.gameVersion ?? "";
    serverLoaderVersion.value = data.loaderVersion ?? "";
    if (engine === "PAPERMC" || engine === "PAPER") {
      await navigateTo(`/servers/${serverId}/plugins`, { replace: true });
      return;
    }
    if (engine === "VANILLA") {
      await navigateTo(`/servers/${serverId}`, { replace: true });
      return;
    }
  } catch {
    /* ignore */
  }
}

async function handleUpload(file: File) {
  if (file.name.toLowerCase().endsWith(".zip")) {
    const maxPreviewSize = 50 << 20; // 50 MB
    if (file.size > maxPreviewSize) {
      zipPreviewFile.value = file;
      zipPreviewEntries.value = [];
      showZipPreview.value = true;
      return;
    }
    try {
      const zip = await JSZip.loadAsync(file);
      const entries: string[] = [];
      zip.forEach((relativePath, entry) => {
        if (!entry.dir) entries.push(relativePath);
      });
      zipPreviewFile.value = file;
      zipPreviewEntries.value = entries;
      zipPreviewContext.value = "server";
      showZipPreview.value = true;
    } catch {
      useToast().show("error", "Invalid ZIP", {
        description: "Could not read archive contents.",
      });
    }
    return;
  }
  await uploadFile(file);
}

async function handleZipConfirm() {
  console.log(
    "[handleZipConfirm] context:",
    zipPreviewContext.value,
    "file:",
    zipPreviewFile.value?.name,
  );
  const file = zipPreviewFile.value;
  const ctx = zipPreviewContext.value;
  closeZipPreview();
  if (!file) return;
  if (ctx === "client") {
    await uploadClientFile(file);
  } else {
    await uploadFile(file);
  }
}

function closeZipPreview() {
  showZipPreview.value = false;
  zipPreviewFile.value = null;
  zipPreviewEntries.value = [];
  zipPreviewContext.value = "server";
}

async function handleToggle(filename: string) {
  await toggleMod(filename);
}
async function handleClientUpload(file: File) {
  console.log(
    "[handleClientUpload] file:",
    file.name,
    "size:",
    file.size,
    "type:",
    file.type,
  );
  if (file.name.toLowerCase().endsWith(".zip")) {
    console.log("[handleClientUpload] detected ZIP, opening preview...");
    const maxPreviewSize = 50 << 20;
    if (file.size > maxPreviewSize) {
      zipPreviewFile.value = file;
      zipPreviewEntries.value = [];
      zipPreviewContext.value = "client";
      showZipPreview.value = true;
      return;
    }
    try {
      const zip = await JSZip.loadAsync(file);
      const entries: string[] = [];
      zip.forEach((relativePath, entry) => {
        if (!entry.dir) entries.push(relativePath);
      });
      zipPreviewFile.value = file;
      zipPreviewEntries.value = entries;
      zipPreviewContext.value = "client";
      showZipPreview.value = true;
    } catch (err) {
      console.error("[handleClientUpload] JSZip failed:", err);
      useToast().show("error", "Invalid ZIP", {
        description: "Could not read archive contents.",
      });
    }
    return;
  }
  await uploadClientFile(file);
}
async function handleClientDelete(filename: string) {
  await deleteClientMod(filename);
}
async function handleClientToggle(filename: string) {
  await toggleClientMod(filename);
}
async function handleClientMove(filename: string, target: "server" | "client") {
  await moveClientMod(filename, target);
  await refresh();
  await refreshClient();
}
async function handleAssetUpload(file: File) {
  if (activeClientSubTab.value === "resourcepacks") {
    await resourcePacks.uploadFile(file);
  } else if (activeClientSubTab.value === "shaderpacks") {
    await shaderPacks.uploadFile(file);
  }
  triggerClientAssetsRefresh();
}

async function handleCopy(
  filename: string,
  source: "server" | "client",
  target: "server" | "client",
) {
  await copyClientMod(filename, source, target);
  await refresh();
  await refreshClient();
}

async function handleInstallFromLibrary(
  projectId: string,
  versionId: string,
  depProjectIds?: string[],
  _target?: "server" | "client" | "both",
) {
  const file = await modrinth.install(projectId, versionId);
  if (!file) return;

  const projectType = modrinth.projectType.value;
  const isAsset =
    projectType === "resourcepack" || projectType === "shaderpack";
  const assetType =
    projectType === "shaderpack" ? "shaderpacks" : "resourcepacks";

  const allDeps = file.dependencies.filter(
    (d) => d.dependency_type !== "incompatible",
  );
  const selectedDeps = depProjectIds
    ? allDeps.filter((d) => depProjectIds.includes(d.project_id))
    : allDeps.filter((d) => d.dependency_type === "required");

  async function installDepsAndMod(
    modFile: NonNullable<typeof file>,
    doDownload: (url: string, filename?: string) => Promise<void>,
    doRefresh: () => Promise<void>,
  ) {
    if (selectedDeps.length > 0) {
      useToast().show(
        "info",
        `Installing ${selectedDeps.length} dependencies...`,
      );
      for (const dep of selectedDeps) {
        const depFile = await modrinth.resolveDependency(
          dep.project_id,
          dep.version_id,
          serverEngine.value,
          serverGameVersion.value,
        );
        if (depFile) await doDownload(depFile.url, depFile.filename);
      }
    }
    await doDownload(modFile.url, modFile.filename);
    await doRefresh();
  }

  if (isAsset) {
    const doDownload =
      assetType === "shaderpacks"
        ? downloadShaderPackFromURL
        : downloadResourcePackFromURL;
    await installDepsAndMod(file, doDownload, async () => {
      triggerClientAssetsRefresh();
    });
    useToast().show("success", "Installed to client", {
      description: file.filename,
    });
    return;
  }

  if (activeMainTab.value === "server") {
    await installDepsAndMod(file, downloadFromURL, refresh);
    useToast().show("success", "Mod installed to server", {
      description: file.filename,
    });
  } else {
    await installDepsAndMod(file, downloadClientFromURL, refreshClient);
    useToast().show("success", "Mod installed to client", {
      description: file.filename,
    });
  }
}

async function handleLoadVersions(projectId: string) {
  const loader = modrinth.projectType.value === "mod" ? serverEngine.value : "";
  const list = await modrinth.getVersions(
    projectId,
    loader,
    serverGameVersion.value,
  );
  if (list.length > 0)
    versionsMap.value = { ...versionsMap.value, [projectId]: list };
}

async function handleLoadVersionDetails(versionId: string) {
  const details = await modrinth.getVersionDetails(versionId);
  if (details)
    versionDetailsMap.value = {
      ...versionDetailsMap.value,
      [versionId]: details,
    };
}

async function handleLoadProject(projectId: string) {
  const project = await modrinth.getProject(projectId);
  if (project)
    depProjectsMap.value = { ...depProjectsMap.value, [projectId]: project };
}

const archiveResult = ref<
  import("~/composables/useClientMods").ArchiveInfo | null
>(null);
const archiveLinks = ref<
  import("~/composables/useClientMods").ArchiveLinkEntry[]
>([]);
const linksLoading = ref(false);

async function handleGenerateArchive(
  ttl: number,
  include: string[] = ["mods"],
) {
  const result = await createArchive(ttl, include);
  if (result) {
    archiveResult.value = result;
    await handleRefreshLinks();
  }
}

async function handleRefreshLinks() {
  linksLoading.value = true;
  try {
    archiveLinks.value = await listArchives();
  } finally {
    linksLoading.value = false;
  }
}

async function handleDeleteLink(token: string) {
  const ok = await deleteArchive(token);
  if (ok)
    archiveLinks.value = archiveLinks.value.filter((l) => l.token !== token);
}

onMounted(async () => {
  await fetchServerInfo();
  if (!serverEngine.value) return;
  await Promise.all([
    refresh(),
    refreshClient(),
    resourcePacks.refresh(),
    shaderPacks.refresh(),
    modrinth.search(serverEngine.value, serverGameVersion.value),
  ]);
});
return {
    fileInput, showUploadModal, showLibraryPanel, showArchiveModal, modUrl, searchQuery, pendingUploadFiles, zipContentsMap, showZipPreview, zipPreviewFile, zipPreviewEntries, zipPreviewContext, isDraggingOver, bulkUploadLoading, installToServer, installToClient, activeMainTab, activeClientSubTab, clientTabs, sideOptions, searchPlaceholder, assetUploadLoading, assetDownloadLoading, uploadButtonLabel, uploadAcceptTypes, uploadSupportText, urlPlaceholder, uploadAllowedExts, anyUploadLoading, uploadTargetLabel, downloadTargetLabel, libraryProjectType, installedSideFilter, clientSideFilter, librarySideFilter, sortBy, showBatchDeleteModal, batchDeleteTarget, batchDeleteFilenames, showCopyAllModal, copyAllLoading, modsToCopy, versionDetailsMap, depProjectsMap, versionsMap, serverEngine, serverLoaderVersion, serverGameVersion, modrinth, archiveResult, archiveLinks, linksLoading,
    mods, clientModList, loading, clientLoading, uploadLoading, clientUploadLoading, downloadLoading, clientDownloadLoading, archiveLoading: clientArchiveLoading, resourcePacks, shaderPacks,
    openBatchDelete, confirmBatchDelete, handleBatchDelete, handleBatchToggle, handleClientBatchDelete, handleClientBatchToggle, openCopyAllConfirm, openFileInManager, handleCopyAll, onFileSelect, handleUploadConfirm, closeUploadModal, onDrop, removePendingFile, handleUpload, handleZipConfirm, closeZipPreview, handleToggle, handleClientUpload, handleClientDelete, handleClientToggle, handleClientMove, handleAssetUpload, handleCopy, handleInstallFromLibrary, handleLoadVersions, handleLoadVersionDetails, handleLoadProject, handleGenerateArchive, handleRefreshLinks, handleDeleteLink, onSearchInput, onProjectTypeChange, performDownloadFromURL,
  };
}
