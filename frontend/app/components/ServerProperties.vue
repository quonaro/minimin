<script setup lang="ts">
import type { Component } from "vue";
import {
  ChevronRight,
  Globe,
  Swords,
  Mountain,
  Zap,
  Settings,
  ShieldAlert,
  MoreHorizontal,
} from "lucide-vue-next";

const props = defineProps<{
  agentId: string;
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
}

type PropType = "text" | "number" | "boolean" | "select";

interface KnownProperty {
  key: string;
  label: string;
  type: PropType;
  options?: string[];
  group: string;
  dangerous?: boolean;
}

const configLoading = ref(true);
const configError = ref<string | null>(null);
const originalProperties = ref<Record<string, string>>({});
const editedProperties = ref<Record<string, string>>({});
const saveLoading = ref(false);
const showRestartBanner = ref(false);
const unlockedDangerous = ref<Set<string>>(new Set());

function unlockDangerous(key: string) {
  unlockedDangerous.value.add(key);
}

const knownProperties: KnownProperty[] = [
  {
    key: "level-name",
    label: "Level Name",
    type: "text",
    group: "DANGEROUS",
  },
  {
    key: "level-seed",
    label: "Level Seed",
    type: "text",
    group: "DANGEROUS",
  },
  {
    key: "level-type",
    label: "Level Type",
    type: "select",
    options: [
      "minecraft:normal",
      "minecraft:flat",
      "minecraft:large_biomes",
      "minecraft:amplified",
      "minecraft:single_biome_surface",
    ],
    group: "DANGEROUS",
  },

  { key: "motd", label: "MOTD", type: "text", group: "Network" },
  {
    key: "max-players",
    label: "Max Players",
    type: "number",
    group: "Network",
  },
  {
    key: "server-port",
    label: "Server Port",
    type: "number",
    group: "Network",
    dangerous: true,
  },
  {
    key: "online-mode",
    label: "Online Mode",
    type: "boolean",
    group: "Network",
    dangerous: true,
  },
  {
    key: "white-list",
    label: "White List",
    type: "boolean",
    group: "Network",
    dangerous: true,
  },
  {
    key: "enforce-secure-profile",
    label: "Enforce Secure Profile",
    type: "boolean",
    group: "Network",
    dangerous: true,
  },

  {
    key: "difficulty",
    label: "Difficulty",
    type: "select",
    options: ["peaceful", "easy", "normal", "hard"],
    group: "Gameplay",
  },
  {
    key: "gamemode",
    label: "Game Mode",
    type: "select",
    options: ["survival", "creative", "adventure", "spectator"],
    group: "Gameplay",
  },
  { key: "pvp", label: "PvP", type: "boolean", group: "Gameplay" },
  {
    key: "hardcore",
    label: "Hardcore",
    type: "boolean",
    group: "Gameplay",
    dangerous: true,
  },
  {
    key: "allow-flight",
    label: "Allow Flight",
    type: "boolean",
    group: "Gameplay",
  },
  {
    key: "spawn-protection",
    label: "Spawn Protection",
    type: "number",
    group: "Gameplay",
  },

  {
    key: "generate-structures",
    label: "Generate Structures",
    type: "boolean",
    group: "World",
  },
  {
    key: "allow-nether",
    label: "Allow Nether",
    type: "boolean",
    group: "World",
  },

  {
    key: "view-distance",
    label: "View Distance",
    type: "number",
    group: "Performance",
  },
  {
    key: "simulation-distance",
    label: "Simulation Distance",
    type: "number",
    group: "Performance",
  },

  {
    key: "enable-command-block",
    label: "Enable Command Block",
    type: "boolean",
    group: "Advanced",
  },
];

const selectOptions: Record<string, string[]> = {};
for (const kp of knownProperties) {
  if (kp.type === "select" && kp.options) {
    selectOptions[kp.key] = kp.options;
  } else if (kp.type === "boolean") {
    selectOptions[kp.key] = ["true", "false"];
  }
}

function parseProperties(content: string): Record<string, string> {
  const props: Record<string, string> = {};
  for (const line of content.split("\n")) {
    const trimmed = line.trim();
    if (trimmed.startsWith("#") || trimmed === "") continue;
    const idx = trimmed.indexOf("=");
    if (idx === -1) continue;
    const key = trimmed.slice(0, idx).trim();
    const val = trimmed.slice(idx + 1);
    if (key) props[key] = val;
  }
  return props;
}

const knownKeys = new Set(knownProperties.map((k) => k.key));

const groupOrder = ["Network", "Gameplay", "World", "Performance", "Advanced"];

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

watch(
  () => changedProperties.value,
  () => {
    showRestartBanner.value = false;
  },
  { deep: true },
);

async function loadConfig() {
  configLoading.value = true;
  configError.value = null;
  try {
    const res = await $fetch<ConfigResponse | { body: ConfigResponse }>(
      `/agent/${props.agentId}/servers/${props.serverId}/config`,
      { baseURL: useApiBase(), credentials: "include" },
    );
    let content = "";
    if (res && typeof res === "object") {
      if ("body" in res) {
        content = (res as any).body.content || "";
      } else if ("content" in res) {
        content = (res as any).content || "";
      }
    }
    const parsed = parseProperties(content);
    originalProperties.value = parsed;
    editedProperties.value = { ...parsed };
  } catch (err: any) {
    configError.value = err?.message || "Unknown error";
  } finally {
    configLoading.value = false;
  }
}

async function saveProperties() {
  if (saveLoading.value) return;
  saveLoading.value = true;
  try {
    await $fetch(`/agent/${props.agentId}/servers/${props.serverId}/config`, {
      baseURL: useApiBase(),
      method: "PATCH",
      credentials: "include",
      body: { properties: changedProperties.value },
    });
    originalProperties.value = { ...editedProperties.value };
    showToast("success", "Properties saved", {
      description: "Configuration updated successfully.",
    });
    showRestartBanner.value = true;
  } catch (err: any) {
    const msg =
      err?.data?.detail || err?.message || "Failed to save properties";
    showToast("error", "Save failed", { description: msg });
  } finally {
    saveLoading.value = false;
  }
}

await loadConfig();
</script>

<template>
  <div>
    <div v-if="configLoading" class="text-gray-500 dark:text-neutral-400">
      Loading config...
    </div>
    <div v-else-if="configError" class="text-red-500 dark:text-red-400">
      Failed to load config: {{ configError }}
    </div>
    <div v-else class="space-y-8">
      <template v-for="group in groupOrder" :key="group">
        <div v-if="groupedItems.groups[group]?.length" class="space-y-4">
          <button
            class="flex items-center gap-2 w-full text-left px-3 py-2 rounded-lg bg-gray-100 dark:bg-neutral-700/50 hover:bg-gray-200 dark:hover:bg-neutral-700 transition-colors"
            @click="toggleGroup(group)"
          >
            <component
              :is="groupIcons[group]"
              class="w-4 h-4 shrink-0 text-gray-500 dark:text-neutral-400"
            />
            <span
              class="text-sm font-semibold text-gray-700 dark:text-neutral-300 uppercase tracking-wider"
            >
              {{ group }}
            </span>
            <ChevronRight
              class="w-4 h-4 transition-transform shrink-0 ml-auto text-gray-500 dark:text-neutral-400"
              :class="collapsedGroups.has(group) ? '' : 'rotate-90'"
            />
          </button>
          <div v-show="!collapsedGroups.has(group)" class="space-y-4">
            <div
              v-for="item in groupedItems.groups[group]"
              :key="item.key"
              class="grid grid-cols-1 md:grid-cols-3 gap-4 items-start"
            >
              <div class="md:col-span-1 pt-2">
                <div class="flex items-center gap-2">
                  <label
                    class="block text-sm font-medium text-gray-700 dark:text-neutral-300"
                  >
                    {{ item.label }}
                  </label>
                  <button
                    v-if="item.dangerous && !unlockedDangerous.has(item.key)"
                    class="text-xs text-primary hover:underline font-medium"
                    @click="unlockDangerous(item.key)"
                  >
                    Unlock
                  </button>
                </div>
                <p
                  v-if="item.dangerous && unlockedDangerous.has(item.key)"
                  class="mt-1 text-xs text-red-500 dark:text-red-400"
                >
                  This setting can break your server. Change with caution.
                </p>
              </div>
              <div class="md:col-span-2">
                <select
                  v-if="item.type === 'select' || item.type === 'boolean'"
                  v-model="editedProperties[item.key]"
                  :disabled="item.dangerous && !unlockedDangerous.has(item.key)"
                  class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <option
                    v-for="opt in selectOptions[item.key]"
                    :key="opt"
                    :value="opt"
                  >
                    {{ opt }}
                  </option>
                </select>
                <input
                  v-else-if="item.type === 'number'"
                  v-model="editedProperties[item.key]"
                  type="number"
                  min="0"
                  :disabled="item.dangerous && !unlockedDangerous.has(item.key)"
                  class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow disabled:opacity-50 disabled:cursor-not-allowed"
                />
                <input
                  v-else
                  v-model="editedProperties[item.key]"
                  type="text"
                  :disabled="item.dangerous && !unlockedDangerous.has(item.key)"
                  class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow disabled:opacity-50 disabled:cursor-not-allowed"
                />
              </div>
            </div>
          </div>
        </div>
      </template>

      <div v-if="groupedItems.unknown.length > 0">
        <button
          class="flex items-center gap-2 w-full text-left px-3 py-2 rounded-lg bg-gray-100 dark:bg-neutral-700/50 hover:bg-gray-200 dark:hover:bg-neutral-700 transition-colors mb-3"
          @click="toggleGroup('unknown')"
        >
          <MoreHorizontal
            class="w-4 h-4 shrink-0 text-gray-500 dark:text-neutral-400"
          />
          <span
            class="text-sm font-semibold text-gray-700 dark:text-neutral-300 uppercase tracking-wider"
          >
            Other Properties
          </span>
          <ChevronRight
            class="w-4 h-4 transition-transform shrink-0 ml-auto text-gray-500 dark:text-neutral-400"
            :class="collapsedGroups.has('unknown') ? '' : 'rotate-90'"
          />
        </button>
        <div v-show="!collapsedGroups.has('unknown')" class="space-y-4">
          <div
            v-for="item in groupedItems.unknown"
            :key="item.key"
            class="grid grid-cols-1 md:grid-cols-3 gap-4 items-start"
          >
            <div class="md:col-span-1 pt-2">
              <label
                class="block text-sm font-medium text-gray-700 dark:text-neutral-300"
              >
                {{ item.label }}
              </label>
            </div>
            <div class="md:col-span-2">
              <input
                v-model="editedProperties[item.key]"
                type="text"
                class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow"
              />
            </div>
          </div>
        </div>
      </div>

      <div v-if="groupedItems.dangerous.length > 0">
        <button
          class="flex items-center gap-2 w-full text-left px-3 py-2 rounded-lg bg-red-100 dark:bg-red-900/20 hover:bg-red-200 dark:hover:bg-red-900/30 transition-colors mb-3"
          @click="toggleGroup('dangerous')"
        >
          <ShieldAlert
            class="w-4 h-4 shrink-0 text-red-600 dark:text-red-400"
          />
          <span
            class="text-sm font-semibold text-red-700 dark:text-red-300 uppercase tracking-wider"
          >
            World Generation (DANGEROUS)
          </span>
          <ChevronRight
            class="w-4 h-4 transition-transform shrink-0 ml-auto text-red-600 dark:text-red-400"
            :class="collapsedGroups.has('dangerous') ? '' : 'rotate-90'"
          />
        </button>
        <div v-show="!collapsedGroups.has('dangerous')" class="space-y-4">
          <div
            v-for="item in groupedItems.dangerous"
            :key="item.key"
            class="grid grid-cols-1 md:grid-cols-3 gap-4 items-start"
          >
            <div class="md:col-span-1 pt-2">
              <label
                class="block text-sm font-medium text-gray-700 dark:text-neutral-300"
              >
                {{ item.label }}
              </label>
            </div>
            <div class="md:col-span-2">
              <select
                v-if="item.type === 'select' || item.type === 'boolean'"
                v-model="editedProperties[item.key]"
                class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow"
              >
                <option
                  v-for="opt in selectOptions[item.key]"
                  :key="opt"
                  :value="opt"
                >
                  {{ opt }}
                </option>
              </select>
              <input
                v-else-if="item.type === 'number'"
                v-model="editedProperties[item.key]"
                type="number"
                min="0"
                class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow"
              />
              <input
                v-else
                v-model="editedProperties[item.key]"
                type="text"
                class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow"
              />
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="showRestartBanner"
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
