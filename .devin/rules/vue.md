---
description: Rules for writing Vue code in the webui frontend
---

# Vue Coding Rules — WebUI Frontend

## File Size
- **Maximum 600 lines per file, including `<template>`, `<script>`, and `<style>`.**
- If a component exceeds this limit, decompose it:
  - Extract child components.
  - Move composables / utilities to `composables/` or `utils/`.
  - Move large `<style>` blocks to scoped CSS modules or Tailwind utilities.

## Component Structure
Use this order inside `.vue` files:
1. `<script setup lang="ts">`
2. `<template>`
3. `<style scoped>` (only if Tailwind classes are insufficient)

## TypeScript
- Strict mode is on. No `any` without a documented reason.
- Define props with `defineProps<{}>()` and emits with `defineEmits<{}>()`.
- Prefer interfaces over types for object shapes.

## Composition API
- Use `<script setup>` and Composition API exclusively. No Options API.
- Extract reusable logic into `composables/` (e.g., `useServer()`, `useConsole()`).
- Keep components focused: one primary responsibility per component.

## Templates
- Use `kebab-case` for custom components in templates.
- Always provide `key` attributes in `v-for`.
- Avoid complex expressions in templates; use computed properties.

## Styling
- TailwindCSS is the default. Use utility classes first.
- Custom scoped styles only for complex overrides or third-party component theming.
- Never use element selectors (e.g., `div { ... }`) in scoped styles.

## State & Fetching
- Server state: use Nuxt composables (`useFetch`, `$fetch`) or a lightweight Pinia store if caching / cross-component sharing is needed.
- Client-only state: Pinia or `ref`/`reactive` inside composables.
- Always handle loading and error states in async operations.

## Performance
- Lazy-load heavy components with `defineAsyncComponent`.
- Debounce rapid user input (search, resize handlers).
- Prefer `v-show` over `v-if` for toggling visibility when the DOM cost is high.

## Linting
- Run `pnpm lint` (or the project's lint command) before pushing.
- Fix all TypeScript and Vue compiler warnings.
