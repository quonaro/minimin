<script setup lang="ts">
definePageMeta({ middleware: "auth" });

const {
  warningModalOpen,
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
} = useFilesExplorer();
</script>

<template>
  <div class="p-6">
    <div class="space-y-6">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Files</h1>
      </div>

      <div class="flex items-center gap-2 flex-wrap text-sm">
        <button
          class="text-primary hover:underline"
          :class="
            dragOverBreadcrumbPath === '' ? 'font-semibold underline' : ''
          "
          @click="navigateToPath('')"
          @dragover="onBreadcrumbDragOver('', $event)"
          @dragleave="onBreadcrumbDragLeave('')"
          @drop="onBreadcrumbDrop('', $event)"
        >
          /
        </button>
        <template
          v-for="(crumb, idx) in breadcrumbs"
          :key="crumb.path || 'root'"
        >
          <span class="text-gray-400 dark:text-neutral-600">/</span>
          <button
            class="text-primary hover:underline"
            :class="
              dragOverBreadcrumbPath === crumb.path
                ? 'font-semibold underline'
                : ''
            "
            @click="navigateToPath(crumb.path)"
            @dragover="onBreadcrumbDragOver(crumb.path, $event)"
            @dragleave="onBreadcrumbDragLeave(crumb.path)"
            @drop="onBreadcrumbDrop(crumb.path, $event)"
          >
            📁 {{ crumb.label }}
          </button>
          <span
            v-if="idx === breadcrumbs.length - 1"
            class="text-xs text-gray-500 dark:text-neutral-400 ml-1"
          >
            (current)
          </span>
        </template>
      </div>

      <div class="flex items-center gap-3 flex-wrap">
        <input
          v-model="search"
          type="text"
          placeholder="Search in current folder..."
          class="w-full md:w-80 px-3 py-2 rounded-xl bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-sm text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-neutral-500 focus:ring-2 focus:ring-primary focus:outline-none"
        />
        <button
          class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-white text-sm font-medium shadow-sm hover:shadow-md active:scale-95 transition-all hover:bg-primary/90"
          @click="contextCreateFile"
        >
          New file
        </button>
        <button
          class="inline-flex items-center gap-2 px-4 py-2 rounded-xl border border-gray-200 dark:border-neutral-700 bg-white dark:bg-neutral-800 text-sm font-medium text-gray-700 dark:text-neutral-300 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-all"
          @click="contextUpload"
        >
          Upload file
        </button>
        <template v-if="selectedEntry">
          <span class="text-sm text-gray-700 dark:text-neutral-300 font-medium">
            {{ selectedEntry.name }}
          </span>
          <button
            class="inline-flex items-center gap-2 px-4 py-2 rounded-xl border border-gray-200 dark:border-neutral-700 bg-white dark:bg-neutral-800 text-sm font-medium text-gray-700 dark:text-neutral-300 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-all"
            @click="openEntry(selectedEntry)"
          >
            Open
          </button>
          <button
            v-if="!selectedEntry.isDir"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-xl border border-gray-200 dark:border-neutral-700 bg-white dark:bg-neutral-800 text-sm font-medium text-gray-700 dark:text-neutral-300 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-all"
            @click="downloadFile(selectedEntry.path)"
          >
            Download
          </button>
          <button
            class="inline-flex items-center gap-2 px-4 py-2 rounded-xl border border-gray-200 dark:border-neutral-700 bg-white dark:bg-neutral-800 text-sm font-medium text-gray-700 dark:text-neutral-300 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-all"
            @click="openRenameModal(selectedEntry.path)"
          >
            Rename
          </button>
          <button
            class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-red-500 hover:bg-red-600 text-white text-sm font-medium shadow-sm hover:shadow-md active:scale-95 transition-all"
            @click="openDeleteModal(selectedEntry.path)"
          >
            Delete
          </button>
        </template>
      </div>

      <div
        class="grid grid-cols-1 xl:grid-cols-[1fr_440px] gap-6 h-[calc(100vh-16rem)] min-h-[500px]"
      >
        <div
          class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden flex flex-col min-h-0"
        >
          <div
            class="px-4 py-2 bg-gray-50 dark:bg-neutral-900 border-b border-gray-200 dark:border-neutral-700 text-xs text-gray-500 dark:text-neutral-400"
          >
            {{ currentPath || "/" }} · {{ displayEntries.length }} items
          </div>
          <div
            class="flex-1 min-h-0 overflow-auto"
            @click="clearSelection"
            @contextmenu="openEmptyAreaContextMenu"
          >
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
              v-else-if="displayEntries.length === 0"
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
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="entry in displayEntries"
                  :key="entry.path"
                  data-entry-row="true"
                  class="border-t border-gray-100 dark:border-neutral-800 hover:bg-gray-50 dark:hover:bg-neutral-900 cursor-pointer"
                  :class="
                    (dragOverDirPath === entry.path && entry.isDir
                      ? 'bg-primary/10 dark:bg-primary/20'
                      : '') ||
                    (selectedEntry?.path === entry.path
                      ? 'bg-blue-50 dark:bg-blue-900/20'
                      : '')
                  "
                  :draggable="entry.name !== '..'"
                  @dragstart="
                    entry.name !== '..'
                      ? onEntryDragStart(entry, $event)
                      : undefined
                  "
                  @dragend="onEntryDragEnd"
                  @click.stop="selectEntry(entry)"
                  @dblclick.stop="openEntry(entry)"
                  @contextmenu="
                    entry.name !== '..'
                      ? openEntryContextMenu(entry, $event)
                      : undefined
                  "
                  @dragover="
                    entry.name !== '..'
                      ? onEntryDragOver(entry, $event)
                      : undefined
                  "
                  @dragleave="onEntryDragLeave(entry)"
                  @drop="
                    entry.name !== '..' ? onEntryDrop(entry, $event) : undefined
                  "
                >
                  <td class="px-3 py-2">
                    <span
                      class="text-left hover:underline block"
                      :class="
                        entry.isDir
                          ? 'text-gray-700 dark:text-neutral-300 font-medium'
                          : 'text-gray-900 dark:text-white'
                      "
                    >
                      {{ entry.isDir ? "📁" : "📄" }} {{ entry.name }}
                    </span>
                  </td>
                  <td class="px-3 py-2 text-gray-500 dark:text-neutral-400">
                    {{ entry.isDir ? "—" : formatSize(entry.size) }}
                  </td>
                  <td class="px-3 py-2 text-gray-500 dark:text-neutral-400">
                    {{ formatDate(entry.modifiedAt) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div
          class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden flex flex-col min-h-0"
        >
          <div
            class="px-4 py-2 bg-gray-50 dark:bg-neutral-900 border-b border-gray-200 dark:border-neutral-700 text-sm font-medium text-gray-900 dark:text-white"
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
              <div
                class="text-xs text-gray-500 dark:text-neutral-400 break-all"
              >
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
                  class="flex-1 min-h-0 w-full rounded-xl border border-gray-200 dark:border-neutral-700 bg-white dark:bg-neutral-900 p-3 font-mono text-xs text-gray-900 dark:text-neutral-100 resize-none focus:ring-2 focus:ring-primary focus:outline-none"
                />
                <div class="flex justify-end">
                  <button
                    class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-white text-sm font-medium shadow-sm hover:shadow-md active:scale-95 transition-all hover:bg-primary/90"
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
        v-if="contextMenuOpen"
        class="fixed inset-0 z-40"
        @click="closeContextMenu"
        @contextmenu.prevent
      >
        <div
          data-files-context-menu="true"
          class="fixed z-50 min-w-[210px] bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-xl shadow-xl py-1"
          :style="{ left: `${contextMenuX}px`, top: `${contextMenuY}px` }"
          @click.stop
        >
          <template v-if="contextTargetType === 'file'">
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700"
              @click="contextOpenTarget"
            >
              Open
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700"
              @click="contextDownloadTarget"
            >
              Download
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700"
              @click="contextRenameTarget"
            >
              Rename
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
              @click="contextDeleteTarget"
            >
              Delete
            </button>
          </template>

          <template v-else-if="contextTargetType === 'dir'">
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700"
              @click="contextOpenTarget"
            >
              Open
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700"
              @click="contextRenameTarget"
            >
              Rename
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
              @click="contextDeleteTarget"
            >
              Delete
            </button>
          </template>

          <template v-else>
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700"
              @click="contextCreateFile"
            >
              New file
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700"
              @click="contextCreateFolder"
            >
              New folder
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700"
              @click="contextUpload"
            >
              Upload
            </button>
            <button
              class="w-full px-3 py-2 text-left text-sm text-gray-700 dark:text-neutral-200 hover:bg-gray-100 dark:hover:bg-neutral-700"
              @click="contextRefresh"
            >
              Refresh
            </button>
          </template>
        </div>
      </div>

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
          <p
            class="text-sm text-gray-700 dark:text-neutral-300 leading-relaxed"
          >
            Any accidental file change may damage saves, world data, or server
            configuration. Continue only if you understand the risk.
          </p>
          <div class="mt-5 flex justify-end">
            <button
              class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-red-600 text-white font-medium shadow-sm hover:shadow-md active:scale-95 transition-all hover:bg-red-700"
              @click="warningModalOpen = false"
            >
              I understand the risk
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
              class="mt-1 w-full px-3 py-2 rounded-xl border border-gray-200 dark:border-neutral-700 bg-white dark:bg-neutral-900 text-sm text-gray-900 dark:text-neutral-100 focus:ring-2 focus:ring-primary focus:outline-none"
            />
          </div>
          <div class="mt-5 flex justify-end gap-2">
            <button
              class="inline-flex items-center gap-2 px-4 py-2 rounded-xl border border-gray-300 dark:border-neutral-700 hover:bg-gray-100 dark:hover:bg-neutral-700 text-sm font-medium text-gray-700 dark:text-neutral-200 transition-all"
              @click="closeConfirmModal"
            >
              Cancel
            </button>
            <button
              class="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-white text-sm font-medium shadow-sm hover:shadow-md active:scale-95 transition-all"
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
  </div>
</template>
