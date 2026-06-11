import { ref, watch, type Ref } from "vue";

export function useDebounce<T>(source: Ref<T>, delay = 150): Ref<T> {
  const debounced = ref(source.value) as Ref<T>;
  let timer: ReturnType<typeof setTimeout> | null = null;

  watch(
    source,
    (val) => {
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => {
        debounced.value = val;
      }, delay);
    },
    { immediate: false },
  );

  return debounced;
}

export function useDebouncedRef<T>(initial: T, delay = 150): Ref<T> {
  const raw = ref(initial) as Ref<T>;
  const debounced = ref(initial) as Ref<T>;
  let timer: ReturnType<typeof setTimeout> | null = null;

  watch(raw, (val) => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => {
      debounced.value = val;
    }, delay);
  });

  return debounced;
}
