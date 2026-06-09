export type LogLevel = "error" | "warn" | "info" | "debug" | "trace" | "fatal";

const LEVEL_RE = /\[(?:[^\]]+\/)?(ERROR|WARN|INFO|DEBUG|TRACE|FATAL)\]/;

export function parseLogLevel(line: string): LogLevel | null {
  const m = LEVEL_RE.exec(line);
  if (m && m[1]) {
    return m[1].toLowerCase() as LogLevel;
  }

  const trimmed = line.trimStart();

  if (
    trimmed.startsWith("Exception") ||
    trimmed.startsWith("Caused by:") ||
    trimmed.startsWith("Error:")
  ) {
    return "error";
  }

  if (/^at\s+\S+\s*\(/.test(trimmed)) {
    return "trace";
  }

  return null;
}

export function getLogLevelClass(
  level: LogLevel | null,
  colored: boolean,
): string {
  if (!colored || !level) {
    return "text-gray-700 dark:text-neutral-300";
  }

  switch (level) {
    case "error":
    case "fatal":
      return "text-red-600 dark:text-red-400";
    case "warn":
      return "text-yellow-600 dark:text-yellow-400";
    case "info":
      return "text-gray-700 dark:text-neutral-300";
    case "debug":
      return "text-gray-500 dark:text-neutral-500";
    case "trace":
      return "text-gray-400 dark:text-neutral-600";
    default:
      return "text-gray-700 dark:text-neutral-300";
  }
}
