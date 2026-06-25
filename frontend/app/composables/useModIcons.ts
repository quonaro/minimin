import type { ModInfo } from "./useMods";

interface BatchResponse {
  icons: Record<string, string>;
}

const cache = new Map<string, string>();
const iconUrls = ref<Record<string, string>>({});

function cacheKey(serverId: string, type: string, filename: string): string {
  return `${serverId}:${type}:${filename}`;
}

export function useModIcons() {
  const pending = ref(false);

  async function loadIcons(
    serverId: string,
    type: "server" | "client",
    mods: ModInfo[]
  ) {
    const needed = mods
      .filter((m) => !m.corrupted)
      .map((m) => m.filename)
      .filter((f) => !cache.has(cacheKey(serverId, type, f)));

    if (needed.length === 0) return;

    pending.value = true;
    try {
      const endpoint = type === "server" ? "mods" : "client-mods";
      const res = await $fetch<BatchResponse>(
        `${useApiBase()}/servers/${serverId}/${endpoint}/icons`,
        {
          method: "POST",
          body: { filenames: needed },
          credentials: "include",
        }
      );
      const icons = res.icons ?? {};
      const next: Record<string, string> = { ...iconUrls.value };
      for (const [filename, dataUrl] of Object.entries(icons)) {
        const key = cacheKey(serverId, type, filename);
        cache.set(key, dataUrl);
        next[key] = dataUrl;
      }
      iconUrls.value = next;
    } catch {
      // silently fail; placeholders remain
    } finally {
      pending.value = false;
    }
  }

  function getIconUrl(
    serverId: string,
    type: "server" | "client",
    filename: string
  ): string {
    return iconUrls.value[cacheKey(serverId, type, filename)] || "";
  }

  return { pending, loadIcons, getIconUrl };
}
