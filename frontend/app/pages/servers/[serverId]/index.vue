<template>
  <div class="p-6">
    <div v-if="server" class="space-y-6">
      <!-- Main server card -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl p-6 md:p-8 shadow-sm"
      >
        <div class="flex flex-col md:flex-row gap-8">
          <!-- Avatar column -->
          <div class="flex flex-col items-center gap-4 shrink-0">
            <div class="relative">
              <div
                class="w-24 h-24 rounded-2xl overflow-hidden shadow-lg bg-gradient-to-br from-gray-200 to-gray-300 dark:from-neutral-700 dark:to-neutral-600 flex items-center justify-center ring-4 ring-white dark:ring-neutral-800"
              >
                <img
                  v-if="iconUrl && !iconError"
                  :src="iconUrl"
                  alt="Server icon"
                  class="w-full h-full object-cover"
                  @error="iconError = true"
                />
                <ServerIcon
                  v-else
                  class="w-10 h-10 text-indigo-500 dark:text-indigo-400"
                />
              </div>
              <span
                :class="[
                  'absolute -bottom-1 -right-1 w-5 h-5 rounded-full border-2 border-white dark:border-neutral-800',
                  server.serverStatus === 'running'
                    ? 'bg-green-500'
                    : server.serverStatus === 'starting'
                      ? 'bg-yellow-500'
                      : server.containerStatus === 'exited'
                        ? 'bg-red-500'
                        : 'bg-gray-400',
                ]"
              />
            </div>
            <button
              class="text-xs font-medium text-gray-500 dark:text-neutral-400 hover:text-primary transition-colors"
              @click="fileInput?.click()"
            >
              Change Icon
            </button>
            <input
              ref="fileInput"
              type="file"
              accept="image/png,image/jpeg,image/jpg,image/gif,image/webp,image/bmp"
              class="hidden"
              @change="onIconFileSelect"
            />
            <server-icon-editor
              v-model="showIconEditor"
              :file="selectedIconFile"
              @save="uploadProcessedIcon"
            />
          </div>

          <!-- Details column -->
          <div class="flex-1 min-w-0">
            <!-- Title row -->
            <div class="mb-6">
              <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
                {{ server.serverId }}
              </h1>
              <div class="flex items-center gap-3 flex-wrap">
                <span
                  :class="[
                    getStatusColor(server.containerStatus),
                    'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium',
                    server.containerStatus === 'running' &&
                      'animate-heartbeat dark:animate-heartbeat-dark',
                  ]"
                >
                  <Activity
                    v-if="server.containerStatus === 'running'"
                    :class="
                      server.containerStatus === 'running' &&
                      'animate-pulse-icon'
                    "
                    class="w-3.5 h-3.5"
                  />
                  container: {{ server.containerStatus }}
                </span>
                <span
                  :class="[
                    getStatusColor(server.serverStatus),
                    'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium',
                    server.serverStatus === 'running' &&
                      'animate-heartbeat dark:animate-heartbeat-dark',
                  ]"
                >
                  <Activity
                    v-if="server.serverStatus === 'running'"
                    :class="
                      server.serverStatus === 'running' && 'animate-pulse-icon'
                    "
                    class="w-3.5 h-3.5"
                  />
                  server: {{ server.serverStatus }}
                </span>
                <span
                  v-if="server.modCount !== undefined && server.modCount > 0"
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400"
                >
                  <Package class="w-3.5 h-3.5" />
                  {{ server.modCount }} mod{{
                    server.modCount === 1 ? "" : "s"
                  }}
                </span>
                <span
                  v-if="containerUptime"
                  class="text-sm text-gray-500 dark:text-neutral-400 flex items-center gap-1.5"
                  :title="`Started at: ${formatTimestamp(containerStartedAt)}`"
                >
                  <Clock
                    class="w-3.5 h-3.5 text-amber-500 dark:text-amber-400"
                  />
                  Container: {{ containerUptime }}
                </span>
                <span
                  v-if="serverUptime"
                  class="text-sm text-gray-500 dark:text-neutral-400 flex items-center gap-1.5"
                  :title="`Started at: ${formatTimestamp(serverStartedAt)}`"
                >
                  <Clock
                    class="w-3.5 h-3.5 text-emerald-500 dark:text-emerald-400"
                  />
                  Server: {{ serverUptime }}
                </span>
                <span
                  v-if="isPending"
                  class="text-xs text-gray-500 dark:text-neutral-400 italic"
                >
                  ({{ server.desiredStatus }}…)
                </span>
              </div>
            </div>

            <!-- Info tiles -->
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-6">
              <!-- Identity -->
              <div
                class="p-4 rounded-2xl bg-gray-50/50 dark:bg-neutral-700/30 border border-gray-100 dark:border-neutral-700 space-y-3"
              >
                <p
                  class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
                >
                  Identity
                </p>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <!-- Server ID -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center text-purple-600 dark:text-purple-400"
                    >
                      <Hash class="w-4 h-4" />
                    </div>
                    <div class="min-w-0">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Server ID
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white truncate font-mono"
                      >
                        {{ serverId }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Network -->
              <div
                class="p-4 rounded-2xl bg-gray-50/50 dark:bg-neutral-700/30 border border-gray-100 dark:border-neutral-700 space-y-3"
              >
                <p
                  class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
                >
                  Network
                </p>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <!-- Game Port -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400"
                    >
                      <Globe class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Game Port (host)
                      </p>
                      <div v-if="!editingPort" class="flex items-center gap-2">
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white"
                        >
                          {{ server.gamePort }}
                        </p>
                        <button
                          v-if="server.containerStatus !== 'running'"
                          class="text-gray-400 hover:text-primary transition-colors"
                          :disabled="portLoading"
                          @click="
                            tempPort = server.gamePort;
                            editingPort = true;
                          "
                        >
                          <Pencil class="w-3.5 h-3.5" />
                        </button>
                      </div>
                      <div v-else class="flex items-center gap-2">
                        <number-input
                          v-model="tempPort"
                          :min="1024"
                          :max="65535"
                          size="sm"
                          class="w-24"
                          @keyup.enter="savePort"
                        />
                        <button
                          class="text-green-500 hover:text-green-600 transition-colors"
                          :disabled="portLoading"
                          @click="savePort"
                        >
                          <Check class="w-4 h-4" />
                        </button>
                        <button
                          class="text-red-500 hover:text-red-600 transition-colors"
                          :disabled="portLoading"
                          @click="editingPort = false"
                        >
                          <XIcon class="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>

                  <!-- Public RCON -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-600 dark:text-emerald-400"
                    >
                      <Terminal class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Public RCON
                      </p>
                      <div
                        v-if="!editingPublicRcon"
                        class="flex items-center gap-2"
                      >
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white"
                        >
                          {{
                            server.publicRcon
                              ? `Yes (${server.rconPort})`
                              : "No"
                          }}
                        </p>
                        <button
                          v-if="server.containerStatus !== 'running'"
                          class="text-gray-400 hover:text-primary transition-colors"
                          :disabled="rconLoading"
                          @click="
                            tempPublicRcon = server.publicRcon;
                            tempRconPort =
                              server.rconPort || server.gamePort + 10;
                            editingPublicRcon = true;
                          "
                        >
                          <Pencil class="w-3.5 h-3.5" />
                        </button>
                      </div>
                      <div v-else class="flex items-center gap-2">
                        <select
                          v-model="tempPublicRcon"
                          class="text-sm bg-white dark:bg-neutral-700 border border-gray-200 dark:border-neutral-600 rounded-lg px-2 py-1 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:outline-none"
                        >
                          <option :value="false">No</option>
                          <option :value="true">Yes</option>
                        </select>
                        <number-input
                          v-if="tempPublicRcon"
                          v-model="tempRconPort"
                          :min="1024"
                          :max="65535"
                          size="sm"
                          class="w-24"
                          @keyup.enter="savePublicRcon"
                        />
                        <button
                          class="text-green-500 hover:text-green-600 transition-colors"
                          :disabled="rconLoading"
                          @click="savePublicRcon"
                        >
                          <Check class="w-4 h-4" />
                        </button>
                        <button
                          class="text-red-500 hover:text-red-600 transition-colors"
                          :disabled="rconLoading"
                          @click="editingPublicRcon = false"
                        >
                          <XIcon class="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Server -->
              <div
                class="p-4 rounded-2xl bg-gray-50/50 dark:bg-neutral-700/30 border border-gray-100 dark:border-neutral-700 space-y-3"
              >
                <p
                  class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
                >
                  Server
                </p>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <!-- Restart Policy -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-teal-100 dark:bg-teal-900/30 flex items-center justify-center text-teal-600 dark:text-teal-400"
                    >
                      <RefreshCw class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Restart Policy
                      </p>
                      <div
                        v-if="!editingRestartPolicy"
                        class="flex items-center gap-2"
                      >
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white"
                        >
                          {{ server.restartPolicy || "no" }}
                        </p>
                        <button
                          v-if="server.containerStatus !== 'running'"
                          class="text-gray-400 hover:text-primary transition-colors"
                          :disabled="restartPolicyLoading"
                          @click="
                            tempRestartPolicy = server.restartPolicy || 'no';
                            editingRestartPolicy = true;
                          "
                        >
                          <Pencil class="w-3.5 h-3.5" />
                        </button>
                      </div>
                      <div v-else class="flex items-center gap-2">
                        <select
                          v-model="tempRestartPolicy"
                          class="text-sm bg-white dark:bg-neutral-700 border border-gray-200 dark:border-neutral-600 rounded-lg px-2 py-1 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:outline-none"
                        >
                          <option value="no">no</option>
                          <option value="always">always</option>
                          <option value="unless-stopped">unless-stopped</option>
                          <option value="on-failure">on-failure</option>
                        </select>
                        <button
                          class="text-green-500 hover:text-green-600 transition-colors"
                          :disabled="restartPolicyLoading"
                          @click="saveRestartPolicy"
                        >
                          <Check class="w-4 h-4" />
                        </button>
                        <button
                          class="text-red-500 hover:text-red-600 transition-colors"
                          :disabled="restartPolicyLoading"
                          @click="editingRestartPolicy = false"
                        >
                          <XIcon class="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>

                  <!-- Engine -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      :class="[
                        'w-9 h-9 shrink-0 rounded-lg flex items-center justify-center text-xs font-bold',
                        getEngineIconColor(server.engineType),
                      ]"
                    >
                      {{ getEngineAbbreviation(server.engineType) }}
                    </div>
                    <div class="min-w-0">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Engine
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white"
                      >
                        {{ server.engineType }}
                      </p>
                    </div>
                  </div>

                  <!-- Version -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-pink-100 dark:bg-pink-900/30 flex items-center justify-center text-pink-600 dark:text-pink-400"
                    >
                      <Tag class="w-4 h-4" />
                    </div>
                    <div class="min-w-0">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Version
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white"
                      >
                        {{ server.gameVersion }}
                      </p>
                    </div>
                  </div>

                  <!-- Mods -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400"
                    >
                      <Package class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Mods
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white"
                      >
                        {{ server.modCount ?? 0 }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Resources -->
              <div
                class="p-4 rounded-2xl bg-gray-50/50 dark:bg-neutral-700/30 border border-gray-100 dark:border-neutral-700 space-y-3"
              >
                <p
                  class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
                >
                  Resources
                </p>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <!-- RAM -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400"
                    >
                      <MemoryStick class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        RAM
                      </p>
                      <div v-if="!editingRam" class="flex items-center gap-2">
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white"
                        >
                          {{
                            server.ramBytes
                              ? (
                                  server.ramBytes /
                                  (1024 * 1024 * 1024)
                                ).toFixed(0) + " GB"
                              : "-"
                          }}
                        </p>
                        <button
                          v-if="server.containerStatus !== 'running'"
                          class="text-gray-400 hover:text-primary transition-colors"
                          :disabled="ramLoading"
                          @click="
                            tempRamGb = server.ramBytes
                              ? Math.round(
                                  server.ramBytes / (1024 * 1024 * 1024),
                                )
                              : 2;
                            editingRam = true;
                          "
                        >
                          <Pencil class="w-3.5 h-3.5" />
                        </button>
                      </div>
                      <div v-else class="flex items-center gap-2">
                        <number-input
                          v-model="tempRamGb"
                          :min="1"
                          :max="128"
                          size="sm"
                          class="w-24"
                          @keyup.enter="saveRam"
                        />
                        <span
                          class="text-sm text-gray-500 dark:text-neutral-400"
                          >GB</span
                        >
                        <button
                          class="text-green-500 hover:text-green-600 transition-colors"
                          :disabled="ramLoading"
                          @click="saveRam"
                        >
                          <Check class="w-4 h-4" />
                        </button>
                        <button
                          class="text-red-500 hover:text-red-600 transition-colors"
                          :disabled="ramLoading"
                          @click="editingRam = false"
                        >
                          <XIcon class="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>

                  <!-- CPUs -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center text-orange-600 dark:text-orange-400"
                    >
                      <Cpu class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        CPUs
                      </p>
                      <div v-if="!editingCpu" class="flex items-center gap-2">
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white"
                        >
                          {{ server.cpus ?? "-" }}
                        </p>
                        <button
                          v-if="server.containerStatus !== 'running'"
                          class="text-gray-400 hover:text-primary transition-colors"
                          :disabled="cpuLoading"
                          @click="
                            tempCpu = server.cpus ?? 1;
                            editingCpu = true;
                          "
                        >
                          <Pencil class="w-3.5 h-3.5" />
                        </button>
                      </div>
                      <div v-else class="flex items-center gap-2">
                        <number-input
                          v-model="tempCpu"
                          :min="0.5"
                          :step="0.5"
                          size="sm"
                          class="w-24"
                          @keyup.enter="saveCpu"
                        />
                        <button
                          class="text-green-500 hover:text-green-600 transition-colors"
                          :disabled="cpuLoading"
                          @click="saveCpu"
                        >
                          <Check class="w-4 h-4" />
                        </button>
                        <button
                          class="text-red-500 hover:text-red-600 transition-colors"
                          :disabled="cpuLoading"
                          @click="editingCpu = false"
                        >
                          <XIcon class="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Storage -->
              <div
                class="p-4 rounded-2xl bg-gray-50/50 dark:bg-neutral-700/30 border border-gray-100 dark:border-neutral-700 space-y-3"
              >
                <p
                  class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
                >
                  Storage
                </p>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <!-- Path (host) -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 col-span-full"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-gray-100 dark:bg-neutral-600/50 flex items-center justify-center text-gray-600 dark:text-gray-400"
                    >
                      <FolderOpen class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Path (host)
                      </p>
                      <div class="flex items-center gap-2">
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white break-all font-mono"
                          :title="server.hostPath"
                        >
                          {{ server.hostPath || "-" }}
                        </p>
                        <button
                          v-if="server.hostPath"
                          class="text-gray-400 hover:text-primary transition-colors shrink-0"
                          :title="'Copy path'"
                          @click="copyToClipboard(server.hostPath!)"
                        >
                          <Copy class="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </div>
                  </div>
                  <!-- Total Volume -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-gray-100 dark:bg-neutral-600/50 flex items-center justify-center text-gray-600 dark:text-gray-400"
                    >
                      <FolderOpen class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Total Volume
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white"
                      >
                        {{
                          serverDisk ? formatBytes(serverDisk.totalBytes) : "—"
                        }}
                      </p>
                    </div>
                  </div>
                  <!-- World -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-600 dark:text-emerald-400"
                    >
                      <Globe class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        World
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white"
                      >
                        {{
                          serverDisk ? formatBytes(serverDisk.worldBytes) : "—"
                        }}
                      </p>
                    </div>
                  </div>
                  <!-- Nether -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-red-100 dark:bg-red-900/30 flex items-center justify-center text-red-600 dark:text-red-400"
                    >
                      <Zap class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Nether
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white"
                      >
                        {{
                          serverDisk
                            ? formatBytes(serverDisk.worldNetherBytes)
                            : "—"
                        }}
                      </p>
                    </div>
                  </div>
                  <!-- The End -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center text-purple-600 dark:text-purple-400"
                    >
                      <Zap class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        The End
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white"
                      >
                        {{
                          serverDisk
                            ? formatBytes(serverDisk.worldEndBytes)
                            : "—"
                        }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex flex-wrap items-center gap-3">
              <div class="flex flex-wrap gap-3">
                <button
                  :disabled="
                    actionLoading ||
                    server?.containerStatus === 'running' ||
                    isPending
                  "
                  class="inline-flex items-center gap-2 bg-emerald-500 hover:bg-emerald-600 text-white px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed shadow-sm hover:shadow-md active:scale-95"
                  @click="doAction('start')"
                >
                  <Loader2
                    v-if="currentAction === 'start'"
                    class="w-4 h-4 animate-spin"
                  />
                  <Play v-else class="w-4 h-4" />
                  Start
                </button>
                <button
                  :disabled="
                    actionLoading ||
                    server?.containerStatus !== 'running' ||
                    isPending
                  "
                  class="inline-flex items-center gap-2 bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed shadow-sm hover:shadow-md active:scale-95"
                  @click="doAction('stop')"
                >
                  <Loader2
                    v-if="currentAction === 'stop'"
                    class="w-4 h-4 animate-spin"
                  />
                  <Square v-else class="w-4 h-4" />
                  Stop
                </button>
                <button
                  :disabled="
                    actionLoading ||
                    server?.containerStatus !== 'running' ||
                    isPending
                  "
                  class="inline-flex items-center gap-2 bg-orange-600 hover:bg-orange-700 text-white px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed shadow-sm hover:shadow-md active:scale-95"
                  @click="doAction('force-stop')"
                >
                  <Loader2
                    v-if="currentAction === 'force-stop'"
                    class="w-4 h-4 animate-spin"
                  />
                  <OctagonAlert v-else class="w-4 h-4" />
                  Force Stop
                </button>
                <button
                  :disabled="
                    actionLoading ||
                    server?.containerStatus !== 'running' ||
                    isPending
                  "
                  class="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed shadow-sm hover:shadow-md active:scale-95"
                  @click="doAction('restart')"
                >
                  <Loader2
                    v-if="currentAction === 'restart'"
                    class="w-4 h-4 animate-spin"
                  />
                  <RotateCcw v-else class="w-4 h-4" />
                  Restart
                </button>
              </div>
              <div class="flex flex-wrap gap-3 ml-auto">
                <button
                  v-if="
                    server.containerStatus === 'running' ||
                    server.containerStatus === 'exited'
                  "
                  :disabled="recreateLoading || isPending"
                  class="inline-flex items-center gap-2 bg-amber-500 hover:bg-amber-600 text-white px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed shadow-sm hover:shadow-md active:scale-95"
                  @click="promptRecreate"
                >
                  <RefreshCw class="w-4 h-4" />
                  Recreate World
                </button>
                <button
                  :disabled="deleteLoading || isPending"
                  class="inline-flex items-center gap-2 bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed shadow-sm hover:shadow-md active:scale-95"
                  @click="promptDelete"
                >
                  <Trash2 class="w-4 h-4" />
                  Delete
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Crash Reports Alert -->
      <div
        v-if="hasCrashReports"
        class="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-2xl p-4 flex items-start gap-3"
      >
        <AlertTriangle
          class="w-5 h-5 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5"
        />
        <div class="flex-1">
          <p class="text-sm font-semibold text-amber-800 dark:text-amber-300">
            {{ crashReports.length }} crash report{{
              crashReports.length === 1 ? "" : "s"
            }}
            found
          </p>
          <p class="text-xs text-amber-700 dark:text-amber-400 mt-0.5">
            Latest: {{ formatTimestamp(latestCrashReportDate || undefined) }}
          </p>
        </div>
        <NuxtLink
          :to="`/servers/${serverId}/crash-reports`"
          class="text-xs font-medium text-amber-700 dark:text-amber-400 hover:underline flex-shrink-0"
        >
          View →
        </NuxtLink>
      </div>

      <!-- Quick Console -->
      <div
        v-if="server?.serverStatus === 'running'"
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden"
      >
        <button
          class="w-full flex items-center justify-between p-4 md:p-6 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
          @click="quickConsoleExpanded = !quickConsoleExpanded"
        >
          <div class="flex items-center gap-3">
            <Terminal class="w-5 h-5 text-gray-500 dark:text-neutral-400" />
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">
              Quick Console
            </h2>
          </div>
          <ChevronDown
            class="w-5 h-5 text-gray-500 dark:text-neutral-400 transition-transform"
            :class="{ 'rotate-180': quickConsoleExpanded }"
          />
        </button>
        <div v-show="quickConsoleExpanded" class="px-4 pb-4 md:px-6 md:pb-6">
          <div class="flex items-center gap-2">
            <input
              v-model="quickConsoleCommand"
              type="text"
              placeholder="Enter command (e.g. /say Hello)"
              class="flex-1 bg-white dark:bg-neutral-700 border border-gray-200 dark:border-neutral-600 rounded-xl px-3 py-2 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:outline-none"
              @keyup.enter="sendQuickConsole"
            />
            <button
              :disabled="quickConsoleLoading || !quickConsoleCommand.trim()"
              class="inline-flex items-center gap-1.5 bg-primary hover:bg-primary/90 text-white px-4 py-2 rounded-xl text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              @click="sendQuickConsole"
            >
              <Loader2
                v-if="quickConsoleLoading"
                class="w-4 h-4 animate-spin"
              />
              <Send v-else class="w-4 h-4" />
              Send
            </button>
          </div>
          <pre
            v-if="quickConsoleOutput"
            class="mt-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 text-xs text-gray-700 dark:text-neutral-300 font-mono whitespace-pre-wrap max-h-40 overflow-y-auto"
            >{{ quickConsoleOutput }}</pre
          >
        </div>
      </div>

      <!-- Real-time Metrics -->
      <div
        v-if="serverMetrics.length > 0"
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden"
      >
        <button
          class="w-full flex items-center justify-between p-4 md:p-6 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
          @click="liveMetricsExpanded = !liveMetricsExpanded"
        >
          <div class="flex items-center gap-3">
            <Activity class="w-5 h-5 text-gray-500 dark:text-neutral-400" />
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">
              Live Metrics
            </h2>
          </div>
          <ChevronDown
            class="w-5 h-5 text-gray-500 dark:text-neutral-400 transition-transform"
            :class="{ 'rotate-180': liveMetricsExpanded }"
          />
        </button>
        <div v-show="liveMetricsExpanded" class="px-4 pb-4 md:px-6 md:pb-6">
          <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
            <!-- RAM -->
            <div class="p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50">
              <div class="text-xs text-gray-500 dark:text-neutral-400 mb-1">
                RAM
              </div>
              <div class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ latestMetric ? formatBytes(latestMetric.ramUsage) : "—" }}
              </div>
              <div class="text-xs text-gray-400 dark:text-neutral-500">
                / {{ latestMetric ? formatBytes(latestMetric.ramLimit) : "—" }}
              </div>
              <svg class="w-full h-8 mt-2" preserveAspectRatio="none">
                <polyline
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  class="text-indigo-500"
                  :points="sparklinePoints(serverMetrics, 'ramUsage')"
                />
              </svg>
            </div>
            <!-- CPU -->
            <div class="p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50">
              <div class="text-xs text-gray-500 dark:text-neutral-400 mb-1">
                CPU
              </div>
              <div class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ latestMetric ? latestMetric.cpu.toFixed(1) + "%" : "—" }}
              </div>
              <svg class="w-full h-8 mt-2" preserveAspectRatio="none">
                <polyline
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  class="text-emerald-500"
                  :points="sparklinePoints(serverMetrics, 'cpu')"
                />
              </svg>
            </div>
            <!-- Online -->
            <div class="p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50">
              <div class="text-xs text-gray-500 dark:text-neutral-400 mb-1">
                Online
              </div>
              <div class="text-lg font-semibold text-gray-900 dark:text-white">
                {{
                  latestMetric
                    ? latestMetric.online + "/" + latestMetric.max
                    : "—"
                }}
              </div>
              <svg class="w-full h-8 mt-2" preserveAspectRatio="none">
                <polyline
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  class="text-sky-500"
                  :points="sparklinePoints(serverMetrics, 'online')"
                />
              </svg>
            </div>
            <!-- TPS -->
            <div class="p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50">
              <div class="text-xs text-gray-500 dark:text-neutral-400 mb-1">
                TPS
              </div>
              <div
                class="text-lg font-semibold"
                :class="
                  latestMetric?.tps == null
                    ? 'text-gray-900 dark:text-white'
                    : latestMetric.tps >= 18
                      ? 'text-green-600 dark:text-green-400'
                      : latestMetric.tps >= 15
                        ? 'text-yellow-600 dark:text-yellow-400'
                        : 'text-red-600 dark:text-red-400'
                "
              >
                {{
                  latestMetric?.tps != null ? latestMetric.tps.toFixed(1) : "—"
                }}
              </div>
              <svg class="w-full h-8 mt-2" preserveAspectRatio="none">
                <polyline
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  class="text-purple-500"
                  :points="sparklinePoints(serverMetrics, 'tps')"
                />
              </svg>
            </div>
            <!-- Network RX -->
            <div class="p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50">
              <div
                class="text-xs text-gray-500 dark:text-neutral-400 mb-1 flex items-center gap-1"
              >
                <ArrowDown class="w-3 h-3" />
                Net RX
              </div>
              <div class="text-lg font-semibold text-gray-900 dark:text-white">
                {{
                  latestMetric
                    ? (latestMetric.rxRate / 1024).toFixed(2) + " MB/s"
                    : "—"
                }}
              </div>
              <div class="text-xs text-gray-400 dark:text-neutral-500">
                {{
                  latestMetric ? latestMetric.rxRate.toFixed(1) + " KB/s" : ""
                }}
              </div>
              <svg class="w-full h-8 mt-2" preserveAspectRatio="none">
                <polyline
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  class="text-blue-500"
                  :points="sparklinePoints(serverMetrics, 'rxRate')"
                />
              </svg>
            </div>
            <!-- Network TX -->
            <div class="p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50">
              <div
                class="text-xs text-gray-500 dark:text-neutral-400 mb-1 flex items-center gap-1"
              >
                <ArrowUp class="w-3 h-3" />
                Net TX
              </div>
              <div class="text-lg font-semibold text-gray-900 dark:text-white">
                {{
                  latestMetric
                    ? (latestMetric.txRate / 1024).toFixed(2) + " MB/s"
                    : "—"
                }}
              </div>
              <div class="text-xs text-gray-400 dark:text-neutral-500">
                {{
                  latestMetric ? latestMetric.txRate.toFixed(1) + " KB/s" : ""
                }}
              </div>
              <svg class="w-full h-8 mt-2" preserveAspectRatio="none">
                <polyline
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  class="text-teal-500"
                  :points="sparklinePoints(serverMetrics, 'txRate')"
                />
              </svg>
            </div>
          </div>
        </div>
      </div>

      <!-- Players Online -->
      <div
        v-if="
          server?.serverStatus === 'running' && onlinePlayers.players.length > 0
        "
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl p-4 md:p-6 shadow-sm"
      >
        <div class="flex items-center justify-between mb-4">
          <h2
            class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2"
          >
            <Users class="w-5 h-5 text-primary" />
            Players Online
          </h2>
          <span class="text-sm text-gray-500 dark:text-neutral-400">
            {{ onlinePlayers.players.length }} / {{ onlinePlayers.max }}
          </span>
        </div>
        <div class="flex flex-wrap gap-3">
          <div
            v-for="name in onlinePlayers.players"
            :key="name"
            class="flex items-center gap-2 px-3 py-2 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
          >
            <img
              :src="`https://mc-heads.net/avatar/${name}/32`"
              alt=""
              class="w-6 h-6 rounded"
              loading="lazy"
            />
            <span class="text-sm font-medium text-gray-900 dark:text-white">{{
              name
            }}</span>
          </div>
        </div>
      </div>

      <!-- Logs -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden"
      >
        <button
          class="w-full flex items-center justify-between p-4 md:p-6 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
          @click="toggleLogs"
        >
          <div class="flex items-center gap-3">
            <Terminal class="w-5 h-5 text-gray-500 dark:text-neutral-400" />
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">
              Logs
            </h2>
          </div>
          <ChevronDown
            class="w-5 h-5 text-gray-500 dark:text-neutral-400 transition-transform"
            :class="{ 'rotate-180': logsExpanded }"
          />
        </button>
        <div
          v-show="logsExpanded"
          class="px-4 pt-4 pb-4 md:px-6 md:pt-6 md:pb-6 h-96"
        >
          <server-logs
            ref="serverLogsRef"
            :server-id="serverId"
            class="h-full"
          />
        </div>
      </div>

      <!-- Server Properties -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl shadow-sm overflow-hidden"
      >
        <button
          class="w-full flex items-center justify-between p-4 md:p-6 hover:bg-gray-50 dark:hover:bg-neutral-700/50 transition-colors"
          @click="propertiesExpanded = !propertiesExpanded"
        >
          <div class="flex items-center gap-3">
            <Pencil class="w-5 h-5 text-gray-500 dark:text-neutral-400" />
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">
              Server Properties
            </h2>
          </div>
          <ChevronDown
            class="w-5 h-5 text-gray-500 dark:text-neutral-400 transition-transform"
            :class="{ 'rotate-180': propertiesExpanded }"
          />
        </button>
        <div
          v-show="propertiesExpanded"
          class="px-4 pt-4 pb-4 md:px-6 md:pt-6 md:pb-6"
        >
          <server-properties :server-id="serverId" />
        </div>
      </div>
    </div>

    <div v-else class="text-gray-500 dark:text-neutral-400">Loading...</div>

    <confirm-dialog
      v-if="server"
      v-model="showDeleteDialog"
      title="Delete Server"
      :description="`Delete server &quot;${serverId}&quot;? The container will be removed.`"
      confirm-label="Delete"
      danger
      show-wipe
      @confirm="onDeleteConfirmed"
    />
    <confirm-dialog
      v-if="server"
      v-model="showRecreateDialog"
      title="Recreate World"
      :description="`This will permanently delete the current world for &quot;${serverId}&quot; and restart the server with a fresh world. This cannot be undone.`"
      confirm-label="Recreate"
      danger
      @confirm="onRecreateConfirmed"
    />
    <confirm-dialog
      v-if="server"
      v-model="showPortConflictDialog"
      title="Port Conflict"
      :description="`Game port ${server.gamePort} is already used by another server. Reassign to a free port?`"
      confirm-label="Reassign"
      simple
      @confirm="onReassignPortsConfirmed"
    />
  </div>
</template>

<script setup lang="ts">
import { formatDuration } from "~/composables/useDuration";
import {
  Activity,
  AlertTriangle,
  ArrowDown,
  ArrowUp,
  Check,
  ChevronDown,
  Clock,
  Copy,
  Cpu,
  FolderOpen,
  Globe,
  Hash,
  MemoryStick,
  OctagonAlert,
  Package,
  Pencil,
  Play,
  RefreshCw,
  RotateCcw,
  Send,
  Server as ServerIcon,
  Square,
  Tag,
  Terminal,
  Trash2,
  Users,
  X as XIcon,
  Zap,
  Loader2,
} from "lucide-vue-next";
import { nextTick, onMounted, onBeforeUnmount, watch } from "vue";
import type { Server } from "~/composables/useServers";
import {
  useServerEvents,
  type MetricsPayload,
} from "~/composables/useServerEvents";

definePageMeta({
  middleware: "auth",
});

const route = useRoute();
const { show: showToast } = useToast();
const serverId = route.params.serverId as string;
const { servers, refresh: refreshServers } = useServers();
const { metricsMap, playersMap } = useServerEvents();
const { disk: serverDisk } = useServerDisk(serverId);
const server = computed<Server | null>(
  () => servers.value.find((s: Server) => s.serverId === serverId) ?? null,
);
usePageTitle(() => server.value?.serverId || serverId);
const actionLoading = ref(false);
const currentAction = ref<"start" | "stop" | "force-stop" | "restart" | null>(
  null,
);
const iconTimestamp = ref(Date.now());
const iconError = ref(false);
const fileInput = ref<HTMLInputElement | null>(null);
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
const deleteLoading = ref(false);
const recreateLoading = ref(false);
const showDeleteDialog = ref(false);
const showRecreateDialog = ref(false);
const showIconEditor = ref(false);
const selectedIconFile = ref<File | null>(null);
const logsExpanded = ref(false);
const propertiesExpanded = ref(false);
const quickConsoleExpanded = ref(true);
const liveMetricsExpanded = ref(true);
const serverLogsRef = ref<{
  scrollToBottom: () => void;
  reconnect: (tail?: number) => void;
} | null>(null);
const editingRam = ref(false);
const tempRamGb = ref<number | null>(null);
const ramLoading = ref(false);
const editingCpu = ref(false);
const tempCpu = ref<number | null>(null);
const cpuLoading = ref(false);

const quickConsoleCommand = ref("");
const quickConsoleOutput = ref("");
const quickConsoleLoading = ref(false);

const crashReports = ref<{ name: string; size: number; modifiedAt: string }[]>(
  [],
);
const crashReportsLoading = ref(false);
const crashReportsFetched = ref(false);

const showPortConflictDialog = ref(false);
const portConflictChecked = ref(false);

const hasPortConflict = computed(() => {
  if (!server.value) return false;
  if (server.value.containerStatus === "running") return false;
  return servers.value.some(
    (s) => s.serverId !== serverId && s.gamePort === server.value!.gamePort,
  );
});

const onlinePlayers = computed(() => {
  return playersMap.value[serverId] ?? { players: [], max: 0 };
});

const hasCrashReports = computed(() => crashReports.value.length > 0);

const latestCrashReportDate = computed(() => {
  if (!crashReports.value.length) return null;
  return crashReports.value[0]?.modifiedAt ?? null;
});

async function fetchCrashReports() {
  if (crashReportsFetched.value) return;
  crashReportsLoading.value = true;
  try {
    const res = await $fetch<
      | { reports?: { name: string; size: number; modifiedAt: string }[] }
      | {
          body?: {
            reports?: { name: string; size: number; modifiedAt: string }[];
          };
        }
    >(`/servers/${serverId}/crash-reports`, {
      baseURL: useApiBase(),
      credentials: "include",
    });
    const body = (res as any).body ?? (res as any);
    crashReports.value = body?.reports ?? [];
    crashReportsFetched.value = true;
  } catch {
    crashReports.value = [];
  } finally {
    crashReportsLoading.value = false;
  }
}

async function sendQuickConsole() {
  const cmd = quickConsoleCommand.value.trim();
  if (!cmd) return;
  quickConsoleLoading.value = true;
  try {
    const res = await $fetch<
      { response?: string } | { body?: { response?: string } }
    >(`/servers/${serverId}/rcon`, {
      baseURL: useApiBase(),
      method: "POST",
      credentials: "include",
      body: { command: cmd },
    });
    const body = (res as any).body ?? (res as any);
    quickConsoleOutput.value = body?.response ?? "No response";
    quickConsoleCommand.value = "";
  } catch (err: any) {
    const msg = err?.data?.detail || err?.message || "Command failed";
    quickConsoleOutput.value = `Error: ${msg}`;
  } finally {
    quickConsoleLoading.value = false;
  }
}

onMounted(() => {
  fetchCrashReports();
});

watch(
  () => server.value?.gamePort,
  () => {
    if (hasPortConflict.value && !portConflictChecked.value) {
      showPortConflictDialog.value = true;
      portConflictChecked.value = true;
    }
  },
  { immediate: true },
);

async function onReassignPortsConfirmed() {
  portLoading.value = true;
  try {
    await $fetch(`/servers/${serverId}/reassign-ports`, {
      baseURL: useApiBase(),
      method: "POST",
      credentials: "include",
    });
    showToast("success", "Ports reassigned", {
      description: "Server ports have been updated to free ones.",
    });
    await refreshServers();
  } catch (err: any) {
    const msg = err?.data?.detail || err?.message || "Failed to reassign ports";
    showToast("error", "Reassign failed", { description: msg });
  } finally {
    portLoading.value = false;
  }
}

const serverMetrics = computed<MetricsPayload[]>(
  () => metricsMap.value[serverId] || [],
);
const latestMetric = computed<MetricsPayload | null>(() => {
  const arr = serverMetrics.value;
  return arr.length > 0 ? (arr[arr.length - 1] ?? null) : null;
});

const iconUrl = computed(() => {
  if (!server.value) return "";
  return `${useApiBase()}/servers/${serverId}/icon?t=${iconTimestamp.value}`;
});

const containerStartedAt = computed(() => {
  if (
    !server.value ||
    server.value.containerStatus !== "running" ||
    !server.value.containerStartedAt
  ) {
    return undefined;
  }
  return server.value.containerStartedAt;
});

const serverStartedAt = computed(() => {
  if (
    !server.value ||
    server.value.serverStatus !== "running" ||
    !server.value.serverStartedAt
  ) {
    return undefined;
  }
  return server.value.serverStartedAt;
});

function toggleLogs() {
  logsExpanded.value = !logsExpanded.value;
  if (logsExpanded.value && serverLogsRef.value) {
    nextTick(() => serverLogsRef.value!.scrollToBottom());
  }
}

function formatTimestamp(iso: string | undefined): string {
  if (!iso) return "";
  return new Date(iso).toLocaleString();
}

const containerUptime = useUptime(containerStartedAt);
const serverUptime = useUptime(serverStartedAt);

function onIconFileSelect(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0] ?? null;
  if (!file) return;
  selectedIconFile.value = file;
  showIconEditor.value = true;
  input.value = "";
}

async function uploadProcessedIcon(blob: Blob) {
  const formData = new FormData();
  formData.append("icon", blob, "icon.png");

  try {
    await $fetch(`/servers/${serverId}/icon`, {
      baseURL: useApiBase(),
      method: "POST",
      body: formData,
      credentials: "include",
    });
    showToast("info", "Icon updated");
    iconTimestamp.value = Date.now();
    iconError.value = false;
  } catch (err: any) {
    const msg = err?.data?.detail || err?.message || "Failed to upload icon";
    showToast("error", "Upload failed", { description: msg });
  }
}

const isPending = computed(() => {
  if (!server.value) return false;
  const d = server.value.desiredStatus;
  return !!d && d !== server.value.containerStatus;
});

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
}

function formatNetworkRate(kbps: number): string {
  if (kbps >= 1024) {
    return (kbps / 1024).toFixed(1) + " MB/s";
  }
  return kbps.toFixed(1) + " KB/s";
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text);
    showToast("info", "Path copied to clipboard");
  } catch {
    showToast("error", "Failed to copy path");
  }
}

function sparklinePoints(
  data: MetricsPayload[],
  key: "ramUsage" | "cpu" | "online" | "tps" | "rxRate" | "txRate",
): string {
  let values: number[];
  if (key === "online") {
    values = data.map((d) => (d.max > 0 ? (d.online / d.max) * 100 : 0));
  } else if (key === "tps") {
    values = data.map((d) => d.tps ?? 0);
  } else {
    values = data.map((d) => (d as any)[key] ?? 0);
  }
  if (values.length < 2) return "";
  const max = Math.max(...values, 0.001);
  const min = Math.min(...values);
  const range = max - min || 1;
  const w = 100;
  const h = 32;
  const step = w / (values.length - 1);
  return values
    .map((v, i) => {
      const x = i * step;
      const y = h - ((v - min) / range) * h;
      return `${x},${y}`;
    })
    .join(" ");
}

function getStatusColor(status: string) {
  switch (status) {
    case "running":
      return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
    case "starting":
      return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
    case "exited":
      return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
    default:
      return "bg-gray-100 text-gray-800 dark:bg-neutral-700 dark:text-neutral-300";
  }
}

function getEngineAbbreviation(engineType: string) {
  switch (engineType.toUpperCase()) {
    case "FORGE":
      return "FO";
    case "FABRIC":
      return "FA";
    case "PAPERMC":
      return "PA";
    case "VANILLA":
      return "MC";
    default:
      return engineType.substring(0, 2).toUpperCase();
  }
}

function getEngineIconColor(engineType: string) {
  switch (engineType.toUpperCase()) {
    case "FORGE":
      return "bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400";
    case "FABRIC":
      return "bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400";
    case "PAPERMC":
      return "bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400";
    case "VANILLA":
      return "bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400";
    default:
      return "bg-gray-100 text-gray-600 dark:bg-neutral-700 dark:text-neutral-400";
  }
}

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
    tempRconPort.value === (server.value.rconPort || server.value.gamePort + 10)
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

async function saveCpu() {
  if (!server.value || tempCpu.value == null) {
    editingCpu.value = false;
    return;
  }
  if (tempCpu.value === server.value.cpus) {
    editingCpu.value = false;
    return;
  }
  if (tempCpu.value < 0.5 || tempCpu.value > 128) {
    showToast("error", "Invalid CPUs", {
      description: "CPUs must be between 0.5 and 128.",
    });
    return;
  }
  cpuLoading.value = true;
  try {
    await $fetch(`/servers/${serverId}`, {
      baseURL: useApiBase(),
      method: "PATCH",
      credentials: "include",
      body: { cpus: tempCpu.value },
    });
    showToast("success", "CPUs updated", {
      description: `CPUs changed to ${tempCpu.value}.`,
    });
    editingCpu.value = false;
    await refreshServers();
  } catch (err: any) {
    const status = err?.status || err?.statusCode;
    const msg = err?.data?.detail || err?.message || "Failed to update CPUs";
    if (status === 409) {
      showToast("error", "Server running", { description: msg });
    } else {
      showToast("error", "Update failed", { description: msg });
    }
  } finally {
    cpuLoading.value = false;
  }
}

async function doAction(action: "start" | "stop" | "restart" | "force-stop") {
  if (actionLoading.value) return;
  actionLoading.value = true;
  currentAction.value = action;
  if (action === "start" || action === "restart") {
    serverLogsRef.value?.reconnect(0);
  }
  try {
    await $fetch(`/servers/${serverId}/${action}`, {
      baseURL: useApiBase(),
      method: "POST",
      credentials: "include",
    });
    showToast("info", `Server ${action} requested`, {
      description: `${serverId} — operation in progress.`,
    });
    await refreshServers();
  } catch (err: any) {
    const status = err?.status || err?.statusCode;
    const msg =
      err?.data?.detail || err?.message || `Failed to ${action} server`;
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
    await navigateTo(`/servers`);
  } catch (err: any) {
    const msg = err?.data?.detail || err?.message || "Failed to delete server";
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
    const msg = err?.data?.detail || err?.message || "Failed to recreate world";
    showToast("error", "Recreate failed", { description: msg });
  } finally {
    recreateLoading.value = false;
  }
}
</script>
