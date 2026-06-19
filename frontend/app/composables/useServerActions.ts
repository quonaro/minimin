import type { Server } from "./useServers";

export function useServerActions(serverId: string, server: Ref<Server | null>) {
  const { show: showToast } = useToast();
  const { refresh: refreshServers } = useServers();

  const actionLoading = ref(false);
  const currentAction = ref<"start" | "stop" | "force-stop" | "restart" | null>(
    null,
  );
  const removeBeforeStart = ref(false);
  const deleteLoading = ref(false);
  const recreateLoading = ref(false);
  const showDeleteDialog = ref(false);
  const showRecreateDialog = ref(false);

  const isPending = computed(() => {
    if (!server.value) return false;
    const d = server.value.desiredStatus;
    return !!d && d !== server.value.containerStatus;
  });

  async function doAction(action: "start" | "stop" | "restart" | "force-stop") {
    if (actionLoading.value) return;
    actionLoading.value = true;
    currentAction.value = action;
    try {
      await $fetch(`/servers/${serverId}/${action}`, {
        baseURL: useApiBase(),
        method: "POST",
        credentials: "include",
        body:
          action === "start"
            ? { removeExisting: removeBeforeStart.value }
            : undefined,
      });
      showToast("info", `Server ${action} requested`, {
        description: `${serverId} — operation in progress.`,
      });
      await refreshServers();
    } catch (err: any) {
      const status = getApiErrorStatus(err);
      const msg = getApiErrorMessage(err, `Failed to ${action} server`);
      if (status === 409) {
        showToast("error", "Operation in progress", { description: msg });
      } else {
        showToast("error", `Server ${action} failed`, { description: msg });
      }
    } finally {
      actionLoading.value = false;
      currentAction.value = null;
    }
  }

  function promptDelete() {
    showDeleteDialog.value = true;
  }

  async function onDeleteConfirmed(wipe: boolean) {
    deleteLoading.value = true;
    try {
      await $fetch(`/servers/${serverId}`, {
        baseURL: useApiBase(),
        method: "DELETE",
        credentials: "include",
        query: { wipe: wipe ? "true" : "false" },
      });
      showToast("success", "Server deleted", {
        description: `${serverId} has been removed.`,
      });
      await refreshServers();
      await navigateTo(`/`, { replace: true });
    } catch (err: any) {
      const msg = getApiErrorMessage(err, "Failed to delete server");
      showToast("error", "Delete failed", { description: msg });
    } finally {
      deleteLoading.value = false;
    }
  }

  function promptRecreate() {
    showRecreateDialog.value = true;
  }

  async function onRecreateConfirmed() {
    recreateLoading.value = true;
    try {
      await $fetch(`/servers/${serverId}/recreate-world`, {
        baseURL: useApiBase(),
        method: "POST",
        credentials: "include",
      });
      showToast("info", "World recreate requested", {
        description: `${serverId} — world will be reset.`,
      });
      await refreshServers();
    } catch (err: any) {
      const msg = getApiErrorMessage(err, "Failed to recreate world");
      showToast("error", "Recreate failed", { description: msg });
    } finally {
      recreateLoading.value = false;
    }
  }

  return {
    actionLoading,
    currentAction,
    removeBeforeStart,
    isPending,
    deleteLoading,
    recreateLoading,
    showDeleteDialog,
    showRecreateDialog,
    doAction,
    promptDelete,
    onDeleteConfirmed,
    promptRecreate,
    onRecreateConfirmed,
  };
}
