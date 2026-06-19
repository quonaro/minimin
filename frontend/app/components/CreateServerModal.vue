<script setup lang="ts">
import type { InstancePreview } from "./CreateServerInstanceDrop.vue";

interface World {
  name: string;
  archivePath: string;
}

const emit = defineEmits<{
  (e: "created", serverId: string): void;
  (e: "close"): void;
}>();

const { show: showToast } = useToast();
const { loadForEngine, gameVersionsFor, loaderVersionsFor } = useVersions();

const loading = ref(false);
const mode = ref<"blank" | "instance">("blank");
const instancePreview = ref<InstancePreview | null>(null);
const selectedWorld = ref<World | null>(null);
const instanceFieldsLocked = ref(false);

const form = reactive({
  serverId: "",
  engineType: "VANILLA",
  gameVersion: "",
  loaderVersion: "",
  ramGb: 2,
  gamePort: 25565,
  publicRcon: false,
  rconPort: 25575,
  levelName: "",
  levelSeed: "",
  levelType: "",
});

const gameVersions = computed(() => gameVersionsFor(form.engineType));
const loaderVersions = computed(() =>
  loaderVersionsFor(form.engineType, form.gameVersion),
);
const showLoaderVersion = computed(
  () => form.engineType === "FABRIC" || form.engineType === "FORGE",
);

watch(
  () => form.engineType,
  async () => {
    const oldGameVersion = form.gameVersion;
    const oldLoaderVersion = form.loaderVersion;
    await loadForEngine(form.engineType);
    const versions = gameVersions.value;
    if (!versions.find((v) => v.id === oldGameVersion)) {
      const selected = versions.find((v) => v.stable) || versions[0];
      form.gameVersion = selected ? selected.id : "";
    }
    if (showLoaderVersion.value) {
      const loaders = loaderVersions.value;
      if (!loaders.find((v) => v.id === oldLoaderVersion)) {
        const first = loaders[0];
        form.loaderVersion = first ? first.id : "";
      }
    } else {
      form.loaderVersion = "";
    }
  },
  { immediate: true },
);

watch(
  () => form.gameVersion,
  () => {
    if (!showLoaderVersion.value) return;
    const loaders = loaderVersions.value;
    if (loaders.find((v) => v.id === form.loaderVersion)) return;
    const first = loaders[0];
    if (first) {
      form.loaderVersion = first.id;
    } else {
      form.loaderVersion = "";
    }
  },
);

const levelTypeOptions = [
  { value: "", label: "Default (normal)" },
  { value: "minecraft:normal", label: "Normal" },
  { value: "minecraft:flat", label: "Flat" },
  { value: "minecraft:large_biomes", label: "Large Biomes" },
  { value: "minecraft:amplified", label: "Amplified" },
  { value: "minecraft:single_biome_surface", label: "Single Biome" },
];

async function submit() {
  if (loading.value) return;
  loading.value = true;
  try {
    if (mode.value === "instance" && instancePreview.value) {
      await createFromInstance();
      return;
    }
    await createBlank();
  } catch (err: any) {
    const msg = getApiErrorMessage(err, "Failed to create server");
    showToast("error", "Create failed", { description: msg });
  } finally {
    loading.value = false;
  }
}

async function createBlank() {
  const body: Record<string, any> = {
    serverId: form.serverId || undefined,
    engineType: form.engineType,
    gameVersion: form.gameVersion,
    loaderVersion: form.loaderVersion || undefined,
    ramBytes: form.ramGb * 1024 * 1024 * 1024,
    gamePort: form.gamePort,
    publicRcon: form.publicRcon,
  };
  if (form.publicRcon) {
    body.rconPort = form.rconPort;
  }
  if (form.levelName) body.levelName = form.levelName;
  if (form.levelSeed) body.levelSeed = form.levelSeed;
  if (form.levelType) body.levelType = form.levelType;

  const res = await $fetch<{
    serverId?: string;
    body?: { serverId?: string };
  }>(`/servers`, {
    baseURL: useApiBase(),
    method: "POST",
    credentials: "include",
    body,
  });
  const sid =
    (res && typeof res === "object" && "serverId" in res && res.serverId) ||
    (res && typeof res === "object" && "body" in res && res.body?.serverId) ||
    "";
  showToast("success", "Server created", {
    description: sid ? `Server ${sid} is starting.` : "Server is starting.",
  });
  emit("created", sid);
}

async function createFromInstance() {
  const preview = instancePreview.value;
  if (!preview) return;

  const fd = new FormData();
  fd.append("token", preview.token);
  fd.append("serverId", form.serverId || "");
  fd.append("ramBytes", String(form.ramGb * 1024 * 1024 * 1024));
  fd.append("gamePort", String(form.gamePort));
  fd.append("engineType", form.engineType);
  fd.append("gameVersion", form.gameVersion);
  fd.append("loaderVersion", form.loaderVersion || "");
  fd.append("rconPort", String(form.publicRcon ? form.rconPort : 0));
  fd.append("publicRcon", String(form.publicRcon));
  if (form.levelName) fd.append("levelName", form.levelName);
  if (form.levelSeed) fd.append("levelSeed", form.levelSeed);
  if (form.levelType) fd.append("levelType", form.levelType);
  if (selectedWorld.value) {
    fd.append("world", selectedWorld.value.archivePath);
  }

  const res = await $fetch<{ serverId?: string }>(`/servers/from-instance`, {
    baseURL: useApiBase(),
    method: "POST",
    credentials: "include",
    body: fd,
  });
  const sid =
    (res && typeof res === "object" && "serverId" in res && res.serverId) || "";
  showToast("success", "Server imported", {
    description: sid ? `Server ${sid} is starting.` : "Server is starting.",
  });
  emit("created", sid);
}

function onInstancePreview(preview: InstancePreview) {
  instancePreview.value = preview;
  instanceFieldsLocked.value = true;
  if (preview.engineType) form.engineType = preview.engineType;
  if (preview.gameVersion) form.gameVersion = preview.gameVersion;
  if (preview.loaderVersion) form.loaderVersion = preview.loaderVersion;
  if (preview.instanceName && !form.serverId) {
    form.serverId = preview.instanceName
      .toLowerCase()
      .replace(/[^a-z0-9_-]+/g, "-")
      .replace(/^-+|-+$/g, "")
      .slice(0, 32);
  }
  const worlds = preview.worlds || [];
  const first = worlds[0];
  if (first) {
    selectedWorld.value = first;
    form.levelName = first.name;
  } else {
    selectedWorld.value = null;
  }
}

function unlockInstanceFields() {
  instanceFieldsLocked.value = false;
}

function onInstanceError(message: string) {
  instancePreview.value = null;
  selectedWorld.value = null;
  instanceFieldsLocked.value = false;
  showToast("error", "Instance preview failed", { description: message });
}

function setMode(next: "blank" | "instance") {
  mode.value = next;
  instancePreview.value = null;
  selectedWorld.value = null;
  instanceFieldsLocked.value = false;
}
</script>

<template>
  <Teleport to="body">
    <div
      class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
      @click.self="emit('close')"
    >
      <div
        class="bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 w-full max-w-lg max-h-[90vh] overflow-y-auto"
      >
        <div class="p-6 border-b border-gray-200 dark:border-neutral-700">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">
            Create Server
          </h2>
          <div class="mt-4 flex bg-gray-100 dark:bg-neutral-700 rounded-lg p-1">
            <button
              type="button"
              class="flex-1 px-3 py-1.5 text-sm font-medium rounded-md transition-colors"
              :class="
                mode === 'blank'
                  ? 'bg-white dark:bg-neutral-600 text-gray-900 dark:text-white shadow-sm'
                  : 'text-gray-600 dark:text-neutral-300'
              "
              @click="setMode('blank')"
            >
              Blank Server
            </button>
            <button
              type="button"
              class="flex-1 px-3 py-1.5 text-sm font-medium rounded-md transition-colors"
              :class="
                mode === 'instance'
                  ? 'bg-white dark:bg-neutral-600 text-gray-900 dark:text-white shadow-sm'
                  : 'text-gray-600 dark:text-neutral-300'
              "
              @click="setMode('instance')"
            >
              From Instance
            </button>
          </div>
        </div>

        <div class="p-6 space-y-5">
          <div>
            <label
              class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
            >
              Server ID
            </label>
            <input
              v-model="form.serverId"
              type="text"
              placeholder="Auto-generated if empty"
              class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label
                class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
              >
                Engine
              </label>
              <select
                v-model="form.engineType"
                :disabled="mode === 'instance' && instanceFieldsLocked"
                class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow disabled:opacity-60 disabled:cursor-not-allowed"
              >
                <option value="VANILLA">Vanilla</option>
                <option value="PAPER">Paper</option>
                <option value="FABRIC">Fabric</option>
                <option value="FORGE">Forge</option>
              </select>
            </div>
            <div>
              <label
                class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
              >
                Version
              </label>
              <select
                v-model="form.gameVersion"
                :disabled="mode === 'instance' && instanceFieldsLocked"
                class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow disabled:opacity-60 disabled:cursor-not-allowed"
              >
                <option v-for="v in gameVersions" :key="v.id" :value="v.id">
                  {{ v.id }}
                </option>
              </select>
            </div>
          </div>

          <div v-if="showLoaderVersion" class="grid grid-cols-1 gap-4">
            <div>
              <label
                class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
              >
                Loader Version
              </label>
              <select
                v-model="form.loaderVersion"
                :disabled="mode === 'instance' && instanceFieldsLocked"
                class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow disabled:opacity-60 disabled:cursor-not-allowed"
              >
                <option v-for="v in loaderVersions" :key="v.id" :value="v.id">
                  {{ v.id }}
                </option>
              </select>
            </div>
          </div>

          <div>
            <label
              class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
            >
              RAM (GB)
            </label>
            <number-input v-model="form.ramGb" :min="1" class="w-full" />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label
                class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
              >
                Game Port
              </label>
              <number-input
                v-model="form.gamePort"
                :min="1024"
                :max="65535"
                class="w-full"
              />
            </div>
            <div class="flex items-center gap-2 pt-6">
              <input
                id="publicRcon"
                v-model="form.publicRcon"
                type="checkbox"
                class="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
              />
              <label
                for="publicRcon"
                class="text-sm font-medium text-gray-700 dark:text-neutral-300 cursor-pointer"
              >
                Public RCON
              </label>
            </div>
          </div>

          <div v-if="form.publicRcon" class="grid grid-cols-1 gap-4">
            <div>
              <label
                class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
              >
                RCON Port
              </label>
              <number-input
                v-model="form.rconPort"
                :min="1024"
                :max="65535"
                class="w-full"
              />
            </div>
          </div>

          <div
            v-if="mode === 'blank'"
            class="p-4 rounded-xl border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/10 space-y-4"
          >
            <h3
              class="text-sm font-semibold text-red-600 dark:text-red-400 uppercase tracking-wider"
            >
              World Generation (DANGEROUS)
            </h3>
            <p class="text-xs text-red-500 dark:text-red-400">
              These settings are applied only on first start. Changing them
              later will create a new world.
            </p>

            <div>
              <label
                class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
              >
                Level Name
              </label>
              <input
                v-model="form.levelName"
                type="text"
                placeholder="world"
                class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow"
              />
            </div>

            <div>
              <label
                class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
              >
                Level Seed
              </label>
              <input
                v-model="form.levelSeed"
                type="text"
                placeholder="Leave empty for random"
                class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow"
              />
            </div>

            <div>
              <label
                class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
              >
                Level Type
              </label>
              <select
                v-model="form.levelType"
                class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow"
              >
                <option
                  v-for="opt in levelTypeOptions"
                  :key="opt.value"
                  :value="opt.value"
                >
                  {{ opt.label }}
                </option>
              </select>
            </div>
          </div>

          <div v-if="mode === 'instance'" class="space-y-3">
            <create-server-instance-drop
              @preview="onInstancePreview"
              @error="onInstanceError"
            />

            <div
              v-if="instancePreview"
              class="rounded-xl border border-gray-200 dark:border-neutral-700 bg-gray-50 dark:bg-neutral-700/30 p-4 space-y-2"
            >
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ instancePreview.instanceName || "Instance" }}
                </span>
                <div class="flex items-center gap-2">
                  <span
                    class="text-xs px-2 py-0.5 rounded-full bg-primary/10 text-primary font-medium uppercase"
                  >
                    {{ instancePreview.format }}
                  </span>
                  <button
                    type="button"
                    class="text-xs font-medium text-primary hover:text-primary/80 transition-colors"
                    @click="unlockInstanceFields"
                  >
                    Change
                  </button>
                </div>
              </div>
              <div class="text-xs text-gray-600 dark:text-neutral-300">
                <span v-if="instancePreview.engineType" class="mr-3">
                  <strong>Engine:</strong>
                  {{ instancePreview.engineType }}
                </span>
                <span v-if="instancePreview.gameVersion" class="mr-3">
                  <strong>Version:</strong>
                  {{ instancePreview.gameVersion }}
                </span>
                <span v-if="instancePreview.loaderVersion">
                  <strong>Loader:</strong>
                  {{ instancePreview.loaderVersion }}
                </span>
              </div>
              <div
                v-if="(instancePreview.detectedPaths || []).length > 0"
                class="text-xs text-gray-500 dark:text-neutral-400"
              >
                Detected: {{ (instancePreview.detectedPaths || []).join(", ") }}
              </div>
              <div
                v-if="(instancePreview.worlds || []).length > 0"
                class="pt-2"
              >
                <label
                  class="block text-xs font-medium text-gray-700 dark:text-neutral-300 mb-1"
                >
                  Import world
                </label>
                <select
                  v-model="selectedWorld"
                  class="w-full text-xs px-2 py-1.5 rounded border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
                >
                  <option
                    v-for="w in instancePreview.worlds"
                    :key="w.archivePath"
                    :value="w"
                  >
                    {{ w.name }}
                  </option>
                </select>
                <p class="text-xs text-gray-500 dark:text-neutral-400 mt-1">
                  Will be used as level name:
                  <strong>{{ selectedWorld?.name || "world" }}</strong>
                </p>
              </div>
            </div>
          </div>
        </div>

        <div
          class="p-6 border-t border-gray-200 dark:border-neutral-700 flex gap-3 justify-end"
        >
          <button
            class="px-4 py-2 rounded-lg text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors font-medium"
            @click="emit('close')"
          >
            Cancel
          </button>
          <button
            :disabled="loading || (mode === 'instance' && !instancePreview)"
            class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
            @click="submit"
          >
            {{
              loading
                ? "Creating..."
                : mode === "instance"
                  ? "Import Server"
                  : "Create Server"
            }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
