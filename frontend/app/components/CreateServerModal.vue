<script setup lang="ts">
const props = defineProps<{
  agentId: string;
}>();

const emit = defineEmits<{
  (e: "created", serverId: string): void;
  (e: "close"): void;
}>();

const { show: showToast } = useToast();
const { loadForEngine, gameVersionsFor, loaderVersionsFor } = useVersions();

const loading = ref(false);

const form = reactive({
  serverId: "",
  engineType: "VANILLA",
  gameVersion: "",
  loaderVersion: "",
  ramGb: 2,
  cpus: 1,
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
    form.loaderVersion = "";
    await loadForEngine(form.engineType);
    const versions = gameVersions.value;
    const selected = versions.find((v) => v.stable) || versions[0];
    form.gameVersion = selected ? selected.id : "";
    if (showLoaderVersion.value) {
      const loaders = loaderVersions.value;
      const first = loaders[0];
      form.loaderVersion = first ? first.id : "";
    }
  },
  { immediate: true },
);

watch(
  () => form.gameVersion,
  () => {
    if (!showLoaderVersion.value) return;
    const loaders = loaderVersions.value;
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
    const body: Record<string, any> = {
      serverId: form.serverId || undefined,
      engineType: form.engineType,
      gameVersion: form.gameVersion,
      loaderVersion: form.loaderVersion || undefined,
      ramBytes: form.ramGb * 1024 * 1024 * 1024,
      cpus: form.cpus,
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
    }>(`/agent/${props.agentId}/servers`, {
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
  } catch (err: any) {
    const msg = err?.data?.detail || err?.message || "Failed to create server";
    showToast("error", "Create failed", { description: msg });
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    @click.self="emit('close')"
  >
    <div
      class="bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 w-full max-w-lg max-h-[90vh] overflow-y-auto"
    >
      <div class="p-6 border-b border-gray-200 dark:border-neutral-700">
        <h2 class="text-xl font-bold text-gray-900 dark:text-white">
          Create Server
        </h2>
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
              class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow"
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
              class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow"
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
              class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-shadow"
            >
              <option v-for="v in loaderVersions" :key="v.id" :value="v.id">
                {{ v.id }}
              </option>
            </select>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label
              class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
            >
              RAM (GB)
            </label>
            <number-input v-model="form.ramGb" :min="1" class="w-full" />
          </div>
          <div>
            <label
              class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
            >
              CPUs
            </label>
            <number-input
              v-model="form.cpus"
              :min="0.5"
              :step="0.5"
              class="w-full"
            />
          </div>
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
          class="p-4 rounded-xl border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/10 space-y-4"
        >
          <h3
            class="text-sm font-semibold text-red-600 dark:text-red-400 uppercase tracking-wider"
          >
            World Generation (DANGEROUS)
          </h3>
          <p class="text-xs text-red-500 dark:text-red-400">
            These settings are applied only on first start. Changing them later
            will create a new world.
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
          :disabled="loading"
          class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          @click="submit"
        >
          {{ loading ? "Creating..." : "Create Server" }}
        </button>
      </div>
    </div>
  </div>
</template>
