<template>
  <div
    class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-background-dark relative overflow-hidden"
  >
    <!-- Abstract background shapes -->
    <div class="absolute inset-0 opacity-20">
      <div
        class="absolute -top-24 -left-24 w-96 h-96 rounded-full bg-primary blur-3xl animate-blob"
      />
      <div
        class="absolute top-1/2 right-0 w-80 h-80 rounded-full bg-primary-300 blur-3xl animate-blob animation-delay-2000"
      />
      <div
        class="absolute -bottom-32 left-1/3 w-96 h-96 rounded-full bg-primary-700 blur-3xl animate-blob animation-delay-4000"
      />
    </div>

    <!-- Dot pattern overlay -->
    <div
      class="absolute inset-0 opacity-10 dark:hidden"
      style="
        background-image: radial-gradient(
          circle,
          rgba(0, 0, 0, 0.5) 1px,
          transparent 1px
        );
        background-size: 24px 24px;
      "
    />
    <div
      class="absolute inset-0 opacity-10 hidden dark:block"
      style="
        background-image: radial-gradient(
          circle,
          rgba(255, 255, 255, 0.5) 1px,
          transparent 1px
        );
        background-size: 24px 24px;
      "
    />

    <div class="relative z-10 text-center max-w-md px-6">
      <div class="mb-6 flex items-center justify-center gap-3">
        <img src="/img/MiniMin_L.avif" alt="MiniMin" class="h-16 w-auto" />
        <img
          src="/img/MiniMin_T.avif"
          alt="MiniMin"
          class="h-12 w-auto dark:hidden"
        />
        <img
          src="/img/MiniMin_T_light.avif"
          alt="MiniMin"
          class="h-12 w-auto hidden dark:block"
        />
      </div>

      <h1 class="text-7xl font-bold text-gray-900 dark:text-white mb-2">404</h1>
      <p class="text-xl text-gray-600 dark:text-neutral-300 mb-2">
        Page not found
      </p>
      <p class="text-sm text-gray-500 dark:text-neutral-400 mb-8">
        The page you are looking for does not exist or has been moved.
      </p>

      <NuxtLink
        to="/"
        class="inline-flex items-center gap-2 px-6 py-3 rounded-lg bg-primary text-white hover:bg-primary/90 transition-colors font-semibold shadow-lg shadow-primary/25"
      >
        <svg
          class="w-5 h-5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18"
          />
        </svg>
        Go back home
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  noSidebar: true,
});

usePageTitle("Page Not Found");
</script>

<style scoped>
@keyframes blob {
  0% {
    transform: translate(0px, 0px) scale(1);
  }
  33% {
    transform: translate(30px, -50px) scale(1.1);
  }
  66% {
    transform: translate(-20px, 20px) scale(0.9);
  }
  100% {
    transform: translate(0px, 0px) scale(1);
  }
}

.animate-blob {
  animation: blob 7s infinite;
}

.animation-delay-2000 {
  animation-delay: 2s;
}

.animation-delay-4000 {
  animation-delay: 4s;
}
</style>
