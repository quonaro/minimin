<template>
  <div class="p-8">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">
          Agents
        </h1>
        <p class="text-gray-600 dark:text-neutral-400">
          Manage your agent connections
        </p>
      </div>
      <button
        @click="showForm = true"
        class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors font-medium"
      >
        Add Agent
      </button>
    </div>

    <div v-if="agents.length === 0" class="text-center py-12">
      <p class="text-gray-500 dark:text-neutral-400">No agents found</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="agent in agents"
        :key="agent.id"
        class="relative bg-white dark:bg-neutral-800 rounded-xl shadow-sm border border-gray-200 dark:border-neutral-700 p-6 hover:shadow-md transition-shadow"
      >
        <button
          :disabled="deleteLoading[agent.id]"
          class="absolute top-4 right-4 text-gray-400 hover:text-red-500 transition-colors disabled:opacity-50"
          title="Delete agent"
          @click.stop="deleteAgent(agent)"
        >
          <Trash2 class="w-4 h-4" />
        </button>

        <div class="mb-4 pr-6">
          <NuxtLink
            :to="`/agent/${agent.id}/`"
            class="text-lg font-semibold text-gray-900 dark:text-white mb-1 hover:text-primary transition-colors block"
          >
            {{ agent.name }}
          </NuxtLink>
          <p class="text-sm text-gray-500 dark:text-neutral-400">
            {{ agent.host }}
          </p>
        </div>

        <div class="space-y-2 mb-4">
          <div
            class="flex items-center gap-2 text-sm text-gray-600 dark:text-neutral-400"
          >
            <Calendar class="w-4 h-4" />
            <span
              >Created:
              {{ new Date(agent.createdAt).toLocaleDateString() }}</span
            >
          </div>
        </div>

        <NuxtLink
          :to="`/agent/${agent.id}/`"
          class="inline-flex items-center gap-2 text-primary font-medium text-sm hover:underline"
        >
          <span>Open agent</span>
          <ChevronRight class="w-4 h-4" />
        </NuxtLink>
      </div>
    </div>
    <AgentFormModal v-model="showForm" />
  </div>
</template>

<script setup lang="ts">
import { Calendar, ChevronRight, Trash2 } from "lucide-vue-next";

definePageMeta({
  middleware: "auth",
});

interface Agent {
  id: string;
  name: string;
  host: string;
  createdAt: string;
}

const showForm = ref(false);
const { show: showToast } = useToast();

const { data: agentsData, refresh } = await useFetch("/agents", {
  baseURL: useApiBase(),
  credentials: "include",
  key: "agents",
});

const agents = computed<Agent[]>(() => {
  const val = agentsData.value;
  if (Array.isArray(val)) return val as Agent[];
  if (val && typeof val === "object" && "body" in val) {
    return (val as any).body || [];
  }
  return [];
});

const deleteLoading = reactive<Record<string, boolean>>({});

async function deleteAgent(agent: Agent) {
  const confirmed = window.confirm(
    `Delete agent "${agent.name}"? This will only remove the connection from the orchestrator; it will not stop or delete any servers on the agent itself.`,
  );
  if (!confirmed) return;

  deleteLoading[agent.id] = true;
  try {
    await $fetch(`/agents/${agent.id}`, {
      baseURL: useApiBase(),
      method: "DELETE",
      credentials: "include",
    });
    showToast("success", "Agent deleted", {
      description: `"${agent.name}" has been removed.`,
    });
    await refresh();
  } catch (err: any) {
    const msg = err?.data?.detail || err?.message || "Failed to delete agent";
    showToast("error", "Delete failed", { description: msg });
  } finally {
    deleteLoading[agent.id] = false;
  }
}
</script>
