<template>
  <div class="p-6">
    <div v-if="server" class="space-y-6">
      <!-- Main server card -->
      <div
        class="bg-white dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 rounded-2xl p-6 md:p-8 shadow-sm"
      >
        <div class="flex flex-col md:flex-row gap-8">
          <!-- Avatar column -->
          <div class="flex flex-col items-center gap-4 shrink-0 max-w-[200px]">
            <div class="relative w-full">
              <div
                class="w-full aspect-square rounded-2xl overflow-hidden shadow-lg bg-gradient-to-br from-gray-200 to-gray-300 dark:from-neutral-700 dark:to-neutral-600 flex items-center justify-center ring-4 ring-white dark:ring-neutral-800"
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
              <button
                v-if="server.hostPath"
                class="absolute inset-0 flex items-center justify-center bg-black/40 opacity-0 hover:opacity-100 transition-opacity rounded-2xl cursor-pointer"
                @click="fileInput?.click()"
              >
                <Camera class="w-6 h-6 text-white" />
              </button>
              <span
                :class="[
                  'absolute -bottom-1 -right-1 w-5 h-5 rounded-full border-2 border-white dark:border-neutral-800',
                  server.serverStatus === 'running'
                    ? 'bg-green-500'
                    : server.serverStatus === 'starting'
                      ? 'bg-yellow-500'
                      : server.serverStatus === 'pulling_image'
                        ? 'bg-blue-500'
                        : server.containerStatus === 'exited'
                          ? 'bg-red-500'
                          : 'bg-gray-400',
                ]"
              />
            </div>
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

            <!-- Actions -->
            <div class="flex flex-col items-stretch gap-2 w-full">
              <!-- Main controls -->
              <div class="flex flex-col gap-2">
                <div class="flex items-center gap-2">
                  <button
                    :disabled="
                      actionLoading ||
                      server?.containerStatus === 'running' ||
                      isPending
                    "
                    class="flex items-center gap-2 flex-1 justify-center bg-gray-100 dark:bg-neutral-700/50 hover:bg-gray-200 dark:hover:bg-neutral-700 text-gray-700 dark:text-gray-200 border border-gray-200 dark:border-neutral-600 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed active:scale-95"
                    @click="doAction('start')"
                  >
                    <Loader2
                      v-if="currentAction === 'start'"
                      class="w-4 h-4 animate-spin"
                    />
                    <Play v-else class="w-4 h-4" />
                    Start
                  </button>
                  <label
                    class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-neutral-400 cursor-pointer select-none shrink-0"
                  >
                    <input
                      v-model="removeBeforeStart"
                      type="checkbox"
                      class="w-3.5 h-3.5 rounded border-gray-300 dark:border-neutral-600 text-indigo-600 dark:text-indigo-400 focus:ring-indigo-500 dark:focus:ring-indigo-400 bg-white dark:bg-neutral-700 cursor-pointer"
                    />
                    Recreate
                  </label>
                </div>
                <button
                  :disabled="
                    actionLoading ||
                    server?.containerStatus !== 'running' ||
                    isPending
                  "
                  class="flex items-center gap-2 w-full justify-center bg-gray-100 dark:bg-neutral-700/50 hover:bg-gray-200 dark:hover:bg-neutral-700 text-gray-700 dark:text-gray-200 border border-gray-200 dark:border-neutral-600 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed active:scale-95"
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
                  class="flex items-center gap-2 w-full justify-center bg-gray-100 dark:bg-neutral-700/50 hover:bg-gray-200 dark:hover:bg-neutral-700 text-gray-700 dark:text-gray-200 border border-gray-200 dark:border-neutral-600 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed active:scale-95"
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
                  class="flex items-center gap-2 w-full justify-center bg-gray-100 dark:bg-neutral-700/50 hover:bg-gray-200 dark:hover:bg-neutral-700 text-gray-700 dark:text-gray-200 border border-gray-200 dark:border-neutral-600 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-60 disabled:cursor-not-allowed active:scale-95"
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

              <!-- Danger zone -->
              <div
                class="mt-3 pt-3 border-t border-gray-200 dark:border-neutral-700 flex flex-col gap-2"
              >
                <button
                  v-if="
                    server.containerStatus === 'running' ||
                    server.containerStatus === 'exited'
                  "
                  :disabled="recreateLoading || isPending"
                  class="flex items-center gap-2 w-full justify-center bg-amber-500/10 hover:bg-amber-500/20 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-800/50 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed active:scale-95"
                  @click="promptRecreate"
                >
                  <RefreshCw class="w-4 h-4" />
                  Recreate World
                </button>
                <button
                  :disabled="deleteLoading || isPending"
                  class="flex items-center gap-2 w-full justify-center bg-red-500/10 hover:bg-red-500/20 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800/50 px-4 py-2 rounded-xl transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed active:scale-95"
                  @click="promptDelete"
                >
                  <Trash2 class="w-4 h-4" />
                  Delete Server
                </button>
              </div>
            </div>
          </div>

          <!-- Details column -->
          <div class="flex-1 min-w-0">
            <!-- Title row -->
            <div class="mb-2">
              <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
                {{ server.serverId }}
              </h1>
            </div>

            <!-- Info tiles -->
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-6">
              <!-- Server Status -->
              <div class="space-y-2">
                <p
                  class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
                >
                  Server Status
                </p>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <!-- Container -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      :class="[
                        'w-9 h-9 shrink-0 rounded-lg flex items-center justify-center',
                        server.containerStatus === 'running'
                          ? 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400'
                          : server.containerStatus === 'starting'
                            ? 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 dark:text-yellow-400'
                            : server.containerStatus === 'exited'
                              ? 'bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400'
                              : 'bg-gray-100 dark:bg-neutral-600/50 text-gray-600 dark:text-gray-400',
                      ]"
                    >
                      <Activity
                        v-if="server.containerStatus === 'running'"
                        class="w-4 h-4 animate-pulse-icon"
                      />
                      <Activity v-else class="w-4 h-4" />
                    </div>
                    <div class="min-w-0">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Container
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white"
                      >
                        {{ server.containerStatus }}
                        <span
                          v-if="containerUptime"
                          class="text-gray-500 dark:text-neutral-400 font-normal"
                        >
                          · {{ containerUptime }}
                        </span>
                      </p>
                    </div>
                  </div>
                  <!-- Server -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      :class="[
                        'w-9 h-9 shrink-0 rounded-lg flex items-center justify-center',
                        server.serverStatus === 'running'
                          ? 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400'
                          : server.serverStatus === 'starting'
                            ? 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 dark:text-yellow-400'
                            : server.serverStatus === 'pulling_image'
                              ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400'
                              : 'bg-gray-100 dark:bg-neutral-600/50 text-gray-600 dark:text-gray-400',
                      ]"
                    >
                      <Activity
                        v-if="server.serverStatus === 'running'"
                        class="w-4 h-4 animate-pulse-icon"
                      />
                      <CloudDownload
                        v-else-if="server.serverStatus === 'pulling_image'"
                        class="w-4 h-4 animate-bounce"
                      />
                      <Activity v-else class="w-4 h-4" />
                    </div>
                    <div class="min-w-0">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Server
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white"
                      >
                        {{ server.serverStatus }}
                        <span
                          v-if="serverUptime"
                          class="text-gray-500 dark:text-neutral-400 font-normal"
                        >
                          · {{ serverUptime }}
                        </span>
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Network -->
              <div class="space-y-2">
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
              <div class="space-y-2">
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

                  <!-- Loader -->
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
                        Loader
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white"
                      >
                        {{ server.loaderVersion || "—" }}
                      </p>
                    </div>
                  </div>

                  <!-- Java Version -->
                  <div
                    class="flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center text-orange-600 dark:text-orange-400"
                    >
                      <Coffee class="w-4 h-4" />
                    </div>
                    <div class="min-w-0">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        Java Version
                      </p>
                      <p
                        class="text-sm font-semibold text-gray-900 dark:text-white"
                      >
                        {{ getJavaVersion(server.imageName) }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Resources -->
              <div class="space-y-2">
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

                  <!-- External Java Args -->
                  <div
                    class="col-span-full flex items-start gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
                  >
                    <div
                      class="w-9 h-9 shrink-0 rounded-lg bg-sky-100 dark:bg-sky-900/30 flex items-center justify-center text-sky-600 dark:text-sky-400"
                    >
                      <Code class="w-4 h-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p
                        class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                      >
                        External Java Args
                      </p>
                      <div
                        v-if="!editingExternalJavaArgs"
                        class="flex items-center gap-2"
                      >
                        <div class="flex flex-wrap gap-1">
                          <span
                            v-if="
                              !server.externalJavaArgs ||
                              server.externalJavaArgs.length === 0
                            "
                            class="text-sm font-semibold text-gray-900 dark:text-white"
                          >
                            None
                          </span>
                          <span
                            v-for="(arg, idx) in server.externalJavaArgs"
                            :key="idx"
                            class="inline-flex items-center px-1.5 py-0.5 rounded text-[11px] font-medium bg-sky-100 dark:bg-sky-900/30 text-sky-700 dark:text-sky-400 border border-sky-200 dark:border-sky-800"
                          >
                            {{ arg }}
                          </span>
                        </div>
                        <button
                          v-if="server.containerStatus !== 'running'"
                          class="text-gray-400 hover:text-primary transition-colors"
                          :disabled="externalJavaArgsLoading"
                          @click="
                            tempExternalJavaArgs = server.externalJavaArgs
                              ? [...server.externalJavaArgs]
                              : [];
                            newExternalArg = '';
                            editingExternalJavaArgs = true;
                          "
                        >
                          <Pencil class="w-3.5 h-3.5" />
                        </button>
                      </div>
                      <div v-else class="flex flex-col gap-2">
                        <div class="flex flex-wrap gap-1">
                          <span
                            v-for="(arg, idx) in tempExternalJavaArgs"
                            :key="idx"
                            class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-medium bg-sky-100 dark:bg-sky-900/30 text-sky-700 dark:text-sky-400 border border-sky-200 dark:border-sky-800"
                          >
                            {{ arg }}
                            <button
                              class="hover:text-red-500 transition-colors"
                              @click="tempExternalJavaArgs.splice(idx, 1)"
                            >
                              <XIcon class="w-3 h-3" />
                            </button>
                          </span>
                        </div>
                        <div class="flex items-center gap-2">
                          <input
                            v-model="newExternalArg"
                            type="text"
                            class="flex-1 min-w-0 text-sm bg-white dark:bg-neutral-700 border border-gray-200 dark:border-neutral-600 rounded-lg px-2 py-1 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:outline-none font-mono"
                            placeholder="-XX:+UseG1GC"
                            @paste="handlePasteExternalJavaArgs($event)"
                            @keyup.enter.prevent="
                              if (newExternalArg.trim()) {
                                tempExternalJavaArgs.push(
                                  newExternalArg.trim(),
                                );
                                newExternalArg = '';
                              }
                            "
                          />
                          <button
                            class="text-primary hover:text-primary/80 transition-colors"
                            @click="
                              if (newExternalArg.trim()) {
                                tempExternalJavaArgs.push(
                                  newExternalArg.trim(),
                                );
                                newExternalArg = '';
                              }
                            "
                          >
                            <Plus class="w-4 h-4" />
                          </button>
                          <button
                            class="text-green-500 hover:text-green-600 transition-colors"
                            :disabled="externalJavaArgsLoading"
                            @click="saveExternalJavaArgs"
                          >
                            <Check class="w-4 h-4" />
                          </button>
                          <button
                            class="text-red-500 hover:text-red-600 transition-colors"
                            :disabled="externalJavaArgsLoading"
                            @click="editingExternalJavaArgs = false"
                          >
                            <XIcon class="w-4 h-4" />
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Storage -->
              <div class="space-y-2">
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

              <!-- Live Metrics -->
              <div class="space-y-2">
                <p
                  class="text-[10px] text-gray-400 dark:text-neutral-500 uppercase tracking-wider font-semibold px-1"
                >
                  Live Metrics
                </p>
                <div class="grid grid-cols-2 sm:grid-cols-2 gap-3">
                  <!-- RAM -->
                  <div
                    class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
                  >
                    <div
                      class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
                    >
                      <div
                        class="w-9 h-9 shrink-0 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400"
                      >
                        <MemoryStick class="w-4 h-4" />
                      </div>
                      <div class="min-w-0">
                        <p
                          class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                        >
                          RAM
                        </p>
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white"
                        >
                          {{
                            latestMetric
                              ? formatBytes(latestMetric.ramUsage) +
                                " \u00B7 " +
                                formatBytes(latestMetric.ramLimit)
                              : "—"
                          }}
                        </p>
                      </div>
                    </div>
                    <template v-if="server?.containerStatus === 'running'">
                      <svg
                        class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
                        viewBox="0 0 100 32"
                        preserveAspectRatio="none"
                      >
                        <polyline
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          class="text-indigo-500"
                          :points="sparklinePoints(serverMetrics, 'ramUsage')"
                        />
                      </svg>
                    </template>
                    <template v-else>
                      <div
                        class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
                      >
                        <span
                          class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
                          >Server is down</span
                        >
                      </div>
                    </template>
                  </div>
                  <!-- CPU -->
                  <div
                    class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
                  >
                    <div
                      class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
                    >
                      <div
                        class="w-9 h-9 shrink-0 rounded-lg bg-green-100 dark:bg-green-900/30 flex items-center justify-center text-green-600 dark:text-green-400"
                      >
                        <Activity class="w-4 h-4" />
                      </div>
                      <div class="min-w-0">
                        <p
                          class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                        >
                          CPU
                        </p>
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white"
                        >
                          {{
                            latestMetric
                              ? latestMetric.cpu.toFixed(1) + "%"
                              : "—"
                          }}
                        </p>
                      </div>
                    </div>
                    <template v-if="server?.containerStatus === 'running'">
                      <svg
                        class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
                        viewBox="0 0 100 32"
                        preserveAspectRatio="none"
                      >
                        <polyline
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          class="text-emerald-500"
                          :points="sparklinePoints(serverMetrics, 'cpu')"
                        />
                      </svg>
                    </template>
                    <template v-else>
                      <div
                        class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
                      >
                        <span
                          class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
                          >Server is down</span
                        >
                      </div>
                    </template>
                  </div>
                  <!-- Online -->
                  <div
                    class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
                  >
                    <div
                      class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
                    >
                      <div
                        class="w-9 h-9 shrink-0 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400"
                      >
                        <Users class="w-4 h-4" />
                      </div>
                      <div class="min-w-0">
                        <p
                          class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                        >
                          Online
                        </p>
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white"
                        >
                          {{
                            latestMetric
                              ? latestMetric.online + "/" + latestMetric.max
                              : "—"
                          }}
                        </p>
                      </div>
                    </div>
                    <template v-if="server?.containerStatus === 'running'">
                      <svg
                        class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
                        viewBox="0 0 100 32"
                        preserveAspectRatio="none"
                      >
                        <polyline
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          class="text-sky-500"
                          :points="sparklinePoints(serverMetrics, 'online')"
                        />
                      </svg>
                    </template>
                    <template v-else>
                      <div
                        class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
                      >
                        <span
                          class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
                          >Server is down</span
                        >
                      </div>
                    </template>
                  </div>
                  <!-- TPS -->
                  <div
                    class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
                  >
                    <div
                      class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
                    >
                      <div
                        class="w-9 h-9 shrink-0 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center text-purple-600 dark:text-purple-400"
                      >
                        <Zap class="w-4 h-4" />
                      </div>
                      <div class="min-w-0">
                        <p
                          class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                        >
                          TPS
                        </p>
                        <p
                          class="text-sm font-semibold"
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
                            latestMetric?.tps != null
                              ? latestMetric.tps.toFixed(1)
                              : "—"
                          }}
                        </p>
                      </div>
                    </div>
                    <template v-if="server?.containerStatus === 'running'">
                      <svg
                        class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
                        viewBox="0 0 100 32"
                        preserveAspectRatio="none"
                      >
                        <polyline
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          class="text-purple-500"
                          :points="sparklinePoints(serverMetrics, 'tps')"
                        />
                      </svg>
                    </template>
                    <template v-else>
                      <div
                        class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
                      >
                        <span
                          class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
                          >Server is down</span
                        >
                      </div>
                    </template>
                  </div>
                  <!-- Net RX -->
                  <div
                    class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
                  >
                    <div
                      class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
                    >
                      <div
                        class="w-9 h-9 shrink-0 rounded-lg bg-cyan-100 dark:bg-cyan-900/30 flex items-center justify-center text-cyan-600 dark:text-cyan-400"
                      >
                        <ArrowDown class="w-4 h-4" />
                      </div>
                      <div class="min-w-0">
                        <p
                          class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                        >
                          Net RX
                        </p>
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white"
                        >
                          {{
                            latestMetric
                              ? latestMetric.rxRate.toFixed(1) + " KB/s"
                              : "—"
                          }}
                        </p>
                      </div>
                    </div>
                    <template v-if="server?.containerStatus === 'running'">
                      <svg
                        class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
                        viewBox="0 0 100 32"
                        preserveAspectRatio="none"
                      >
                        <polyline
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          class="text-cyan-500"
                          :points="sparklinePoints(serverMetrics, 'rxRate')"
                        />
                      </svg>
                    </template>
                    <template v-else>
                      <div
                        class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
                      >
                        <span
                          class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
                          >Server is down</span
                        >
                      </div>
                    </template>
                  </div>
                  <!-- Net TX -->
                  <div
                    class="relative flex items-center gap-3 p-3 rounded-xl bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700 group overflow-hidden"
                  >
                    <div
                      class="flex items-center gap-3 transition-all duration-300 group-hover:opacity-0 group-hover:scale-95 shrink-0"
                    >
                      <div
                        class="w-9 h-9 shrink-0 rounded-lg bg-teal-100 dark:bg-teal-900/30 flex items-center justify-center text-teal-600 dark:text-teal-400"
                      >
                        <ArrowUp class="w-4 h-4" />
                      </div>
                      <div class="min-w-0">
                        <p
                          class="text-[11px] text-gray-500 dark:text-neutral-400 uppercase tracking-wider font-semibold"
                        >
                          Net TX
                        </p>
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white"
                        >
                          {{
                            latestMetric
                              ? latestMetric.txRate.toFixed(1) + " KB/s"
                              : "—"
                          }}
                        </p>
                      </div>
                    </div>
                    <template v-if="server?.containerStatus === 'running'">
                      <svg
                        class="absolute inset-0 w-full h-full opacity-0 transition-all duration-300 group-hover:opacity-100 p-2"
                        viewBox="0 0 100 32"
                        preserveAspectRatio="none"
                      >
                        <polyline
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          class="text-teal-500"
                          :points="sparklinePoints(serverMetrics, 'txRate')"
                        />
                      </svg>
                    </template>
                    <template v-else>
                      <div
                        class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
                      >
                        <span
                          class="text-xs text-gray-500 dark:text-neutral-400 font-medium"
                          >Server is down</span
                        >
                      </div>
                    </template>
                  </div>
                </div>
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
  Camera,
  Check,
  ChevronDown,
  CloudDownload,
  Code,
  Coffee,
  Copy,
  FolderOpen,
  Globe,
  Hash,
  MemoryStick,
  OctagonAlert,
  Package,
  Pencil,
  Play,
  Plus,
  RefreshCw,
  RotateCcw,
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
const removeBeforeStart = ref(false);
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
const liveMetricsExpanded = ref(true);
const serverLogsRef = ref<{
  scrollToBottom: () => void;
  reconnect: (tail?: number) => void;
} | null>(null);
const editingRam = ref(false);
const tempRamGb = ref<number | null>(null);
const ramLoading = ref(false);
const editingExternalJavaArgs = ref(false);
const tempExternalJavaArgs = ref<string[]>([]);
const newExternalArg = ref("");
const externalJavaArgsLoading = ref(false);

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
    case "pulling_image":
      return "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400";
    case "exited":
      return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
    default:
      return "bg-gray-100 text-gray-800 dark:bg-neutral-700 dark:text-neutral-300";
  }
}

function getJavaVersion(imageName?: string): string {
  if (!imageName) return "—";
  const tag = imageName.split(":")[1] || "";
  const match = tag.match(/^java(\d+)$/);
  return match && match[1] ? match[1] : tag;
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
      err?.data?.detail ||
      err?.message ||
      "Failed to update external Java args";
    if (status === 409) {
      showToast("error", "Server running", { description: msg });
    } else {
      showToast("error", "Update failed", { description: msg });
    }
  } finally {
    externalJavaArgsLoading.value = false;
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
