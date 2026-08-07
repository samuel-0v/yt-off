import {
  CheckCircle2,
  ChevronRight,
  FileAudio,
  FileVideo,
  HardDrive,
  Layers3,
  Music2,
  SlidersHorizontal,
} from 'lucide-react'
import { useEffect, useMemo, useState, type ReactNode } from 'react'
import type { DownloadOption } from '../types/download'
import { formatBytes } from '../utils/format'

type FormatSelectorProps = {
  options: DownloadOption[]
  selectedFormatId: string
  onSelect: (formatId: string) => void
}

type SelectionMode = 'video' | 'audio'

export function FormatSelector({ options, selectedFormatId, onSelect }: FormatSelectorProps) {
  const videoOptions = useMemo(() => options.filter((option) => option.has_video), [options])
  const audioOptions = useMemo(() => options.filter((option) => !option.has_video && option.has_audio), [options])
  const selectedOption = useMemo(
    () => options.find((option) => option.format_id === selectedFormatId),
    [options, selectedFormatId],
  )
  const [mode, setMode] = useState<SelectionMode>(videoOptions.length > 0 ? 'video' : 'audio')
  const [extension, setExtension] = useState('all')

  const modeOptions = mode === 'video' ? videoOptions : audioOptions
  const extensions = useMemo(() => uniqueExtensions(modeOptions), [modeOptions])
  const visibleOptions = useMemo(
    () => modeOptions.filter((option) => extension === 'all' || option.extension === extension),
    [extension, modeOptions],
  )

  useEffect(() => {
    if (mode === 'video' && videoOptions.length === 0 && audioOptions.length > 0) {
      setMode('audio')
    }
    if (mode === 'audio' && audioOptions.length === 0 && videoOptions.length > 0) {
      setMode('video')
    }
  }, [audioOptions.length, mode, videoOptions.length])

  useEffect(() => {
    if (!selectedOption) {
      return
    }

    setMode(selectedOption.has_video ? 'video' : 'audio')
  }, [selectedOption])

  if (options.length === 0) {
    return (
      <p className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-200">
        Nenhuma opção de qualidade foi encontrada para este vídeo.
      </p>
    )
  }

  function selectMode(nextMode: SelectionMode) {
    const nextOptions = nextMode === 'video' ? videoOptions : audioOptions
    setMode(nextMode)
    setExtension('all')
    if (nextOptions[0]) {
      onSelect(nextOptions[0].format_id)
    }
  }

  function selectExtension(nextExtension: string) {
    const nextOptions = modeOptions.filter((option) => nextExtension === 'all' || option.extension === nextExtension)
    setExtension(nextExtension)
    if (nextOptions.length > 0 && !nextOptions.some((option) => option.format_id === selectedFormatId)) {
      onSelect(nextOptions[0].format_id)
    }
  }

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
              <SlidersHorizontal aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <h2 className="text-base font-semibold text-zinc-950 dark:text-zinc-50">Escolha o arquivo</h2>
              <p className="text-sm text-zinc-600 dark:text-zinc-300">
                {videoOptions.length} opções de vídeo · {audioOptions.length} opções de áudio
              </p>
            </div>
          </div>
        </div>

        {selectedOption ? <SelectedSummary option={selectedOption} /> : null}
      </div>

      <div className="mt-5 grid gap-4 lg:grid-cols-[220px_minmax(0,1fr)]">
        <div className="space-y-4">
          <div>
            <p className="mb-2 text-xs font-semibold uppercase text-zinc-500 dark:text-zinc-400">Tipo</p>
            <div className="grid gap-2">
              <ModeButton
                active={mode === 'video'}
                count={videoOptions.length}
                disabled={videoOptions.length === 0}
                icon={<FileVideo aria-hidden="true" className="h-4 w-4" />}
                label="Vídeo + áudio"
                onClick={() => selectMode('video')}
              />
              <ModeButton
                active={mode === 'audio'}
                count={audioOptions.length}
                disabled={audioOptions.length === 0}
                icon={<Music2 aria-hidden="true" className="h-4 w-4" />}
                label="Somente áudio"
                onClick={() => selectMode('audio')}
              />
            </div>
          </div>

          <div>
            <p className="mb-2 text-xs font-semibold uppercase text-zinc-500 dark:text-zinc-400">Extensão</p>
            <div className="flex flex-wrap gap-2">
              <ExtensionButton
                active={extension === 'all'}
                label="Todos"
                onClick={() => selectExtension('all')}
              />
              {extensions.map((item) => (
                <ExtensionButton
                  active={extension === item}
                  key={item}
                  label={item.toUpperCase()}
                  onClick={() => selectExtension(item)}
                />
              ))}
            </div>
          </div>
        </div>

        <div className="space-y-3">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-sm font-semibold text-zinc-950 dark:text-zinc-50">
                {mode === 'video' ? 'Qualidades de vídeo' : 'Opções de áudio'}
              </p>
              <p className="text-sm text-zinc-600 dark:text-zinc-300">
                {visibleOptions.length} opções disponíveis
              </p>
            </div>
            <Layers3 aria-hidden="true" className="h-5 w-5 text-zinc-400" />
          </div>

          <div className="grid gap-2">
            {visibleOptions.map((option, index) => (
              <OptionRow
                key={option.format_id}
                option={option}
                recommended={mode === 'video' && index === 0}
                selected={option.format_id === selectedFormatId}
                onSelect={() => onSelect(option.format_id)}
              />
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}

function ModeButton({
  active,
  count,
  disabled,
  icon,
  label,
  onClick,
}: {
  active: boolean
  count: number
  disabled: boolean
  icon: ReactNode
  label: string
  onClick: () => void
}) {
  return (
    <button
      aria-pressed={active}
      className={[
        'flex min-h-11 items-center justify-between gap-3 rounded-lg border px-3 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-50',
        active
          ? 'border-emerald-500 bg-emerald-50 text-emerald-900 dark:border-emerald-400 dark:bg-emerald-400/10 dark:text-emerald-200'
          : 'border-zinc-200 bg-zinc-50 text-zinc-700 hover:bg-zinc-100 dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-200 dark:hover:bg-zinc-800',
      ].join(' ')}
      disabled={disabled}
      type="button"
      onClick={onClick}
    >
      <span className="flex items-center gap-2">
        {icon}
        {label}
      </span>
      <span className="rounded-md bg-white px-2 py-0.5 text-xs text-zinc-500 dark:bg-zinc-900 dark:text-zinc-400">
        {count}
      </span>
    </button>
  )
}

function ExtensionButton({ active, label, onClick }: { active: boolean; label: string; onClick: () => void }) {
  return (
    <button
      aria-pressed={active}
      className={[
        'min-h-9 rounded-lg border px-3 text-xs font-bold transition',
        active
          ? 'border-zinc-950 bg-zinc-950 text-white dark:border-emerald-400 dark:bg-emerald-400 dark:text-zinc-950'
          : 'border-zinc-200 bg-white text-zinc-700 hover:bg-zinc-100 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800',
      ].join(' ')}
      type="button"
      onClick={onClick}
    >
      {label}
    </button>
  )
}

function OptionRow({
  option,
  recommended,
  selected,
  onSelect,
}: {
  option: DownloadOption
  recommended: boolean
  selected: boolean
  onSelect: () => void
}) {
  const Icon = option.has_video ? FileVideo : FileAudio

  return (
    <button
      aria-pressed={selected}
      className={[
        'group rounded-lg border p-3 text-left transition',
        selected
          ? 'border-emerald-500 bg-emerald-50 ring-2 ring-emerald-100 dark:border-emerald-400 dark:bg-emerald-400/10 dark:ring-emerald-400/20'
          : 'border-zinc-200 bg-white hover:border-zinc-400 dark:border-zinc-800 dark:bg-zinc-950 dark:hover:border-zinc-600',
      ].join(' ')}
      type="button"
      onClick={onSelect}
    >
      <span className="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
        <span className="flex min-w-0 gap-3">
          <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
            <Icon aria-hidden="true" className="h-5 w-5" />
          </span>
          <span className="min-w-0">
            <span className="flex flex-wrap items-center gap-2">
              <span className="text-base font-semibold text-zinc-950 dark:text-zinc-50">
                {qualityText(option)}
              </span>
              <span className="rounded-md bg-zinc-100 px-2 py-0.5 text-xs font-bold text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
                {option.extension.toUpperCase()}
              </span>
              {recommended ? (
                <span className="rounded-md bg-emerald-100 px-2 py-0.5 text-xs font-bold text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-300">
                  Recomendado
                </span>
              ) : null}
            </span>
            <span className="mt-1 block text-sm text-zinc-600 dark:text-zinc-300">
              {optionDescription(option)}
            </span>
          </span>
        </span>

        <span className="grid gap-2 text-xs text-zinc-600 sm:grid-cols-3 md:min-w-80 dark:text-zinc-300">
          <Meta label="Resolução" value={option.resolution || option.quality || '-'} />
          <Meta label="Codecs" value={codecSummary(option)} />
          <Meta label="Tamanho" value={option.estimated_size ? formatBytes(option.estimated_size) : '-'} />
        </span>
      </span>

      <span className="mt-3 flex items-center justify-between gap-3 border-t border-zinc-100 pt-3 text-xs text-zinc-500 dark:border-zinc-800 dark:text-zinc-400">
        <span className="font-mono">{option.format_id}</span>
        {selected ? (
          <span className="inline-flex items-center gap-1.5 font-semibold text-emerald-700 dark:text-emerald-300">
            <CheckCircle2 aria-hidden="true" className="h-4 w-4" />
            Selecionado
          </span>
        ) : (
          <span className="inline-flex items-center gap-1.5 font-semibold text-zinc-600 dark:text-zinc-300">
            Escolher
            <ChevronRight aria-hidden="true" className="h-4 w-4" />
          </span>
        )}
      </span>
    </button>
  )
}

function SelectedSummary({ option }: { option: DownloadOption }) {
  return (
    <div className="rounded-lg border border-zinc-200 bg-zinc-50 px-4 py-3 dark:border-zinc-800 dark:bg-zinc-950">
      <p className="text-xs font-semibold uppercase text-zinc-500 dark:text-zinc-400">Seleção atual</p>
      <div className="mt-2 flex flex-wrap items-center gap-2">
        <span className="text-sm font-semibold text-zinc-950 dark:text-zinc-50">{qualityText(option)}</span>
        <span className="rounded-md bg-white px-2 py-0.5 text-xs font-bold text-zinc-700 dark:bg-zinc-900 dark:text-zinc-200">
          {option.extension.toUpperCase()}
        </span>
      </div>
      <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-300">{optionDescription(option)}</p>
      <p className="mt-2 inline-flex items-center gap-1.5 text-xs font-medium text-zinc-500 dark:text-zinc-400">
        <HardDrive aria-hidden="true" className="h-3.5 w-3.5" />
        {option.estimated_size ? formatBytes(option.estimated_size) : 'Tamanho estimado indisponível'}
      </p>
    </div>
  )
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <span className="rounded-md bg-zinc-50 px-2 py-1 dark:bg-zinc-900">
      <span className="block text-[11px] font-medium text-zinc-500 dark:text-zinc-400">{label}</span>
      <span className="mt-0.5 block truncate font-semibold text-zinc-800 dark:text-zinc-100">{value}</span>
    </span>
  )
}

function uniqueExtensions(options: DownloadOption[]) {
  return Array.from(new Set(options.map((option) => option.extension).filter(Boolean))).sort((left, right) => {
    return extensionPriority(right) - extensionPriority(left) || left.localeCompare(right)
  })
}

function extensionPriority(extension: string) {
  switch (extension) {
    case 'mp4':
    case 'm4a':
      return 4
    case 'webm':
      return 3
    case 'mp3':
      return 2
    default:
      return 1
  }
}

function qualityText(option: DownloadOption) {
  if (!option.has_video) {
    return option.extension ? `Áudio ${option.extension.toUpperCase()}` : 'Áudio'
  }

  return option.quality || option.resolution || option.label
}

function optionDescription(option: DownloadOption) {
  const type = option.has_video && option.has_audio
    ? 'Vídeo + áudio'
    : option.has_audio
      ? 'Somente áudio'
      : 'Somente vídeo'
  const codecs = codecSummary(option)

  return codecs === '-' ? type : `${type} · ${codecs}`
}

function codecSummary(option: DownloadOption) {
  const video = humanCodec(option.video_codec)
  const audio = humanCodec(option.audio_codec)

  if (video && audio) {
    return `${video} + ${audio}`
  }
  if (video) {
    return video
  }
  if (audio) {
    return audio
  }

  return '-'
}

function humanCodec(codec?: string) {
  const value = codec?.trim()
  if (!value) {
    return ''
  }

  const normalized = value.toLowerCase()
  if (normalized.startsWith('avc1') || normalized.includes('h264')) {
    return 'H264'
  }
  if (normalized.startsWith('av01') || normalized.includes('av1')) {
    return 'AV1'
  }
  if (normalized.includes('vp9')) {
    return 'VP9'
  }
  if (normalized.includes('vp8')) {
    return 'VP8'
  }
  if (normalized.startsWith('mp4a') || normalized.includes('aac')) {
    return 'AAC'
  }
  if (normalized.includes('opus')) {
    return 'Opus'
  }
  if (normalized.includes('vorbis')) {
    return 'Vorbis'
  }

  return value.toUpperCase()
}
