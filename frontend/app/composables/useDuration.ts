import { onScopeDispose, ref, type Ref } from 'vue'

export function formatDuration(ms: number): string {
  if (ms < 0) return '0s'

  const seconds = Math.floor(ms / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) {
    const remHours = hours % 24
    return remHours > 0 ? `${days}d ${remHours}h` : `${days}d`
  }
  if (hours > 0) {
    const remMinutes = minutes % 60
    return remMinutes > 0 ? `${hours}h ${remMinutes}m` : `${hours}h`
  }
  if (minutes > 0) {
    const remSeconds = seconds % 60
    return remSeconds > 0 ? `${minutes}m ${remSeconds}s` : `${minutes}m`
  }
  return `${seconds}s`
}

export function useUptime(startedAt: string | undefined): Ref<string> {
  const uptime = ref('')

  function update() {
    if (!startedAt) {
      uptime.value = ''
      return
    }
    const start = new Date(startedAt).getTime()
    const now = Date.now()
    uptime.value = formatDuration(now - start)
  }

  update()
  const interval = setInterval(update, 1000)

  onScopeDispose(() => clearInterval(interval))

  return uptime
}
