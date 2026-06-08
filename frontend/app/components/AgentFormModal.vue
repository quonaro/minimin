<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="modelValue"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
        @click.self="close"
      >
        <Transition
          enter-active-class="transition-transform duration-200"
          enter-from-class="scale-95 opacity-0"
          enter-to-class="scale-100 opacity-100"
          leave-active-class="transition-transform duration-200"
          leave-from-class="scale-100 opacity-100"
          leave-to-class="scale-95 opacity-0"
        >
          <div
            v-if="modelValue"
            class="bg-white dark:bg-background-dark rounded-2xl shadow-xl border border-gray-200 dark:border-border-dark w-full max-w-md mx-4 p-6"
          >
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-xl font-bold text-gray-900 dark:text-white">
                Register Agent
              </h2>
              <button
                @click="close"
                class="p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-white/10 transition-colors text-gray-500 dark:text-neutral-400"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  class="w-5 h-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>

            <form @submit.prevent="handleSubmit" class="space-y-4">
              <div>
                <label
                  for="agent-name"
                  class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
                >
                  Name
                </label>
                <input
                  id="agent-name"
                  v-model="form.name"
                  type="text"
                  required
                  placeholder="My Agent"
                  class="w-full px-3 py-2 rounded-lg bg-gray-50 dark:bg-white/5 border border-gray-300 dark:border-border-dark text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary text-sm"
                />
              </div>

              <div>
                <label
                  for="agent-host"
                  class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
                >
                  Host
                </label>
                <input
                  id="agent-host"
                  v-model="form.host"
                  type="url"
                  required
                  placeholder="http://localhost:8080"
                  class="w-full px-3 py-2 rounded-lg bg-gray-50 dark:bg-white/5 border border-gray-300 dark:border-border-dark text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary text-sm"
                />
              </div>

              <div>
                <label
                  for="agent-api-key"
                  class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1"
                >
                  API Key
                </label>
                <input
                  id="agent-api-key"
                  v-model="form.apiKey"
                  type="text"
                  required
                  placeholder="agent-api-key"
                  class="w-full px-3 py-2 rounded-lg bg-gray-50 dark:bg-white/5 border border-gray-300 dark:border-border-dark text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary text-sm"
                />
              </div>

              <div
                v-if="checkStatus === 'success'"
                class="text-sm text-green-500 dark:text-green-400"
              >
                Connection OK
              </div>
              <div
                v-else-if="checkStatus === 'error'"
                class="text-sm text-red-500 dark:text-red-400"
              >
                {{ checkError || "Connection failed" }}
              </div>
              <div
                v-else-if="error"
                class="text-sm text-red-500 dark:text-red-400"
              >
                {{ error }}
              </div>

              <div class="flex gap-3 pt-2">
                <template v-if="!confirmShown">
                  <button
                    type="button"
                    @click="close"
                    class="flex-1 px-4 py-2 rounded-lg border border-gray-300 dark:border-border-dark text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-white/10 transition-colors text-sm font-medium"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    :disabled="checking"
                    @click="checkConnection"
                    class="flex-1 px-4 py-2 rounded-lg border border-gray-300 dark:border-border-dark text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-white/10 transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {{ checking ? "Checking..." : "Check" }}
                  </button>
                  <button
                    v-if="checkStatus !== 'idle'"
                    type="submit"
                    :disabled="loading"
                    :class="[
                      'flex-1 px-4 py-2 rounded-lg text-white transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed',
                      checkStatus === 'error'
                        ? 'bg-orange-500 hover:bg-orange-600'
                        : 'bg-primary hover:bg-primary/90',
                    ]"
                  >
                    {{ loading ? "Registering..." : "Register" }}
                  </button>
                </template>
                <template v-else>
                  <div
                    class="flex-1 text-sm text-orange-500 dark:text-orange-400 self-center"
                  >
                    Agent is not responding. Register anyway?
                  </div>
                  <button
                    type="button"
                    @click="confirmShown = false"
                    class="px-4 py-2 rounded-lg border border-gray-300 dark:border-border-dark text-gray-700 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-white/10 transition-colors text-sm font-medium"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    :disabled="loading"
                    @click="submit"
                    class="px-4 py-2 rounded-lg bg-orange-500 text-white hover:bg-orange-600 transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {{ loading ? "Registering..." : "Register Anyway" }}
                  </button>
                </template>
              </div>
            </form>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
}>();

const left = [
  "admiring",
  "adoring",
  "affectionate",
  "amazing",
  "angry",
  "beautiful",
  "blissful",
  "bold",
  "brave",
  "busy",
  "calm",
  "charming",
  "clever",
  "cool",
  "confident",
  "crazy",
  "curious",
  "determined",
  "eager",
  "elegant",
  "energetic",
  "excited",
  "fierce",
  "friendly",
  "funny",
  "gentle",
  "gifted",
  "happy",
  "hopeful",
  "hungry",
  "interesting",
  "jolly",
  "keen",
  "kind",
  "laughing",
  "loving",
  "modest",
  "nice",
  "optimistic",
  "peaceful",
  "practical",
  "quiet",
  "relaxed",
  "romantic",
  "sad",
  "serious",
  "silly",
  "sleepy",
  "smart",
  "strong",
  "sweet",
  "tender",
  "thirsty",
  "vibrant",
  "witty",
  "wonderful",
  "youthful",
  "zealous",
];

const right = [
  "babbage",
  "banach",
  "bartik",
  "bell",
  "blackwell",
  "bohr",
  "booth",
  "boyd",
  "brown",
  "buck",
  "burnell",
  "carson",
  "chandrasekhar",
  "church",
  "clarke",
  "curie",
  "darwin",
  "davinci",
  "dijkstra",
  "dirac",
  "dubinsky",
  "easley",
  "edison",
  "einstein",
  "elion",
  "engelbart",
  "euler",
  "feynman",
  "franklin",
  "galileo",
  "gates",
  "goldstine",
  "goodall",
  "gould",
  "greider",
  "hamilton",
  "hawking",
  "heisenberg",
  "hill",
  "hodgkin",
  "hoover",
  "hopper",
  "hugle",
  "hurd",
  "jackson",
  "jemison",
  "jennings",
  "jepsen",
  "johnson",
  "joliot",
  "jump",
  "kalam",
  "kapitsa",
  "kepler",
  "khorana",
  "kilby",
  "kirch",
  "knuth",
  "kowalevski",
  "lalande",
  "lamarr",
  "leakey",
  "leavitt",
  "lewin",
  "lichterman",
  "liskov",
  "lovelace",
  "lumiere",
  "mahavira",
  "mandelbrot",
  "maxwell",
  "mayer",
  "mccarthy",
  "mcclintock",
  "mclean",
  "mcnulty",
  "mendel",
  "mendeleev",
  "meitner",
  "milton",
  "mirzakhani",
  "morse",
  "napier",
  "nash",
  "neumann",
  "newton",
  "nobel",
  "noether",
  "northcutt",
  "noyce",
  "panini",
  "pare",
  "pasteur",
  "pauli",
  "payne",
  "perlman",
  "pike",
  "poincare",
  "polanyi",
  "ptolemy",
  "raman",
  "ramanujan",
  "ride",
  "ritchie",
  "roentgen",
  "rosario",
  "rubin",
  "saha",
  "sammet",
  "schrodinger",
  "shannon",
  "shaw",
  "sinoussi",
  "snell",
  "spence",
  "stallman",
  "stone",
  "swanson",
  "swartz",
  "tesla",
  "tharp",
  "thompson",
  "torvalds",
  "turing",
  "varahamihira",
  "visvesvaraya",
  "volhard",
  "wescoff",
  "wilson",
  "wozniak",
  "wright",
  "yalow",
  "yonath",
];

function generateDockerName() {
  const a = left[Math.floor(Math.random() * left.length)];
  const b = right[Math.floor(Math.random() * right.length)];
  return `${a}_${b}`;
}

const form = reactive({
  name: "",
  host: "",
  apiKey: "",
});

const loading = ref(false);
const error = ref("");
const checking = ref(false);
const checkStatus = ref<"idle" | "success" | "error">("idle");
const checkError = ref("");
const confirmShown = ref(false);

function close() {
  emit("update:modelValue", false);
  resetForm();
}

function resetForm() {
  form.name = generateDockerName();
  form.host = "";
  form.apiKey = "";
  error.value = "";
  loading.value = false;
  checking.value = false;
  checkStatus.value = "idle";
  checkError.value = "";
  confirmShown.value = false;
}

const { show: showToast } = useToast();

async function checkConnection() {
  if (!form.host || !form.apiKey) {
    checkStatus.value = "error";
    checkError.value = "Host and API key are required";
    return;
  }
  confirmShown.value = false;
  checking.value = true;
  checkStatus.value = "idle";
  checkError.value = "";
  error.value = "";

  try {
    const result = await $fetch<{ valid: boolean; error?: string }>(
      "/agents/check",
      {
        baseURL: useApiBase(),
        method: "POST",
        credentials: "include",
        body: {
          host: form.host,
          apiKey: form.apiKey,
        },
      },
    );
    if (result.valid) {
      checkStatus.value = "success";
      showToast("success", "Connection OK", {
        description: `Agent at ${form.host} responded.`,
      });
    } else {
      checkStatus.value = "error";
      checkError.value = result.error || "Connection failed";
      showToast("error", "Connection failed", {
        description: checkError.value,
      });
    }
  } catch (err: any) {
    checkStatus.value = "error";
    checkError.value =
      err?.data?.detail || err?.message || "Could not reach agent";
    showToast("error", "Connection failed", { description: checkError.value });
  } finally {
    checking.value = false;
  }
}

function handleSubmit() {
  if (checkStatus.value === "success" || confirmShown.value) {
    submit();
  } else if (checkStatus.value === "error") {
    confirmShown.value = true;
  } else {
    checkStatus.value = "error";
    checkError.value = "Please check the connection first";
  }
}

async function submit() {
  loading.value = true;
  error.value = "";

  try {
    await $fetch("/agents", {
      baseURL: useApiBase(),
      method: "POST",
      credentials: "include",
      body: {
        name: form.name,
        host: form.host,
        apiKey: form.apiKey,
      },
    });

    showToast("success", "Agent registered", {
      description: `${form.name} was added successfully.`,
    });
    await refreshNuxtData("agents");
    close();
  } catch (err: any) {
    error.value =
      err?.data?.detail || err?.message || "Failed to register agent";
    showToast("error", "Failed to register agent", {
      description: error.value,
    });
  } finally {
    loading.value = false;
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) resetForm();
  },
);
</script>
