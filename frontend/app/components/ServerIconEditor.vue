<script setup lang="ts">
import { X, Save, Image as ImageIcon } from "lucide-vue-next";

const props = defineProps<{
  modelValue: boolean;
  file: File | null;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", v: boolean): void;
  (e: "save", blob: Blob): void;
}>();


const pixelSize = ref(4);
const fileSize = ref(0);
const canvasRef = ref<HTMLCanvasElement | null>(null);
const offscreenRef = ref<HTMLCanvasElement | null>(null);
const sourceImage = ref<HTMLImageElement | null>(null);
const isProcessing = ref(false);

const formattedSize = computed(() => {
  if (fileSize.value === 0) return "—";
  if (fileSize.value < 1024) return `${fileSize.value} B`;
  return `${(fileSize.value / 1024).toFixed(1)} KB`;
});

function close() {
  emit("update:modelValue", false);
}

function loadImage(file: File): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = reject;
    img.src = URL.createObjectURL(file);
  });
}

function drawPixelated() {
  const canvas = canvasRef.value;
  const offscreen = offscreenRef.value;
  const img = sourceImage.value;
  if (!canvas || !offscreen || !img) return;

  const ctx = canvas.getContext("2d");
  const offCtx = offscreen.getContext("2d");
  if (!ctx || !offCtx) return;

  const size = 64;
  const ps = Math.max(1, Math.min(size, pixelSize.value));
  const small = Math.max(1, Math.floor(size / ps));

  // Resize offscreen canvas to small size
  offscreen.width = small;
  offscreen.height = small;

  // Draw source image covering the small canvas (center crop)
  const srcRatio = img.naturalWidth / img.naturalHeight;
  let sx = 0;
  let sy = 0;
  let sWidth = img.naturalWidth;
  let sHeight = img.naturalHeight;
  if (srcRatio > 1) {
    sWidth = img.naturalHeight;
    sx = (img.naturalWidth - sWidth) / 2;
  } else if (srcRatio < 1) {
    sHeight = img.naturalWidth;
    sy = (img.naturalHeight - sHeight) / 2;
  }
  offCtx.clearRect(0, 0, small, small);
  offCtx.drawImage(img, sx, sy, sWidth, sHeight, 0, 0, small, small);

  // Scale back up to 64x64 with nearest-neighbor
  ctx.clearRect(0, 0, size, size);
  ctx.imageSmoothingEnabled = false;
  ctx.drawImage(offscreen, 0, 0, small, small, 0, 0, size, size);

  // Update file size estimate
  canvas.toBlob(
    (blob) => {
      if (blob) {
        fileSize.value = blob.size;
      }
    },
    "image/png",
    1.0,
  );
}

async function processFile() {
  if (!props.file) return;
  isProcessing.value = true;
  try {
    const img = await loadImage(props.file);
    if (sourceImage.value) {
      URL.revokeObjectURL(sourceImage.value.src);
    }
    sourceImage.value = img;
    drawPixelated();
  } catch {
    // ignore
  } finally {
    isProcessing.value = false;
  }
}

function save() {
  const canvas = canvasRef.value;
  if (!canvas) return;
  canvas.toBlob(
    (blob) => {
      if (!blob) return;
      emit("save", blob);
      close();
    },
    "image/png",
    1.0,
  );
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      pixelSize.value = 4;
      fileSize.value = 0;
      nextTick(() => processFile());
    } else {
      if (sourceImage.value) {
        URL.revokeObjectURL(sourceImage.value.src);
        sourceImage.value = null;
      }
    }
  },
);

watch(pixelSize, () => {
  drawPixelated();
});
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
      >
        <div
          class="absolute inset-0 bg-black/60 backdrop-blur-sm"
          @click="close"
        />
        <div
          class="relative w-full max-w-md bg-white dark:bg-neutral-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-neutral-700 overflow-hidden"
        >
          <!-- Header -->
          <div
            class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-neutral-700"
          >
            <div class="flex items-center gap-2">
              <ImageIcon class="w-5 h-5 text-primary" />
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                Edit Server Icon
              </h3>
            </div>
            <button
              class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors"
              @click="close"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Content -->
          <div class="p-6 space-y-6">
            <!-- Preview -->
            <div class="flex justify-center">
              <div
                class="relative w-64 h-64 rounded-xl overflow-hidden bg-neutral-900 ring-4 ring-gray-100 dark:ring-neutral-700 shadow-inner"
              >
                <canvas
                  ref="canvasRef"
                  width="64"
                  height="64"
                  class="w-full h-full image-render-pixelated"
                />
                <div
                  v-if="isProcessing"
                  class="absolute inset-0 flex items-center justify-center bg-black/30"
                >
                  <span class="text-white text-sm font-medium">Loading…</span>
                </div>
              </div>
            </div>

            <!-- Hidden offscreen canvas -->
            <canvas ref="offscreenRef" class="hidden" />

            <!-- Pixel Size Slider -->
            <div>
              <div class="flex items-center justify-between mb-2">
                <label
                  class="text-sm font-medium text-gray-700 dark:text-neutral-300"
                >
                  Pixel Size
                </label>
                <span class="text-sm font-mono text-primary font-semibold">
                  {{ pixelSize }} px
                </span>
              </div>
              <input
                v-model.number="pixelSize"
                type="range"
                min="1"
                max="32"
                class="w-full h-2 bg-gray-200 dark:bg-neutral-700 rounded-lg appearance-none cursor-pointer accent-primary"
              />
              <div
                class="flex justify-between mt-1 text-[11px] text-gray-400 dark:text-neutral-500"
              >
                <span>Smooth</span>
                <span>Pixelated</span>
              </div>
            </div>

            <!-- File Size -->
            <div
              class="flex items-center justify-between p-3 rounded-lg bg-gray-50 dark:bg-neutral-700/50 border border-gray-100 dark:border-neutral-700"
            >
              <span class="text-sm text-gray-500 dark:text-neutral-400">
                Output Size
              </span>
              <span
                class="text-sm font-mono font-semibold text-gray-900 dark:text-white"
              >
                {{ formattedSize }}
              </span>
            </div>
          </div>

          <!-- Footer -->
          <div
            class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-neutral-700 bg-gray-50 dark:bg-neutral-800/50"
          >
            <button
              class="px-4 py-2 rounded-lg text-sm font-medium text-gray-600 dark:text-neutral-300 hover:bg-gray-100 dark:hover:bg-neutral-700 transition-colors"
              @click="close"
            >
              Cancel
            </button>
            <button
              class="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-primary hover:bg-primary/90 text-white transition-colors shadow-sm"
              @click="save"
            >
              <Save class="w-4 h-4" />
              Save Icon
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.image-render-pixelated {
  image-rendering: pixelated;
  image-rendering: crisp-edges;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
