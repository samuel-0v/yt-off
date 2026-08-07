import { Clock, FileDown, Gauge, HardDrive } from 'lucide-react'
import type { ReactNode } from 'react'
import { ProgressBar } from './ProgressBar'
import type { DownloadTask } from '../types/download'
import { formatBytes, statusClassName, statusLabel } from '../utils/format'

export type DownloadCardItem = DownloadTask & {
  title?: string
  formatLabel?: string
}

type DownloadCardProps = {
  download: DownloadCardItem
}

export function DownloadCard({ download }: DownloadCardProps) {
  return (
    <article className="rounded-lg border border-zinc-200 bg-white p-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex min-w-0 gap-3">
          <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300">
            <FileDown aria-hidden="true" className="h-5 w-5" />
          </div>
          <div className="min-w-0">
            <h3 className="truncate text-sm font-semibold text-zinc-950 dark:text-zinc-50">
            {download.title ?? download.filename ?? download.id}
            </h3>
            {download.formatLabel ? (
              <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-300">{download.formatLabel}</p>
            ) : null}
            {download.owner_username ? (
              <p className="mt-1 text-xs font-semibold text-zinc-500 dark:text-zinc-400">{download.owner_username}</p>
            ) : null}
          </div>
        </div>
        <span className={statusClassName(download.status)}>{statusLabel(download.status)}</span>
      </div>

      <div className="mt-4">
        <ProgressBar eta={download.eta} progress={download.progress} speed={download.speed} />
      </div>

      <dl className="mt-4 grid gap-2 text-sm text-zinc-600 sm:grid-cols-3 dark:text-zinc-300">
        <Metric icon={<Gauge aria-hidden="true" className="h-4 w-4" />} label="Velocidade" value={download.speed || '-'} />
        <Metric icon={<Clock aria-hidden="true" className="h-4 w-4" />} label="ETA" value={download.eta || '-'} />
        <Metric
          icon={<HardDrive aria-hidden="true" className="h-4 w-4" />}
          label="Tamanho"
          value={download.file_size ? formatBytes(download.file_size) : '-'}
        />
      </dl>

      {download.error ? (
        <p className="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200">
          {download.error}
        </p>
      ) : null}
    </article>
  )
}

function Metric({ icon, label, value }: { icon: ReactNode; label: string; value: string }) {
  return (
    <div className="rounded-lg bg-zinc-50 px-3 py-2 dark:bg-zinc-950">
      <dt className="flex items-center gap-1.5 text-xs font-medium text-zinc-500 dark:text-zinc-400">
        {icon}
        {label}
      </dt>
      <dd className="mt-1 truncate font-semibold tabular-nums text-zinc-800 dark:text-zinc-100">{value}</dd>
    </div>
  )
}
