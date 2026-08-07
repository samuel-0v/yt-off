import {
  Download,
  FileAudio,
  FileVideo,
  FolderPlus,
  PauseCircle,
  PlayCircle,
  Trash2,
  XCircle,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import type { DownloadTask, FileInfo } from '../types/download'
import { formatBytes, formatDate, statusClassName, statusLabel } from '../utils/format'
import { ProgressBar } from './ProgressBar'

type FileCardProps = {
  download?: DownloadTask
  file?: FileInfo
  deleting: boolean
  cancelling: boolean
  downloadUrl?: string
  previewUrl?: string
  selectable?: boolean
  selected?: boolean
  showOwner?: boolean
  onDelete: (name: string) => void
  onCancel: (id: string) => void
  onAddToGroup?: (download: DownloadTask) => void
  onSelectionChange?: (name: string, selected: boolean) => void
}

export function FileCard({
  download,
  file,
  deleting,
  cancelling,
  downloadUrl,
  previewUrl,
  selectable = false,
  selected = false,
  showOwner = false,
  onDelete,
  onCancel,
  onAddToGroup,
  onSelectionChange,
}: FileCardProps) {
  const [playerOpen, setPlayerOpen] = useState(false)
  const name = file?.name ?? download?.filename ?? download?.id ?? 'Arquivo'
  const size = file?.size ?? download?.file_size ?? 0
  const extension = file?.extension ?? download?.extension
  const available = Boolean(file)
  const active = download?.status === 'queued' || download?.status === 'running'
  const mediaKind = getMediaKind(extension)
  const playable = Boolean(available && previewUrl && mediaKind)

  useEffect(() => {
    setPlayerOpen(false)
  }, [name])

  return (
    <article className="rounded-lg border border-zinc-200 bg-white p-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
      <div className="grid gap-4 md:grid-cols-[112px_minmax(0,1fr)_auto] md:items-start">
        <div className="flex aspect-video w-full items-center justify-center rounded-lg bg-zinc-100 text-sm font-bold uppercase text-zinc-500 md:w-28 dark:bg-zinc-800 dark:text-zinc-300">
          {mediaKind === 'video' ? (
            <FileVideo aria-hidden="true" className="h-7 w-7" />
          ) : mediaKind === 'audio' ? (
            <FileAudio aria-hidden="true" className="h-7 w-7" />
          ) : (
            extension || 'file'
          )}
        </div>

        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            {selectable ? (
              <label className="inline-flex min-h-8 cursor-pointer items-center gap-2 rounded-md border border-zinc-200 px-2.5 text-xs font-semibold text-zinc-700 transition hover:bg-zinc-50 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800">
                <input
                  checked={selected}
                  className="h-4 w-4 rounded border-zinc-300 text-emerald-600 focus:ring-emerald-500"
                  type="checkbox"
                  onChange={(event) => onSelectionChange?.(name, event.target.checked)}
                />
                Selecionar
              </label>
            ) : null}
            {download ? <span className={statusClassName(download.status)}>{statusLabel(download.status)}</span> : null}
            {!available && !active ? (
              <span className="rounded-md bg-zinc-100 px-2.5 py-1 text-xs font-semibold text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300">
                Arquivo removido
              </span>
            ) : null}
          </div>

          <h2 className="mt-3 truncate text-base font-semibold text-zinc-950 dark:text-zinc-50">{name}</h2>
          <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-sm text-zinc-600 dark:text-zinc-300">
            <span>{size > 0 ? formatBytes(size) : 'Tamanho indisponível'}</span>
            {extension ? <span>{extension.toUpperCase()}</span> : null}
            {download?.created_at ? <span>{formatDate(download.created_at)}</span> : null}
            {showOwner && download?.owner_username ? <span>{download.owner_username}</span> : null}
          </div>

          {download && (download.status === 'queued' || download.status === 'running') ? (
            <div className="mt-4">
              <ProgressBar eta={download.eta} progress={download.progress} speed={download.speed} />
            </div>
          ) : null}

          {download?.error ? (
            <p className="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200">
              {download.error}
            </p>
          ) : null}

          {playerOpen && playable ? (
            <div className="mt-4 rounded-lg border border-zinc-200 bg-zinc-50 p-3 dark:border-zinc-800 dark:bg-zinc-950">
              {mediaKind === 'video' ? (
                <video
                  className="aspect-video w-full rounded-lg bg-black"
                  controls
                  playsInline
                  preload="metadata"
                  src={previewUrl}
                />
              ) : (
                <audio className="w-full" controls preload="metadata" src={previewUrl} />
              )}
            </div>
          ) : null}
        </div>

        <div className="flex flex-wrap gap-2 md:justify-end">
          {playable ? (
            <button
              className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg border border-zinc-200 px-3 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-100 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
              type="button"
              onClick={() => setPlayerOpen((current) => !current)}
            >
              {playerOpen ? (
                <PauseCircle aria-hidden="true" className="h-4 w-4" />
              ) : (
                <PlayCircle aria-hidden="true" className="h-4 w-4" />
              )}
              {playerOpen ? 'Ocultar' : 'Reproduzir'}
            </button>
          ) : null}

          {active && download ? (
            <button
              className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg border border-red-200 px-3 text-sm font-semibold text-red-700 transition hover:bg-red-50 disabled:cursor-not-allowed disabled:border-zinc-200 disabled:text-zinc-400 dark:border-red-500/30 dark:text-red-300 dark:hover:bg-red-500/10"
              disabled={cancelling}
              type="button"
              onClick={() => onCancel(download.id)}
            >
              <XCircle aria-hidden="true" className="h-4 w-4" />
              {cancelling ? 'Cancelando...' : 'Cancelar'}
            </button>
          ) : available && downloadUrl ? (
            <a
              className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg bg-zinc-950 px-3 text-sm font-semibold text-white transition hover:bg-zinc-800 dark:bg-emerald-500 dark:text-zinc-950 dark:hover:bg-emerald-400"
              download
              href={downloadUrl}
            >
              <Download aria-hidden="true" className="h-4 w-4" />
              Baixar
            </a>
          ) : (
            <button
              className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg bg-zinc-200 px-3 text-sm font-semibold text-zinc-500 dark:bg-zinc-800 dark:text-zinc-500"
              disabled
              type="button"
            >
              <Download aria-hidden="true" className="h-4 w-4" />
              Baixar
            </button>
          )}

          {download && onAddToGroup ? (
            <button
              className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg border border-zinc-200 px-3 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-100 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
              type="button"
              onClick={() => onAddToGroup(download)}
            >
              <FolderPlus aria-hidden="true" className="h-4 w-4" />
              Grupo
            </button>
          ) : null}

          {!active ? (
            <button
              className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg border border-red-200 px-3 text-sm font-semibold text-red-700 transition hover:bg-red-50 disabled:cursor-not-allowed disabled:border-zinc-200 disabled:text-zinc-400 dark:border-red-500/30 dark:text-red-300 dark:hover:bg-red-500/10"
              disabled={!available || deleting}
              type="button"
              onClick={() => onDelete(name)}
            >
              <Trash2 aria-hidden="true" className="h-4 w-4" />
              {deleting ? 'Excluindo...' : 'Excluir'}
            </button>
          ) : null}
        </div>
      </div>
    </article>
  )
}

function getMediaKind(extension?: string): 'audio' | 'video' | '' {
  switch ((extension ?? '').toLowerCase()) {
    case 'mp4':
    case 'm4v':
    case 'mov':
    case 'webm':
      return 'video'
    case 'm4a':
    case 'mp3':
    case 'aac':
    case 'opus':
    case 'ogg':
    case 'oga':
    case 'wav':
    case 'flac':
      return 'audio'
    default:
      return ''
  }
}
