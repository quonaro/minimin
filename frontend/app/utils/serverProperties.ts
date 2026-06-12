export type ServerPropType = "text" | "number" | "boolean" | "select";

export interface KnownProperty {
  key: string;
  label: string;
  type: ServerPropType;
  options?: string[];
  group: string;
  dangerous?: boolean;
}

export const knownProperties: KnownProperty[] = [
  {
    key: "level-name",
    label: "Level Name",
    type: "text",
    group: "DANGEROUS",
  },
  {
    key: "level-seed",
    label: "Level Seed",
    type: "text",
    group: "DANGEROUS",
  },
  {
    key: "level-type",
    label: "Level Type",
    type: "select",
    options: [
      "minecraft:normal",
      "minecraft:flat",
      "minecraft:large_biomes",
      "minecraft:amplified",
      "minecraft:single_biome_surface",
    ],
    group: "DANGEROUS",
  },

  { key: "motd", label: "MOTD", type: "text", group: "Network" },
  {
    key: "max-players",
    label: "Max Players",
    type: "number",
    group: "Network",
  },
  {
    key: "server-port",
    label: "Server Port",
    type: "number",
    group: "Network",
    dangerous: true,
  },
  {
    key: "online-mode",
    label: "Online Mode",
    type: "boolean",
    group: "Network",
    dangerous: true,
  },
  {
    key: "white-list",
    label: "White List",
    type: "boolean",
    group: "Network",
    dangerous: true,
  },
  {
    key: "enforce-secure-profile",
    label: "Enforce Secure Profile",
    type: "boolean",
    group: "Network",
    dangerous: true,
  },

  {
    key: "difficulty",
    label: "Difficulty",
    type: "select",
    options: ["peaceful", "easy", "normal", "hard"],
    group: "Gameplay",
  },
  {
    key: "gamemode",
    label: "Game Mode",
    type: "select",
    options: ["survival", "creative", "adventure", "spectator"],
    group: "Gameplay",
  },
  { key: "pvp", label: "PvP", type: "boolean", group: "Gameplay" },
  {
    key: "hardcore",
    label: "Hardcore",
    type: "boolean",
    group: "Gameplay",
    dangerous: true,
  },
  {
    key: "allow-flight",
    label: "Allow Flight",
    type: "boolean",
    group: "Gameplay",
  },
  {
    key: "spawn-protection",
    label: "Spawn Protection",
    type: "number",
    group: "Gameplay",
  },

  {
    key: "generate-structures",
    label: "Generate Structures",
    type: "boolean",
    group: "World",
  },
  {
    key: "allow-nether",
    label: "Allow Nether",
    type: "boolean",
    group: "World",
  },

  {
    key: "view-distance",
    label: "View Distance",
    type: "number",
    group: "Performance",
  },
  {
    key: "simulation-distance",
    label: "Simulation Distance",
    type: "number",
    group: "Performance",
  },

  {
    key: "enable-command-block",
    label: "Enable Command Block",
    type: "boolean",
    group: "Advanced",
  },
];

export const selectOptions: Record<string, string[]> = {};
for (const kp of knownProperties) {
  if (kp.type === "select" && kp.options) {
    selectOptions[kp.key] = kp.options;
  } else if (kp.type === "boolean") {
    selectOptions[kp.key] = ["true", "false"];
  }
}

export function parseProperties(content: string): Record<string, string> {
  const props: Record<string, string> = {};
  for (const line of content.split("\n")) {
    const trimmed = line.trim();
    if (trimmed.startsWith("#") || trimmed === "") continue;
    const idx = trimmed.indexOf("=");
    if (idx === -1) continue;
    const key = trimmed.slice(0, idx).trim();
    const val = trimmed.slice(idx + 1);
    if (key) props[key] = val;
  }
  return props;
}

export const knownKeys = new Set(knownProperties.map((k) => k.key));
export const groupOrder = ["Network", "Gameplay", "World", "Performance", "Advanced"];
