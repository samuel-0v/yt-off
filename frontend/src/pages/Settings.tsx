import {
  BookOpen,
  Cookie,
  Database,
  FileCheck2,
  Globe2,
  HardDrive,
  Network,
  RefreshCw,
  RotateCcw,
  Save,
  Server,
  Settings2,
  ShieldCheck,
  Trash2,
  Upload,
} from 'lucide-react'
import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import {
  deleteCookies,
  getCookiesInfo,
  getNetworkInfo,
  getSettings,
  getStorage,
  getSystemInfo,
  getYTDLPVersion,
  saveCookies,
  updateSettings,
  uploadCookies,
} from '../api/client'
import { NetworkQRCode } from '../components/NetworkQRCode'
import { Skeleton } from '../components/Skeleton'
import { useTheme } from '../context/ThemeContext'
import { useToast } from '../context/ToastContext'
import type { AppSettings, CookiesInfo, NetworkInfo, StorageInfo, SystemInfo, YTDLPVersionInfo } from '../types/system'
import { formatBytes, formatDate } from '../utils/format'

const defaultSettings: AppSettings = {
  download_directory: '/downloads',
  max_concurrent_downloads: 2,
  language: 'pt-BR',
  theme: 'system',
  app_name: 'yt-off',
  backend_port: '18080',
  automatic_updates: false,
  show_hidden_files: false,
}

type SettingsData = {
  settings: AppSettings
  system: SystemInfo
  network: NetworkInfo
  storage: StorageInfo
  ytdlp: YTDLPVersionInfo
  cookies: CookiesInfo
}

export function Settings() {
  const [data, setData] = useState<SettingsData | null>(null)
  const [form, setForm] = useState<AppSettings>(defaultSettings)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [cookiesSaving, setCookiesSaving] = useState(false)
  const [cookiesDeleting, setCookiesDeleting] = useState(false)
  const [cookiesText, setCookiesText] = useState('')
  const [error, setError] = useState('')
  const { setThemePreference } = useTheme()
  const { showToast } = useToast()

  const backendPortChanged = useMemo(
    () => Boolean(data && form.backend_port !== data.settings.backend_port),
    [data, form.backend_port],
  )

  useEffect(() => {
    void loadSettings()
  }, [])

  async function loadSettings(showLoading = true) {
    if (showLoading) {
      setLoading(true)
    }
    setError('')

    try {
      const [settings, system, network, storage, ytdlp, cookies] = await Promise.all([
        getSettings(),
        getSystemInfo(),
        getNetworkInfo(),
        getStorage(),
        getYTDLPVersion().catch(() => ({ current: '' })),
        getCookiesInfo(),
      ])

      setData({ settings, system, network, storage, ytdlp, cookies })
      setForm(settings)
      setThemePreference(settings.theme)
    } catch {
      setError('Não foi possível carregar as configurações.')
      showToast({ type: 'error', title: 'Erro ao carregar configurações' })
    } finally {
      if (showLoading) {
        setLoading(false)
      }
    }
  }

  async function handleSave() {
    setSaving(true)
    setError('')

    try {
      const previousBackendPort = data?.settings.backend_port
      const updated = await updateSettings(form)
      setThemePreference(updated.theme)
      showToast({
        type: 'success',
        title: 'Configurações salvas',
        message:
          previousBackendPort && previousBackendPort !== updated.backend_port
            ? 'A porta do backend será aplicada em uma reinicialização futura.'
            : undefined,
      })
      await loadSettings(false)
    } catch {
      setError('Não foi possível salvar as configurações. Verifique os valores informados.')
      showToast({ type: 'error', title: 'Erro ao salvar configurações' })
    } finally {
      setSaving(false)
    }
  }

  async function handleRestoreDefaults() {
    setSaving(true)
    setError('')

    try {
      const updated = await updateSettings(defaultSettings)
      setThemePreference(updated.theme)
      showToast({ type: 'success', title: 'Padrões restaurados' })
      await loadSettings(false)
    } catch {
      setError('Não foi possível restaurar os padrões.')
      showToast({ type: 'error', title: 'Erro ao restaurar padrões' })
    } finally {
      setSaving(false)
    }
  }

  async function handleSaveCookies() {
    if (!cookiesText.trim()) {
      showToast({ type: 'error', title: 'Cole o conteúdo do cookies.txt' })
      return
    }

    setCookiesSaving(true)
    setError('')

    try {
      const cookies = await saveCookies(cookiesText)
      setCookiesText('')
      setData((current) => (current ? { ...current, cookies } : current))
      showToast({ type: 'success', title: 'Cookies salvos' })
    } catch {
      showToast({ type: 'error', title: 'Cookies inválidos', message: 'Use o formato Netscape cookies.txt.' })
    } finally {
      setCookiesSaving(false)
    }
  }

  async function handleUploadCookies(file: File | undefined) {
    if (!file) {
      return
    }

    setCookiesSaving(true)
    setError('')

    try {
      const cookies = await uploadCookies(file)
      setData((current) => (current ? { ...current, cookies } : current))
      showToast({ type: 'success', title: 'Arquivo de cookies enviado', message: file.name })
    } catch {
      showToast({ type: 'error', title: 'Arquivo de cookies inválido', message: 'Use o formato Netscape cookies.txt.' })
    } finally {
      setCookiesSaving(false)
    }
  }

  async function handleDeleteCookies() {
    setCookiesDeleting(true)

    try {
      await deleteCookies()
      const cookies = await getCookiesInfo()
      setData((current) => (current ? { ...current, cookies } : current))
      showToast({ type: 'info', title: 'Cookies removidos' })
    } catch {
      showToast({ type: 'error', title: 'Não foi possível remover os cookies' })
    } finally {
      setCookiesDeleting(false)
    }
  }

  async function handleCopyNetworkURL() {
    if (!data?.network.url) {
      return
    }

    try {
      await copyTextToClipboard(data.network.url)
      showToast({ type: 'success', title: 'Link copiado', message: data.network.url })
    } catch {
      showToast({ type: 'error', title: 'Não foi possível copiar o link' })
    }
  }

  return (
    <section className="mx-auto flex w-full max-w-7xl flex-col gap-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-emerald-700 dark:text-emerald-300">Preferências locais</p>
          <h2 className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-zinc-50">
            Configurações
          </h2>
          <p className="mt-2 max-w-2xl text-sm text-zinc-600 dark:text-zinc-300">
            Ajuste a aplicação, consulte o sistema e veja o endereço para uso na rede local.
          </p>
        </div>
        <button
          className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg border border-zinc-200 bg-white px-4 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-50 disabled:cursor-not-allowed dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800"
          disabled={loading || saving}
          type="button"
          onClick={() => void loadSettings()}
        >
          <RefreshCw aria-hidden="true" className={loading ? 'h-4 w-4 animate-spin' : 'h-4 w-4'} />
          Atualizar
        </button>
      </div>

      {error ? (
        <p className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200">
          {error}
        </p>
      ) : null}

      {loading ? (
        <div className="grid gap-4 lg:grid-cols-[minmax(0,1.1fr)_minmax(360px,0.9fr)]">
          <PanelSkeleton />
          <PanelSkeleton />
        </div>
      ) : (
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1.05fr)_minmax(380px,0.95fr)]">
          <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
                <Settings2 aria-hidden="true" className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-base font-semibold text-zinc-950 dark:text-zinc-50">Aplicação</h3>
                <p className="text-sm text-zinc-600 dark:text-zinc-300">Preferências salvas no SQLite.</p>
              </div>
            </div>

            <div className="mt-6 grid gap-4 md:grid-cols-2">
              <TextField
                label="Nome da aplicação"
                value={form.app_name}
                onChange={(value) => setForm((current) => ({ ...current, app_name: value }))}
              />
              <TextField
                label="Diretório de downloads"
                value={form.download_directory}
                onChange={(value) => setForm((current) => ({ ...current, download_directory: value }))}
              />
              <TextField
                label="Máximo de downloads simultâneos"
                min={1}
                max={10}
                type="number"
                value={form.max_concurrent_downloads.toString()}
                onChange={(value) =>
                  setForm((current) => ({
                    ...current,
                    max_concurrent_downloads: Number.parseInt(value, 10) || 1,
                  }))
                }
              />
              <TextField
                label="Porta do backend"
                type="number"
                value={form.backend_port}
                onChange={(value) => setForm((current) => ({ ...current, backend_port: value }))}
              />
              <SelectField
                label="Idioma"
                value={form.language}
                options={[
                  { label: 'Português (Brasil)', value: 'pt-BR' },
                  { label: 'English', value: 'en-US' },
                  { label: 'Español', value: 'es' },
                ]}
                onChange={(value) => setForm((current) => ({ ...current, language: value }))}
              />
              <SelectField
                label="Tema padrão"
                value={form.theme}
                options={[
                  { label: 'Sistema', value: 'system' },
                  { label: 'Claro', value: 'light' },
                  { label: 'Escuro', value: 'dark' },
                ]}
                onChange={(value) =>
                  setForm((current) => ({ ...current, theme: value as AppSettings['theme'] }))
                }
              />
            </div>

            <div className="mt-5 grid gap-3 md:grid-cols-2">
              <ToggleField
                checked={form.automatic_updates}
                description="Estrutura preparada para uma fase futura."
                label="Atualizações automáticas"
                onChange={(checked) => setForm((current) => ({ ...current, automatic_updates: checked }))}
              />
              <ToggleField
                checked={form.show_hidden_files}
                description="Inclui arquivos iniciados por ponto na lista."
                label="Mostrar arquivos ocultos"
                onChange={(checked) => setForm((current) => ({ ...current, show_hidden_files: checked }))}
              />
            </div>

            {backendPortChanged ? (
              <p className="mt-5 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-200">
                A porta do backend será aplicada em uma reinicialização futura.
              </p>
            ) : null}

            <div className="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
              <button
                className="inline-flex min-h-11 items-center justify-center gap-2 rounded-lg border border-zinc-300 px-4 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-100 disabled:cursor-not-allowed dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
                disabled={saving}
                type="button"
                onClick={() => void handleRestoreDefaults()}
              >
                <RotateCcw aria-hidden="true" className="h-4 w-4" />
                Restaurar padrão
              </button>
              <button
                className="inline-flex min-h-11 items-center justify-center gap-2 rounded-lg bg-emerald-600 px-4 text-sm font-semibold text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:bg-zinc-400 dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
                disabled={saving}
                type="button"
                onClick={() => void handleSave()}
              >
                <Save aria-hidden="true" className="h-4 w-4" />
                {saving ? 'Salvando...' : 'Salvar'}
              </button>
            </div>
          </section>

          <div className="space-y-4">
            <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
                  <Cookie aria-hidden="true" className="h-5 w-5" />
                </div>
                <div>
                  <h3 className="text-base font-semibold text-zinc-950 dark:text-zinc-50">Cookies do YouTube</h3>
                  <p className="text-sm text-zinc-600 dark:text-zinc-300">Usado quando o YouTube pede confirmação de navegador.</p>
                </div>
              </div>
              <Link
                className="mt-4 inline-flex min-h-10 w-full items-center justify-center gap-2 rounded-lg border border-zinc-200 bg-zinc-50 px-4 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-200 dark:hover:bg-zinc-800"
                to="/tutorial"
              >
                <BookOpen aria-hidden="true" className="h-4 w-4" />
                Como obter cookies.txt
              </Link>

              {data?.cookies ? (
                <div className="mt-5 space-y-3">
                  <div className="rounded-lg bg-zinc-50 p-3 dark:bg-zinc-950">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <div className="flex items-center gap-2">
                        <FileCheck2
                          aria-hidden="true"
                          className={data.cookies.exists && data.cookies.valid ? 'h-5 w-5 text-emerald-600 dark:text-emerald-300' : 'h-5 w-5 text-zinc-400'}
                        />
                        <div>
                          <p className="text-sm font-semibold text-zinc-950 dark:text-zinc-50">
                            {data.cookies.exists
                              ? data.cookies.valid
                                ? 'Cookies configurados'
                                : 'Cookies inválidos'
                              : 'Cookies não configurados'}
                          </p>
                          <p className="text-xs text-zinc-500 dark:text-zinc-400">
                            {data.cookies.exists
                              ? `${data.cookies.file_name} · ${formatBytes(data.cookies.size)}${data.cookies.updated_at ? ` · ${formatDate(data.cookies.updated_at)}` : ''}`
                              : 'Arquivo esperado: youtube.txt'}
                          </p>
                        </div>
                      </div>

                      {data.cookies.exists ? (
                        <button
                          className="inline-flex min-h-9 items-center justify-center gap-2 rounded-lg border border-red-200 px-3 text-sm font-semibold text-red-700 transition hover:bg-red-50 disabled:cursor-not-allowed dark:border-red-500/30 dark:text-red-300 dark:hover:bg-red-500/10"
                          disabled={cookiesDeleting}
                          type="button"
                          onClick={() => void handleDeleteCookies()}
                        >
                          <Trash2 aria-hidden="true" className="h-4 w-4" />
                          {cookiesDeleting ? 'Removendo...' : 'Remover'}
                        </button>
                      ) : null}
                    </div>
                  </div>

                  <label className="flex min-h-11 cursor-pointer items-center justify-center gap-2 rounded-lg border border-zinc-300 bg-white px-4 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-50 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-200 dark:hover:bg-zinc-800">
                    <Upload aria-hidden="true" className="h-4 w-4" />
                    Enviar arquivo cookies.txt
                    <input
                      accept=".txt,text/plain"
                      className="sr-only"
                      disabled={cookiesSaving}
                      type="file"
                      onChange={(event) => {
                        void handleUploadCookies(event.target.files?.[0])
                        event.currentTarget.value = ''
                      }}
                    />
                  </label>

                  <label className="block">
                    <span className="text-sm font-semibold text-zinc-800 dark:text-zinc-100">Ou cole o conteúdo do cookies.txt</span>
                    <textarea
                      className="mt-2 min-h-32 w-full resize-y rounded-lg border border-zinc-300 bg-white px-3 py-3 font-mono text-xs text-zinc-950 outline-none transition placeholder:font-sans placeholder:text-zinc-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-50 dark:focus:border-emerald-400 dark:focus:ring-emerald-400/20"
                      placeholder="# Netscape HTTP Cookie File"
                      spellCheck={false}
                      value={cookiesText}
                      onChange={(event) => setCookiesText(event.target.value)}
                    />
                  </label>

                  <button
                    className="inline-flex min-h-10 w-full items-center justify-center gap-2 rounded-lg bg-emerald-600 px-4 text-sm font-semibold text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:bg-zinc-400 dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
                    disabled={cookiesSaving || !cookiesText.trim()}
                    type="button"
                    onClick={() => void handleSaveCookies()}
                  >
                    <Save aria-hidden="true" className="h-4 w-4" />
                    {cookiesSaving ? 'Salvando...' : 'Salvar cookies'}
                  </button>
                </div>
              ) : null}
            </section>

            <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
                  <Network aria-hidden="true" className="h-5 w-5" />
                </div>
                <div>
                  <h3 className="text-base font-semibold text-zinc-950 dark:text-zinc-50">Rede Local</h3>
                  <p className="text-sm text-zinc-600 dark:text-zinc-300">Endereço sugerido para outros dispositivos.</p>
                </div>
              </div>
              {data?.network ? (
                <div className="mt-5 grid gap-4 lg:grid-cols-[minmax(0,1fr)_280px]">
                  <div className="space-y-3">
                    <InfoRow label="Hostname" value={data.network.hostname} />
                    <InfoRow label="IP local" value={data.network.local_ip} />
                    <InfoRow label="Frontend" value={data.network.frontend_port} />
                    <InfoRow label="Backend" value={data.network.backend_port} />
                    <a
                      className="block break-all rounded-lg bg-zinc-950 px-4 py-3 text-sm font-semibold text-white transition hover:bg-zinc-800 dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
                      href={data.network.url}
                    >
                      {data.network.url}
                    </a>
                  </div>
                  <NetworkQRCode url={data.network.url} onCopy={() => void handleCopyNetworkURL()} />
                </div>
              ) : null}
            </section>

            <section className="grid gap-4 sm:grid-cols-2">
              <SettingCard
                description={`${formatBytes(data?.storage.used)} usados de ${formatBytes(data?.storage.total)}.`}
                icon={<HardDrive aria-hidden="true" className="h-5 w-5" />}
                label="Espaço livre"
                value={formatBytes(data?.storage.free)}
              />
              <SettingCard
                description={`${data?.system.os}/${data?.system.architecture}`}
                icon={<ShieldCheck aria-hidden="true" className="h-5 w-5" />}
                label="Backend"
                value={data?.system.backend_version || '-'}
              />
              <SettingCard
                description="Status do Docker Engine."
                icon={<Server aria-hidden="true" className="h-5 w-5" />}
                label="Docker"
                tone={data?.system.docker === 'connected' ? 'success' : 'danger'}
                value={data?.system.docker || '-'}
              />
              <SettingCard
                description="Persistência de tarefas e preferências."
                icon={<Database aria-hidden="true" className="h-5 w-5" />}
                label="SQLite"
                tone={data?.system.sqlite === 'connected' ? 'success' : 'danger'}
                value={data?.system.sqlite || '-'}
              />
              <SettingCard
                description="Consulta de formatos e downloads."
                icon={<Globe2 aria-hidden="true" className="h-5 w-5" />}
                label="yt-dlp"
                value={data?.ytdlp.current || data?.system.yt_dlp_version || '-'}
              />
              <SettingCard
                description="Conversão e merge de vídeo com áudio."
                icon={<Settings2 aria-hidden="true" className="h-5 w-5" />}
                label="FFmpeg"
                value={data?.system.ffmpeg_version || '-'}
              />
            </section>
          </div>
        </div>
      )}
    </section>
  )
}

function TextField({
  label,
  max,
  min,
  type = 'text',
  value,
  onChange,
}: {
  label: string
  max?: number
  min?: number
  type?: 'text' | 'number'
  value: string
  onChange: (value: string) => void
}) {
  return (
    <label className="block">
      <span className="text-sm font-semibold text-zinc-800 dark:text-zinc-100">{label}</span>
      <input
        className="mt-2 min-h-11 w-full rounded-lg border border-zinc-300 bg-white px-3 text-sm text-zinc-950 outline-none transition focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-50 dark:focus:border-emerald-400 dark:focus:ring-emerald-400/20"
        max={max}
        min={min}
        type={type}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
    </label>
  )
}

function SelectField({
  label,
  options,
  value,
  onChange,
}: {
  label: string
  options: Array<{ label: string; value: string }>
  value: string
  onChange: (value: string) => void
}) {
  return (
    <label className="block">
      <span className="text-sm font-semibold text-zinc-800 dark:text-zinc-100">{label}</span>
      <select
        className="mt-2 min-h-11 w-full rounded-lg border border-zinc-300 bg-white px-3 text-sm font-semibold text-zinc-950 outline-none transition focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-50 dark:focus:border-emerald-400 dark:focus:ring-emerald-400/20"
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  )
}

function ToggleField({
  checked,
  description,
  label,
  onChange,
}: {
  checked: boolean
  description: string
  label: string
  onChange: (checked: boolean) => void
}) {
  return (
    <label className="flex cursor-pointer gap-3 rounded-lg border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-800 dark:bg-zinc-950">
      <input
        checked={checked}
        className="mt-1 h-4 w-4 rounded border-zinc-300 text-emerald-600 focus:ring-emerald-500"
        type="checkbox"
        onChange={(event) => onChange(event.target.checked)}
      />
      <span>
        <span className="block text-sm font-semibold text-zinc-900 dark:text-zinc-50">{label}</span>
        <span className="mt-1 block text-sm text-zinc-600 dark:text-zinc-300">{description}</span>
      </span>
    </label>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4 rounded-lg bg-zinc-50 px-3 py-2 text-sm dark:bg-zinc-950">
      <span className="text-zinc-500 dark:text-zinc-400">{label}</span>
      <span className="min-w-0 truncate font-semibold text-zinc-950 dark:text-zinc-50">{value}</span>
    </div>
  )
}

async function copyTextToClipboard(value: string) {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(value)
    return
  }

  const textArea = document.createElement('textarea')
  textArea.value = value
  textArea.setAttribute('readonly', 'true')
  textArea.style.position = 'fixed'
  textArea.style.left = '-9999px'
  textArea.style.top = '0'
  document.body.appendChild(textArea)
  textArea.focus()
  textArea.select()

  try {
    const copied = document.execCommand('copy')
    if (!copied) {
      throw new Error('copy failed')
    }
  } finally {
    document.body.removeChild(textArea)
  }
}

function SettingCard({
  description,
  icon,
  label,
  tone = 'neutral',
  value,
}: {
  description: string
  icon: ReactNode
  label: string
  tone?: 'neutral' | 'success' | 'danger'
  value: string
}) {
  return (
    <article className="rounded-lg border border-zinc-200 bg-white p-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
      <div className="flex items-center gap-3">
        <div
          className={[
            'flex h-10 w-10 shrink-0 items-center justify-center rounded-lg',
            tone === 'success'
              ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
              : '',
            tone === 'danger' ? 'bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-300' : '',
            tone === 'neutral' ? 'bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200' : '',
          ].join(' ')}
        >
          {icon}
        </div>
        <div className="min-w-0">
          <p className="text-sm font-semibold text-zinc-700 dark:text-zinc-200">{label}</p>
          <p className="mt-1 truncate text-lg font-semibold text-zinc-950 dark:text-zinc-50">{value}</p>
        </div>
      </div>
      <p className="mt-4 text-sm text-zinc-600 dark:text-zinc-300">{description}</p>
    </article>
  )
}

function PanelSkeleton() {
  return (
    <div className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
      <Skeleton className="h-5 w-1/3" />
      <Skeleton className="mt-5 h-11 w-full" />
      <Skeleton className="mt-3 h-11 w-full" />
      <Skeleton className="mt-3 h-20 w-full" />
    </div>
  )
}
