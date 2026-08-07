import { ArrowDown, CheckCircle2, Download, Film, ListChecks, Link2, PlayCircle } from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { ApiClientError, createDownload, getDownload, getFormats } from '../api/client'
import { DownloadCard, type DownloadCardItem } from '../components/DownloadCard'
import { EmptyState } from '../components/EmptyState'
import { FormatSelector } from '../components/FormatSelector'
import { Skeleton } from '../components/Skeleton'
import { UrlInput } from '../components/UrlInput'
import { useToast } from '../context/ToastContext'
import { useUser } from '../context/UserContext'
import type { DownloadOption, VideoInfo } from '../types/download'
import { formatBytes, formatDuration } from '../utils/format'

const steps = [
  { label: 'Cole o link', icon: Link2 },
  { label: 'Escolha a qualidade', icon: ListChecks },
  { label: 'Faça o download', icon: Download },
  { label: 'Acompanhe o progresso', icon: PlayCircle },
]

export function Home() {
  const [url, setUrl] = useState('')
  const [video, setVideo] = useState<VideoInfo | null>(null)
  const [selectedFormatId, setSelectedFormatId] = useState('')
  const [downloads, setDownloads] = useState<DownloadCardItem[]>([])
  const [analysisLoading, setAnalysisLoading] = useState(false)
  const [downloadLoading, setDownloadLoading] = useState(false)
  const [analysisError, setAnalysisError] = useState('')
  const [downloadError, setDownloadError] = useState('')
  const downloadsRef = useRef<DownloadCardItem[]>([])
  const statusRef = useRef<Record<string, DownloadCardItem['status']>>({})
  const { showToast } = useToast()
  const { activeUser } = useUser()

  const selectedOption = useMemo(
    () => video?.options.find((option) => option.format_id === selectedFormatId),
    [selectedFormatId, video],
  )
  const activeDownloadKey = downloads
    .filter((download) => download.status === 'queued' || download.status === 'running')
    .map((download) => `${download.id}:${download.status}`)
    .join('|')

  useEffect(() => {
    downloadsRef.current = downloads
  }, [downloads])

  useEffect(() => {
    if (!activeDownloadKey) {
      return
    }

    let cancelled = false

    const pollDownloads = async () => {
      const activeDownloads = downloadsRef.current.filter(
        (download) => download.status === 'queued' || download.status === 'running',
      )

      const updates = await Promise.all(
        activeDownloads.map(async (download) => {
          try {
            return await getDownload(download.id)
          } catch {
            return {
              ...download,
              status: 'failed' as const,
              error: 'Não foi possível consultar o progresso.',
            }
          }
        }),
      )

      if (cancelled) {
        return
      }

      updates.forEach((update) => {
        const previousStatus = statusRef.current[update.id]
        statusRef.current[update.id] = update.status

        if (previousStatus === update.status) {
          return
        }

        if (update.status === 'completed') {
          showToast({ type: 'success', title: 'Download concluído', message: update.filename })
        }

        if (update.status === 'cancelled') {
          showToast({ type: 'info', title: 'Download cancelado', message: update.filename })
        }

        if (update.status === 'failed') {
          showToast({ type: 'error', title: 'Erro no download', message: update.error })
        }
      })

      setDownloads((currentDownloads) =>
        currentDownloads.map((download) => {
          const update = updates.find((item) => item.id === download.id)

          if (!update) {
            return download
          }

          return {
            ...download,
            ...update,
          }
        }),
      )
    }

    void pollDownloads()
    const interval = window.setInterval(() => {
      void pollDownloads()
    }, 2000)

    return () => {
      cancelled = true
      window.clearInterval(interval)
    }
  }, [activeDownloadKey, showToast])

  async function handleAnalyze() {
    const trimmedURL = url.trim()

    setAnalysisError('')
    setDownloadError('')

    if (!trimmedURL) {
      setAnalysisError('Informe uma URL para analisar.')
      return
    }

    setAnalysisLoading(true)

    try {
      const info = await getFormats(trimmedURL)
      setVideo(info)
      setSelectedFormatId(info.options[0]?.format_id ?? '')
    } catch (error) {
      setVideo(null)
      setSelectedFormatId('')
      const message = formatAnalysisError(error)
      setAnalysisError(message)
      showToast({ type: 'error', title: 'Erro ao analisar vídeo', message })
    } finally {
      setAnalysisLoading(false)
    }
  }

  async function handleCreateDownload() {
    if (!video || !selectedOption) {
      setDownloadError('Escolha uma qualidade antes de iniciar.')
      return
    }

    setDownloadLoading(true)
    setDownloadError('')

    try {
      const response = await createDownload(url.trim(), selectedOption.format_id, selectedOption.extension, activeUser?.id)
      statusRef.current[response.id] = response.status

      setDownloads((currentDownloads) => [
        {
          id: response.id,
          user_id: activeUser?.id,
          owner_username: activeUser?.username,
          status: response.status,
          progress: 0,
          title: video.title,
          formatLabel: selectedOption.label,
        },
        ...currentDownloads,
      ])
      showToast({ type: 'success', title: 'Download iniciado', message: selectedOption.label })
    } catch {
      setDownloadError('Não foi possível iniciar o download.')
      showToast({ type: 'error', title: 'Erro ao iniciar download' })
    } finally {
      setDownloadLoading(false)
    }
  }

  return (
    <section className="mx-auto flex w-full max-w-7xl flex-col gap-6">
      <div className="flex flex-col gap-2">
        <p className="text-sm font-semibold text-emerald-700 dark:text-emerald-300">Fluxo local</p>
        <h2 className="text-2xl font-semibold text-zinc-950 dark:text-zinc-50">
          Baixe vídeos para a sua máquina
        </h2>
        <p className="max-w-2xl text-sm text-zinc-600 dark:text-zinc-300">
          Informe uma URL, escolha a qualidade e acompanhe o progresso direto no navegador.
        </p>
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_380px]">
        <div className="space-y-6">
          <div className="grid gap-3 md:grid-cols-4">
            {steps.map((step, index) => {
              const Icon = step.icon

              return (
                <div
                  className="rounded-lg border border-zinc-200 bg-white p-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
                  key={step.label}
                >
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
                      <Icon aria-hidden="true" className="h-5 w-5" />
                    </div>
                    <span className="text-sm font-semibold text-zinc-400">{index + 1}</span>
                  </div>
                  <p className="mt-4 text-sm font-semibold text-zinc-950 dark:text-zinc-50">{step.label}</p>
                </div>
              )
            })}
          </div>

          <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
            <UrlInput
              loading={analysisLoading}
              url={url}
              onSubmit={handleAnalyze}
              onUrlChange={setUrl}
            />

            {analysisLoading ? (
              <div className="mt-5 grid gap-4 sm:grid-cols-[220px_minmax(0,1fr)]">
                <Skeleton className="aspect-video h-auto w-full" />
                <div className="space-y-3">
                  <Skeleton className="h-5 w-2/3" />
                  <Skeleton className="h-4 w-1/3" />
                  <Skeleton className="h-4 w-1/2" />
                </div>
              </div>
            ) : null}

            {analysisError ? (
              <p className="mt-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200">
                {analysisError}
              </p>
            ) : null}
          </section>

          {video ? (
            <>
              <VideoSummary video={video} />

              <FormatSelector
                options={video.options}
                selectedFormatId={selectedFormatId}
                onSelect={setSelectedFormatId}
              />

              <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
                <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                  <div>
                    <h2 className="text-base font-semibold text-zinc-950 dark:text-zinc-50">Iniciar download</h2>
                    <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-300">
                      {selectedOption ? describeSelectedOption(selectedOption) : 'Selecione uma opção.'}
                    </p>
                  </div>
                  <button
                    className="inline-flex min-h-11 items-center justify-center gap-2 rounded-lg bg-emerald-600 px-5 text-sm font-semibold text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:bg-zinc-400 dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
                    disabled={!selectedOption || downloadLoading}
                    type="button"
                    onClick={handleCreateDownload}
                  >
                    <Download aria-hidden="true" className="h-4 w-4" />
                    {downloadLoading ? 'Iniciando...' : 'Baixar'}
                  </button>
                </div>

                {downloadError ? (
                  <p className="mt-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200">
                    {downloadError}
                  </p>
                ) : null}
              </section>
            </>
          ) : null}
        </div>

        <aside className="space-y-4">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h2 className="text-lg font-semibold text-zinc-950 dark:text-zinc-50">Downloads ativos</h2>
              <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-300">
                Atualização automática a cada 2 segundos.
              </p>
            </div>
            <ArrowDown aria-hidden="true" className="hidden h-5 w-5 text-zinc-400 xl:block" />
          </div>

          {downloads.length > 0 ? (
            <div className="space-y-3">
              {downloads.map((download) => (
                <DownloadCard download={download} key={download.id} />
              ))}
            </div>
          ) : (
            <EmptyState
              title="Nenhum download ativo"
              description="Os downloads iniciados nesta sessão aparecem aqui com progresso, velocidade e ETA."
            />
          )}
        </aside>
      </div>
    </section>
  )
}

function formatAnalysisError(error: unknown): string {
  if (error instanceof ApiClientError && error.message) {
    if (error.message.includes('youtube requires cookies')) {
      return 'O YouTube pediu confirmação. Salve os cookies exportados em cookies/youtube.txt e tente novamente.'
    }
    if (error.message.includes('rate limiting')) {
      return 'O YouTube está limitando as tentativas. Tente novamente mais tarde ou use cookies/youtube.txt.'
    }
  }

  return 'Não foi possível analisar o vídeo.'
}

function VideoSummary({ video }: { video: VideoInfo }) {
  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
      <div className="grid gap-5 lg:grid-cols-[minmax(260px,360px)_minmax(0,1fr)]">
        <div className="aspect-video overflow-hidden rounded-lg bg-zinc-100 dark:bg-zinc-800">
          {video.thumbnail ? (
            <img alt="" className="h-full w-full object-cover" src={video.thumbnail} />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-zinc-400">
              <Film aria-hidden="true" className="h-10 w-10" />
            </div>
          )}
        </div>
        <div className="min-w-0">
          <div className="flex items-center gap-2 text-sm font-semibold text-emerald-700 dark:text-emerald-300">
            <CheckCircle2 aria-hidden="true" className="h-4 w-4" />
            Vídeo analisado
          </div>
          <h2 className="mt-3 text-xl font-semibold text-zinc-950 dark:text-zinc-50">{video.title}</h2>
          {video.channel || video.uploader ? (
            <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">{video.channel || video.uploader}</p>
          ) : null}
          <dl className="mt-5 grid gap-3 sm:grid-cols-3">
            <VideoMetric label="Duração" value={formatDuration(video.duration)} />
            <VideoMetric label="Maior resolução" value={getMaxResolution(video.options)} />
            <VideoMetric label="Formatos" value={video.options.length.toString()} />
          </dl>
        </div>
      </div>
    </section>
  )
}

function VideoMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg bg-zinc-50 px-3 py-3 dark:bg-zinc-950">
      <dt className="text-xs font-medium text-zinc-500 dark:text-zinc-400">{label}</dt>
      <dd className="mt-1 truncate text-sm font-semibold text-zinc-950 dark:text-zinc-50">{value}</dd>
    </div>
  )
}

function describeSelectedOption(option: DownloadOption): string {
  const size = option.estimated_size ? `, ${formatBytes(option.estimated_size)}` : ''

  if (option.has_video && option.has_audio) {
    return `${option.label} selecionado para vídeo com áudio${size}.`
  }

  if (option.has_audio) {
    return `${option.label} selecionado para áudio${size}.`
  }

  return `${option.label} selecionado${size}.`
}

function getMaxResolution(options: DownloadOption[]): string {
  const heights = options
    .map((option) => option.resolution || option.quality)
    .map((value) => Number.parseInt(value.replace(/\D/g, ''), 10))
    .filter((value) => Number.isFinite(value) && value > 0)

  if (heights.length === 0) {
    return 'Auto'
  }

  return `${Math.max(...heights)}p`
}
