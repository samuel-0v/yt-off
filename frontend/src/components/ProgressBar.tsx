import { formatProgress } from '../utils/format'

type ProgressBarProps = {
  progress: number
  eta?: string
  speed?: string
}

export function ProgressBar({ progress, eta, speed }: ProgressBarProps) {
  const normalizedProgress = Math.min(100, Math.max(0, progress || 0))
  const label = `${formatProgress(normalizedProgress)}%`

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-center justify-between gap-2 text-sm">
        <span className="font-medium text-zinc-700 dark:text-zinc-200">Progresso</span>
        <div className="flex items-center gap-3 text-zinc-600 dark:text-zinc-300">
          {speed ? <span className="tabular-nums">{speed}</span> : null}
          {eta ? <span className="tabular-nums">ETA {eta}</span> : null}
          <span className="font-semibold tabular-nums text-zinc-800 dark:text-zinc-100">{label}</span>
        </div>
      </div>
      <div className="h-3 overflow-hidden rounded-full bg-zinc-200 shadow-inner dark:bg-zinc-800">
        <div
          className="h-full rounded-full bg-gradient-to-r from-emerald-500 via-sky-500 to-cyan-500 transition-all duration-500 ease-out"
          style={{ width: `${normalizedProgress}%` }}
        />
      </div>
    </div>
  )
}
