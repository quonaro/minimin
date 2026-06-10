import { toValue, type MaybeRefOrGetter } from "vue";

export function usePageTitle(title: MaybeRefOrGetter<string>) {
  useHead(() => ({
    title: toValue(title) ? `${toValue(title)} | MiniMin` : "MiniMin",
  }));
}
