import type { Server } from "./useServers";

export function useServerConfigEdits(
  serverId: string,
  server: Ref<Server | null>,
) {
  const { show: showToast } = useToast();
  const { refresh: refreshServers } = useServers();

  const editingPort = ref(false);
  const tempPort = ref<number | null>(null);
  const portLoading = ref(false);

  const editingRestartPolicy = ref(false);
  const tempRestartPolicy = ref<string>("");
  const restartPolicyLoading = ref(false);

  const editingPublicRcon = ref(false);
  const tempPublicRcon = ref(false);
  const tempRconPort = ref<number | null>(null);
  const rconLoading = ref(false);

  const editingRam = ref(false);
  const tempRamGb = ref<number | null>(null);
  const ramLoading = ref(false);

  const editingExternalJavaArgs = ref(false);
  const tempExternalJavaArgs = ref<string[]>([]);
  const newExternalArg = ref("");
  const externalJavaArgsLoading = ref(false);

  async function savePort() {
    if (
      !server.value ||
      tempPort.value == null ||
      tempPort.value === server.value.gamePort
    ) {
      editingPort.value = false;
      return;
    }
    if (tempPort.value < 1024 || tempPort.value > 65535) {
      showToast("error", "Invalid port", {
        description: "Port must be between 1024 and 65535.",
      });
      return;
    }
    portLoading.value = true;
    try {
      await $fetch(`/servers/${serverId}`, {
        baseURL: useApiBase(),
        method: "PATCH",
        credentials: "include",
        body: { gamePort: tempPort.value },
      });
      showToast("success", "Port updated", {
        description: `Game port changed to ${tempPort.value}.`,
      });
      editingPort.value = false;
      await refreshServers();
    } catch (err: any) {
      const status = err?.status || err?.statusCode;
      const msg = err?.data?.detail || err?.message || "Failed to update port";
      if (status === 409) {
        showToast("error", "Port unavailable", { description: msg });
      } else {
        showToast("error", "Update failed", { description: msg });
      }
    } finally {
      portLoading.value = false;
    }
  }

  async function saveRestartPolicy() {
    if (
      !server.value ||
      tempRestartPolicy.value === (server.value.restartPolicy || "no")
    ) {
      editingRestartPolicy.value = false;
      return;
    }
    restartPolicyLoading.value = true;
    try {
      await $fetch(`/servers/${serverId}`, {
        baseURL: useApiBase(),
        method: "PATCH",
        credentials: "include",
        body: { restartPolicy: tempRestartPolicy.value },
      });
      showToast("success", "Restart policy updated", {
        description: `Policy changed to ${tempRestartPolicy.value}.`,
      });
      editingRestartPolicy.value = false;
      await refreshServers();
    } catch (err: any) {
      const msg =
        err?.data?.detail || err?.message || "Failed to update restart policy";
      showToast("error", "Update failed", { description: msg });
    } finally {
      restartPolicyLoading.value = false;
    }
  }

  async function savePublicRcon() {
    if (!server.value || tempRconPort.value == null) {
      editingPublicRcon.value = false;
      return;
    }
    if (
      tempPublicRcon.value &&
      (tempRconPort.value < 1024 || tempRconPort.value > 65535)
    ) {
      showToast("error", "Invalid port", {
        description: "Port must be between 1024 and 65535.",
      });
      return;
    }
    if (
      tempPublicRcon.value === server.value.publicRcon &&
      tempRconPort.value ===
        (server.value.rconPort || server.value.gamePort + 10)
    ) {
      editingPublicRcon.value = false;
      return;
    }
    rconLoading.value = true;
    try {
      await $fetch(`/servers/${serverId}`, {
        baseURL: useApiBase(),
        method: "PATCH",
        credentials: "include",
        body: {
          publicRcon: tempPublicRcon.value,
          rconPort: tempPublicRcon.value ? tempRconPort.value : undefined,
        },
      });
      showToast("success", "Public RCON updated", {
        description: tempPublicRcon.value
          ? `RCON enabled on port ${tempRconPort.value}.`
          : "RCON disabled.",
      });
      editingPublicRcon.value = false;
      await refreshServers();
    } catch (err: any) {
      const status = err?.status || err?.statusCode;
      const msg = err?.data?.detail || err?.message || "Failed to update RCON";
      if (status === 409) {
        showToast("error", "Port unavailable", { description: msg });
      } else {
        showToast("error", "Update failed", { description: msg });
      }
    } finally {
      rconLoading.value = false;
    }
  }

  async function saveRam() {
    if (!server.value || tempRamGb.value == null) {
      editingRam.value = false;
      return;
    }
    const currentGb = server.value.ramBytes
      ? Math.round(server.value.ramBytes / (1024 * 1024 * 1024))
      : 0;
    if (tempRamGb.value === currentGb) {
      editingRam.value = false;
      return;
    }
    if (tempRamGb.value < 1 || tempRamGb.value > 128) {
      showToast("error", "Invalid RAM", {
        description: "RAM must be between 1 and 128 GB.",
      });
      return;
    }
    ramLoading.value = true;
    try {
      await $fetch(`/servers/${serverId}`, {
        baseURL: useApiBase(),
        method: "PATCH",
        credentials: "include",
        body: { ramBytes: tempRamGb.value * 1024 * 1024 * 1024 },
      });
      showToast("success", "RAM updated", {
        description: `RAM changed to ${tempRamGb.value} GB.`,
      });
      editingRam.value = false;
      await refreshServers();
    } catch (err: any) {
      const status = err?.status || err?.statusCode;
      const msg = err?.data?.detail || err?.message || "Failed to update RAM";
      if (status === 409) {
        showToast("error", "Server running", { description: msg });
      } else {
        showToast("error", "Update failed", { description: msg });
      }
    } finally {
      ramLoading.value = false;
    }
  }

  function handlePasteExternalJavaArgs(event: ClipboardEvent) {
    const text = event.clipboardData?.getData("text") || "";
    const args = text
      .split(/\r?\n|\s+/)
      .map((s) => s.trim())
      .filter((s) => s.length > 0);
    if (args.length > 1) {
      event.preventDefault();
      args.forEach((arg) => tempExternalJavaArgs.value.push(arg));
      newExternalArg.value = "";
    }
  }

  async function saveExternalJavaArgs() {
    if (!server.value) {
      editingExternalJavaArgs.value = false;
      return;
    }
    const args = tempExternalJavaArgs.value.filter((s) => s.trim().length > 0);
    const current = server.value.externalJavaArgs ?? [];
    const same =
      args.length === current.length && args.every((a, i) => a === current[i]);
    if (same) {
      editingExternalJavaArgs.value = false;
      return;
    }
    externalJavaArgsLoading.value = true;
    try {
      await $fetch(`/servers/${serverId}`, {
        baseURL: useApiBase(),
        method: "PATCH",
        credentials: "include",
        body: { externalJavaArgs: args },
      });
      showToast("success", "External Java args updated", {
        description: args.length > 0 ? args.join(" ") : "Cleared",
      });
      editingExternalJavaArgs.value = false;
      await refreshServers();
    } catch (err: any) {
      const status = err?.status || err?.statusCode;
      const msg =
        err?.data?.detail || err?.message || "Failed to update external Java args";
      if (status === 409) {
        showToast("error", "Server running", { description: msg });
      } else {
        showToast("error", "Update failed", { description: msg });
      }
    } finally {
      externalJavaArgsLoading.value = false;
    }
  }

  return {
    editingPort,
    tempPort,
    portLoading,
    savePort,
    editingRestartPolicy,
    tempRestartPolicy,
    restartPolicyLoading,
    saveRestartPolicy,
    editingPublicRcon,
    tempPublicRcon,
    tempRconPort,
    rconLoading,
    savePublicRcon,
    editingRam,
    tempRamGb,
    ramLoading,
    saveRam,
    editingExternalJavaArgs,
    tempExternalJavaArgs,
    newExternalArg,
    externalJavaArgsLoading,
    handlePasteExternalJavaArgs,
    saveExternalJavaArgs,
  };
}
