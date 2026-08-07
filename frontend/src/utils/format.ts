import type { DownloadStatus } from '../types/download'

export function formatBytes(bytes?: number): string {
  if (!bytes || bytes < 0) {
    return 'Indisponível'
  }

  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB']
  let value = bytes
  let unitIndex = 0

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }

  return `${value.toFixed(value >= 10 ? 0 : 1)} ${units[unitIndex]}`
}

export function formatDuration(totalSeconds: number): string {
  if (!totalSeconds || totalSeconds < 0) {
    return '--:--'
  }

  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  const paddedMinutes = minutes.toString().padStart(2, '0')
  const paddedSeconds = seconds.toString().padStart(2, '0')

  if (hours > 0) {
    return `${hours}:${paddedMinutes}:${paddedSeconds}`
  }

  return `${minutes}:${paddedSeconds}`
}

export function formatDate(value?: string): string {
  if (!value) {
    return ''
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }

  return new Intl.DateTimeFormat('pt-BR', {
    dateStyle: 'short',
    timeStyle: 'short',
  }).format(date)
}

export function formatProgress(progress: number): string {
  const normalizedProgress = Math.min(100, Math.max(0, progress || 0))

  if (Number.isInteger(normalizedProgress)) {
    return normalizedProgress.toString()
  }

  return normalizedProgress.toFixed(1)
}

export function statusLabel(status: DownloadStatus): string {
  const labels: Record<DownloadStatus, string> = {
    queued: 'Na fila',
    running: 'Baixando',
    completed: 'Concluído',
    failed: 'Falhou',
    cancelled: 'Cancelado',
  }

  return labels[status] ?? status
}

export function statusClassName(status: DownloadStatus): string {
  const baseClassName = 'inline-flex w-fit items-center rounded-md px-2.5 py-1 text-xs font-semibold'

  if (status === 'completed') {
    return `${baseClassName} bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-300`
  }

  if (status === 'failed') {
    return `${baseClassName} bg-red-100 text-red-800 dark:bg-red-500/15 dark:text-red-300`
  }

  if (status === 'cancelled') {
    return `${baseClassName} bg-zinc-200 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300`
  }

  if (status === 'running') {
    return `${baseClassName} bg-sky-100 text-sky-800 dark:bg-sky-500/15 dark:text-sky-300`
  }

  return `${baseClassName} bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-300`
}
