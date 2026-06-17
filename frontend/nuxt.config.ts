/// <reference types="node" />
// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,
  compatibilityDate: "2025-07-15",
  devtools: { enabled: false },
  modules: ["@nuxtjs/tailwindcss", "@nuxtjs/color-mode"],
  colorMode: {
    preference: "system",
    fallback: "light",
    classSuffix: "",
  },
  tailwindcss: {
    viewer: false,
  },
  app: {
    head: {
      title: "MiniMin",
      meta: [
        { charset: "utf-8" },
        { name: "viewport", content: "width=device-width, initial-scale=1" },
      ],
      link: [
        { rel: "icon", type: "image/x-icon", href: "/favicon/favicon.ico" },
        {
          rel: "icon",
          type: "image/png",
          sizes: "96x96",
          href: "/favicon/favicon-96x96.png",
        },
        { rel: "icon", type: "image/svg+xml", href: "/favicon/favicon.svg" },
        {
          rel: "apple-touch-icon",
          sizes: "180x180",
          href: "/favicon/apple-touch-icon.png",
        },
      ],
    },
  },
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL || "http://localhost:8081",
      wsBase: process.env.WS_BASE_URL || "",
    },
  },
  nitro: {
    // Proxy is handled by server/middleware/proxy.ts (production)
  },
});
