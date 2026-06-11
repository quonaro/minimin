<script setup lang="ts">
import type { Component } from "vue";
import { watch } from "vue";
import {
  Globe,
  Swords,
  Mountain,
  Zap,
  Settings,
  ShieldAlert,
  MoreHorizontal,
} from "lucide-vue-next";
import {
  knownProperties,
  parseProperties,
  selectOptions,
  knownKeys,
  groupOrder,
  type PropType,
} from "~/utils/serverProperties";

const props = defineProps<{
  serverId: string;
}>();

const groupIcons: Record<string, Component> = {
  Network: Globe,
  Gameplay: Swords,
  World: Mountain,
  Performance: Zap,
  Advanced: Settings,
};

const { show: showToast } = useToast();

interface ConfigResponse {
  content: string;
  initialized?: boolean;
  pendingProperties?: Record<string, string>;
}

const { config: serverConfig, refresh: refreshConfig } = useServerConfig(
  props.serverId,
);

const configLoading = ref(true);
const configError = ref<string | null>(null);
const configUninitialized = ref(false);
const originalProperties = ref<Record<string, string>>({});
const editedProperties = ref<Record<string, string>>({});
const saveLoading = ref(false);
const unlockedDangerous = reactive<Set<string>>(new Set());
const pendingProperties = ref<Record<string, string>>({});

function unlockDangerous(key: string) {
  unlockedDangerous.add(key);
}

const collapsedGroups = ref<Set<string>>(new Set(["dangerous", "unknown"]));

function toggleGroup(group: string) {
  const next = new Set(collapsedGroups.value);
  if (next.has(group)) {
    next.delete(group);
  } else {
    next.add(group);
  }
  collapsedGroups.value = next;
}

const groupedItems = computed(() => {
  const groups: Record<
    string,
    { key: string; label: string; type: PropType; dangerous?: boolean }[]
  > = {};
  const dangerous: { key: string; label: string; type: PropType }[] = [];
  const unknown: {
    key: string;
    label: string;
    type: PropType;
    dangerous?: boolean;
  }[] = [];
  const props = editedProperties.value;

  for (const kp of knownProperties) {
    if (!(kp.key in props)) continue;
    if (kp.group === "DANGEROUS") {
      dangerous.push({ key: kp.key, label: kp.label, type: kp.type });
      continue;
    }
    const g = kp.group;
    if (!groups[g]) groups[g] = [];
    groups[g].push({
      key: kp.key,
      label: kp.label,
      type: kp.type,
      dangerous: kp.dangerous,
    });
  }

  for (const key of Object.keys(props).sort()) {
    if (!knownKeys.has(key)) {
      unknown.push({ key, label: key, type: "text" });
    }
  }

  return { groups, dangerous, unknown };
});

const changedProperties = computed(() => {
  const changed: Record<string, string> = {};
  for (const key of Object.keys(editedProperties.value)) {
    if (editedProperties.value[key] !== originalProperties.value[key]) {
      changed[key] = editedProperties.value[key]!;
    }
  }
  return changed;
});

function getPropertyBadgeStatus(
  key: string,
): "modified" | "restart-required" | null {
  if (pendingProperties.value[key] !== undefined) {
    return "restart-required";
  }
  if (changedProperties.value[key] !== undefined) {
    return "modified";
  }
  return null;
}

async function saveProperties() {
  if (saveLoading.value) return;
  saveLoading.value = true;
  try {
    await $fetch(`/servers/${props.serverId}/config`, {
      baseURL: useApiBase(),
      method: "PATCH",
      credentials: "include",
      body: { properties: changedProperties.value },
    });
    const changedKeys = Object.keys(changedProperties.value);
    originalProperties.value = { ...editedProperties.value };
    for (const key of changedKeys) {
      pendingProperties.value[key] = editedProperties.value[key]!;
    }
    showToast("success", "Properties saved", {
      description: "Configuration updated successfully.",
    });
  } catch (err: any) {
    const status = err?.response?.status || err?.statusCode;
    let msg = err?.data?.detail || err?.message || "Failed to save properties";
    if (status === 409) {
      msg = "Server volume not initialized. Start the server first.";
    }
    showToast("error", "Save failed", { description: msg });
  } finally {
    saveLoading.value = false;
  }
}

watch(
  () => serverConfig.value,
  (cfg) => {
    if (!cfg) {
      configLoading.value = true;
      return;
    }
    configLoading.value = false;
    configError.value = null;
    if (!cfg.initialized) {
      configUninitialized.value = true;
      originalProperties.value = {};
      editedProperties.value = {};
      pendingProperties.value = cfg.pendingProperties || {};
    } else {
      configUninitialized.value = false;
      const parsed = parseProperties(cfg.content);
      originalProperties.value = parsed;
      editedProperties.value = { ...parsed };
      pendingProperties.value = cfg.pendingProperties || {};
    }
  },
  { immediate: true },
);
</script>

<template>
  <div>
    <div v-if="configLoading" class="text-gray-500 dark:text-neutral-400">
      Loading config...
    </div>
    <div v-else-if="configError" class="text-red-500 dark:text-red-400">
      Failed to load config: {{ configError }}
    </div>
    <div
      v-else-if="configUninitialized"
      class="text-amber-600 dark:text-amber-400"
    >
      <p class="font-medium">Server not initialized</p>
      <p class="text-sm mt-1">
        Start the server at least once so that the configuration files are
        generated.
      </p>
    </div>
    <div v-else class="space-y-8">
      <template v-for="group in groupOrder" :key="group">
        <server-property-group
          v-if="groupedItems.groups[group]?.length"
          :name="group"
          :icon="groupIcons[group]"
          :collapsed="collapsedGroups.has(group)"
          @toggle="toggleGroup(group)"
        >
          <server-property-field
            v-for="item in groupedItems.groups[group]"
            :key="item.key"
            :item-key="item.key"
            :label="item.label"
            :type="item.type"
            :dangerous="item.dangerous"
            :model-value="editedProperties[item.key] ?? ''"
            :badge-status="getPropertyBadgeStatus(item.key)"
            :unlocked="unlockedDangerous.has(item.key)"
            :options="selectOptions[item.key]"
            @update:model-value="editedProperties[item.key] = $event"
            @unlock="unlockDangerous(item.key)"
          />
        </server-property-group>
      </template>

      <server-property-group
        v-if="groupedItems.unknown.length > 0"
        name="Other Properties"
        :icon="MoreHorizontal"
        :collapsed="collapsedGroups.has('unknown')"
        @toggle="toggleGroup('unknown')"
      >
        <server-property-field
          v-for="item in groupedItems.unknown"
          :key="item.key"
          :item-key="item.key"
          :label="item.label"
          :type="item.type"
          :model-value="editedProperties[item.key] ?? ''"
          :badge-status="getPropertyBadgeStatus(item.key)"
          @update:model-value="editedProperties[item.key] = $event"
        />
      </server-property-group>

      <server-property-group
        v-if="groupedItems.dangerous.length > 0"
        name="World Generation (DANGEROUS)"
        :icon="ShieldAlert"
        :collapsed="collapsedGroups.has('dangerous')"
        dangerous
        @toggle="toggleGroup('dangerous')"
      >
        <server-property-field
          v-for="item in groupedItems.dangerous"
          :key="item.key"
          :item-key="item.key"
          :label="item.label"
          :type="item.type"
          :model-value="editedProperties[item.key] ?? ''"
          :badge-status="getPropertyBadgeStatus(item.key)"
          :options="selectOptions[item.key]"
          @update:model-value="editedProperties[item.key] = $event"
        />
      </server-property-group>

      <div
        v-if="Object.keys(pendingProperties).length > 0"
        class="p-4 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg"
      >
        <p class="text-sm text-amber-800 dark:text-amber-300">
          Saved properties require a server restart to take effect.
        </p>
      </div>

      <div class="flex gap-2">
        <button
          :disabled="saveLoading || Object.keys(changedProperties).length === 0"
          class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          @click="saveProperties"
        >
          Save Properties
        </button>
      </div>
    </div>
  </div>
</template>
