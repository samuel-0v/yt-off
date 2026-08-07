import {
  CheckSquare,
  Download,
  DownloadCloud,
  Filter,
  FolderPlus,
  Search,
  SlidersHorizontal,
  Square,
  User,
  Users,
  X,
} from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import {
  addGroupItem,
  cancelDownload,
  deleteFile,
  getFileDownloadURL,
  getFilePreviewURL,
  listGroups,
  listDownloads,
  listFiles,
} from '../api/client'
import { ConfirmDialog } from '../components/ConfirmDialog'
import { EmptyState } from '../components/EmptyState'
import { FileCard } from '../components/FileCard'
import { DownloadSkeleton } from '../components/Skeleton'
import { useToast } from '../context/ToastContext'
import { useUser } from '../context/UserContext'
import type { DownloadGroup, DownloadStatus, DownloadTask, FileInfo } from '../types/download'
import { formatBytes, statusLabel } from '../utils/format'

type DownloadFileItem = {
  id: string
  download?: DownloadTask
  file?: FileInfo
}

type FilterValue = 'all' | 'completed' | 'active' | 'cancelled' | 'failed'
type SortValue = 'date' | 'size' | 'name' | 'status'
type ScopeValue = 'mine' | 'all'

const filters: Array<{ label: string; value: FilterValue }> = [
  { label: 'Todos', value: 'all' },
  { label: 'Concluídos', value: 'completed' },
  { label: 'Em andamento', value: 'active' },
  { label: 'Cancelados', value: 'cancelled' },
  { label: 'Falharam', value: 'failed' },
]

export function Downloads() {
  const [downloads, setDownloads] = useState<DownloadTask[]>([])
  const [files, setFiles] = useState<FileInfo[]>([])
  const [groups, setGroups] = useState<DownloadGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [deletingName, setDeletingName] = useState('')
  const [cancellingID, setCancellingID] = useState('')
  const [addingToGroup, setAddingToGroup] = useState(false)
  const [pendingDeleteName, setPendingDeleteName] = useState('')
  const [groupTarget, setGroupTarget] = useState<DownloadTask | null>(null)
  const [selectedGroupID, setSelectedGroupID] = useState('')
  const [searchTerm, setSearchTerm] = useState('')
  const [filter, setFilter] = useState<FilterValue>('all')
  const [sort, setSort] = useState<SortValue>('date')
  const [scope, setScope] = useState<ScopeValue>('mine')
  const [selectedNames, setSelectedNames] = useState<string[]>([])
  const statusRef = useRef<Record<string, DownloadStatus>>({})
  const { showToast } = useToast()
  const { activeUser } = useUser()

  const items = useMemo<DownloadFileItem[]>(() => {
    const scopedFiles = scope === 'all'
      ? files
      : files.filter((file) => downloads.some((download) => download.filename === file.name))
    const filesByName = new Map(scopedFiles.map((file) => [file.name, file]))
    const namesFromDownloads = new Set<string>()

    const downloadItems = downloads.map((download) => {
      if (download.filename) {
        namesFromDownloads.add(download.filename)
      }

      return {
        id: download.id,
        download,
        file: download.filename ? filesByName.get(download.filename) : undefined,
      }
    })

    const fileItems = scopedFiles
      .filter((file) => !namesFromDownloads.has(file.name))
      .map((file) => ({
        id: `file:${file.name}`,
        file,
      }))

    return [...downloadItems, ...fileItems]
  }, [downloads, files, scope])

  const visibleItems = useMemo(() => {
    const normalizedSearch = searchTerm.trim().toLocaleLowerCase('pt-BR')

    return items
      .filter((item) => matchesFilter(item, filter))
      .filter((item) => {
        if (!normalizedSearch) {
          return true
        }

        const values = [
          getItemName(item),
          getItemExtension(item),
          item.download?.owner_username ?? '',
          item.download ? statusLabel(item.download.status) : 'arquivo',
        ]

        return values.some((value) => value.toLocaleLowerCase('pt-BR').includes(normalizedSearch))
      })
      .sort((first, second) => compareItems(first, second, sort))
  }, [filter, items, searchTerm, sort])

  const activeDownloadKey = downloads
    .filter((download) => download.status === 'queued' || download.status === 'running')
    .map((download) => `${download.id}:${download.status}`)
    .join('|')

  const visibleFiles = items.map((item) => item.file).filter((file): file is FileInfo => Boolean(file))
  const totalSize = visibleFiles.reduce((total, file) => total + file.size, 0)
  const activeCount = downloads.filter((download) => download.status === 'queued' || download.status === 'running').length
  const selectedNameSet = useMemo(() => new Set(selectedNames), [selectedNames])
  const downloadableVisibleNames = useMemo(
    () => visibleItems.map((item) => item.file?.name).filter((name): name is string => Boolean(name)),
    [visibleItems],
  )
  const allVisibleSelected = downloadableVisibleNames.length > 0
    && downloadableVisibleNames.every((name) => selectedNameSet.has(name))

  useEffect(() => {
    if (!activeUser) {
      return
    }

    void loadData()
  }, [activeUser?.id, scope])

  useEffect(() => {
    const availableNames = new Set(files.map((file) => file.name))
    setSelectedNames((current) => current.filter((name) => availableNames.has(name)))
  }, [files])

  useEffect(() => {
    if (!activeDownloadKey) {
      return
    }

    const interval = window.setInterval(() => {
      void loadData(false)
    }, 2000)

    return () => {
      window.clearInterval(interval)
    }
  }, [activeDownloadKey])

  async function loadData(showLoading = true) {
    if (!activeUser) {
      return
    }

    if (showLoading) {
      setLoading(true)
    }
    setError('')

    try {
      const [downloadItems, fileItems, groupItems] = await Promise.all([
        listDownloads(scope, activeUser.id),
        listFiles(),
        listGroups('mine', activeUser.id),
      ])
      notifyStatusChanges(downloadItems)
      setDownloads(downloadItems)
      setFiles(fileItems)
      setGroups(groupItems)
    } catch {
      setError('Não foi possível carregar downloads.')
      showToast({ type: 'error', title: 'Não foi possível carregar downloads' })
    } finally {
      if (showLoading) {
        setLoading(false)
      }
    }
  }

  function notifyStatusChanges(downloadItems: DownloadTask[]) {
    const hasPreviousState = Object.keys(statusRef.current).length > 0

    downloadItems.forEach((download) => {
      const previousStatus = statusRef.current[download.id]
      statusRef.current[download.id] = download.status

      if (!hasPreviousState || previousStatus === download.status) {
        return
      }

      if (download.status === 'completed') {
        showToast({ type: 'success', title: 'Download concluído', message: download.filename })
      }

      if (download.status === 'cancelled') {
        showToast({ type: 'info', title: 'Download cancelado', message: download.filename })
      }

      if (download.status === 'failed') {
        showToast({ type: 'error', title: 'Erro no download', message: download.error })
      }
    })
  }

  async function handleDelete(name: string) {
    setDeletingName(name)
    setError('')

    try {
      await deleteFile(name)
      setPendingDeleteName('')
      await loadData()
      showToast({ type: 'success', title: 'Arquivo excluído', message: name })
    } catch {
      setError('Não foi possível excluir arquivo.')
      showToast({ type: 'error', title: 'Erro ao excluir arquivo', message: name })
    } finally {
      setDeletingName('')
    }
  }

  async function handleCancel(id: string) {
    setCancellingID(id)
    setError('')

    try {
      await cancelDownload(id)
      await loadData(false)
      showToast({ type: 'info', title: 'Download cancelado' })
    } catch {
      setError('Não foi possível cancelar download.')
      showToast({ type: 'error', title: 'Não foi possível cancelar download' })
    } finally {
      setCancellingID('')
    }
  }

  function handleSelectionChange(name: string, selected: boolean) {
    setSelectedNames((current) => {
      if (selected) {
        return current.includes(name) ? current : [...current, name]
      }

      return current.filter((item) => item !== name)
    })
  }

  function handleToggleVisibleSelection() {
    if (allVisibleSelected) {
      setSelectedNames((current) => current.filter((name) => !downloadableVisibleNames.includes(name)))
      return
    }

    setSelectedNames((current) => Array.from(new Set([...current, ...downloadableVisibleNames])))
  }

  function handleDownloadSelected() {
    if (selectedNames.length === 0) {
      return
    }

    selectedNames.forEach((name) => {
      triggerFileDownload(getFileDownloadURL(name), name)
    })

    showToast({
      type: 'success',
      title: 'Downloads iniciados',
      message: `${selectedNames.length} arquivo${selectedNames.length === 1 ? '' : 's'} selecionado${selectedNames.length === 1 ? '' : 's'}.`,
    })
  }

  function handleOpenGroupDialog(download: DownloadTask) {
    setGroupTarget(download)
    setSelectedGroupID(groups[0]?.id ?? '')
  }

  async function handleAddToGroup() {
    if (!activeUser || !groupTarget || !selectedGroupID) {
      return
    }

    setAddingToGroup(true)
    setError('')
    try {
      await addGroupItem(selectedGroupID, activeUser.id, groupTarget.id)
      setGroupTarget(null)
      await loadData(false)
      showToast({ type: 'success', title: 'Download adicionado ao grupo' })
    } catch {
      setError('Não foi possível adicionar download ao grupo.')
      showToast({ type: 'error', title: 'Erro ao adicionar ao grupo' })
    } finally {
      setAddingToGroup(false)
    }
  }

  return (
    <section className="mx-auto flex w-full max-w-7xl flex-col gap-6">
      <div className="flex flex-col gap-2">
        <p className="text-sm font-semibold text-emerald-700 dark:text-emerald-300">Biblioteca local</p>
        <h2 className="text-2xl font-semibold text-zinc-950 dark:text-zinc-50">Histórico e arquivos</h2>
        <p className="max-w-2xl text-sm text-zinc-600 dark:text-zinc-300">
          Pesquise downloads, acompanhe itens em andamento e gerencie arquivos salvos em /downloads.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <SummaryCard label="Downloads" value={downloads.length.toString()} />
        <SummaryCard label="Em andamento" value={activeCount.toString()} />
        <SummaryCard label="Arquivos" value={`${files.length} · ${formatBytes(totalSize)}`} />
      </div>

      <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
        <div className="mb-4 flex flex-wrap gap-2">
          <button
            className={[
              'inline-flex min-h-10 items-center justify-center gap-2 rounded-lg px-3 text-sm font-semibold transition',
              scope === 'mine'
                ? 'bg-zinc-950 text-white dark:bg-emerald-400 dark:text-zinc-950'
                : 'bg-zinc-100 text-zinc-700 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-200 dark:hover:bg-zinc-700',
            ].join(' ')}
            type="button"
            onClick={() => setScope('mine')}
          >
            <User aria-hidden="true" className="h-4 w-4" />
            Meus
          </button>
          <button
            className={[
              'inline-flex min-h-10 items-center justify-center gap-2 rounded-lg px-3 text-sm font-semibold transition',
              scope === 'all'
                ? 'bg-zinc-950 text-white dark:bg-emerald-400 dark:text-zinc-950'
                : 'bg-zinc-100 text-zinc-700 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-200 dark:hover:bg-zinc-700',
            ].join(' ')}
            type="button"
            onClick={() => setScope('all')}
          >
            <Users aria-hidden="true" className="h-4 w-4" />
            Todos
          </button>
        </div>

        <div className="grid gap-3 xl:grid-cols-[minmax(260px,1fr)_auto_auto] xl:items-center">
          <label className="relative block">
            <span className="sr-only">Pesquisar downloads</span>
            <Search
              aria-hidden="true"
              className="pointer-events-none absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-zinc-400"
            />
            <input
              className="min-h-11 w-full rounded-lg border border-zinc-300 bg-white pl-10 pr-3 text-sm text-zinc-950 outline-none transition placeholder:text-zinc-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-50 dark:focus:border-emerald-400 dark:focus:ring-emerald-400/20"
              placeholder="Pesquisar..."
              type="search"
              value={searchTerm}
              onChange={(event) => setSearchTerm(event.target.value)}
            />
          </label>

          <div className="flex flex-wrap gap-2">
            {filters.map((item) => (
              <button
                className={[
                  'inline-flex min-h-10 items-center justify-center rounded-lg px-3 text-sm font-semibold transition',
                  item.value === filter
                    ? 'bg-zinc-950 text-white dark:bg-emerald-400 dark:text-zinc-950'
                    : 'bg-zinc-100 text-zinc-700 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-200 dark:hover:bg-zinc-700',
                ].join(' ')}
                key={item.value}
                type="button"
                onClick={() => setFilter(item.value)}
              >
                <Filter aria-hidden="true" className="mr-1.5 h-3.5 w-3.5" />
                {item.label}
              </button>
            ))}
          </div>

          <label className="flex min-h-11 items-center gap-2 rounded-lg border border-zinc-300 bg-white px-3 text-sm text-zinc-700 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-200">
            <SlidersHorizontal aria-hidden="true" className="h-4 w-4 text-zinc-400" />
            <span className="sr-only">Ordenar por</span>
            <select
              className="w-full bg-transparent font-semibold outline-none"
              value={sort}
              onChange={(event) => setSort(event.target.value as SortValue)}
            >
              <option value="date">Data</option>
              <option value="size">Tamanho</option>
              <option value="name">Nome</option>
              <option value="status">Status</option>
            </select>
          </label>
        </div>

        <div className="mt-4 flex flex-col gap-3 border-t border-zinc-100 pt-4 sm:flex-row sm:items-center sm:justify-between dark:border-zinc-800">
          <div className="flex flex-wrap gap-2">
            <button
              className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg border border-zinc-200 px-3 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-100 disabled:cursor-not-allowed disabled:opacity-50 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
              disabled={downloadableVisibleNames.length === 0}
              type="button"
              onClick={handleToggleVisibleSelection}
            >
              {allVisibleSelected ? (
                <CheckSquare aria-hidden="true" className="h-4 w-4" />
              ) : (
                <Square aria-hidden="true" className="h-4 w-4" />
              )}
              {allVisibleSelected ? 'Desmarcar visíveis' : 'Selecionar visíveis'}
            </button>
            {selectedNames.length > 0 ? (
              <button
                className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg border border-zinc-200 px-3 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-100 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
                type="button"
                onClick={() => setSelectedNames([])}
              >
                <X aria-hidden="true" className="h-4 w-4" />
                Limpar seleção
              </button>
            ) : null}
          </div>

          <button
            className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg bg-zinc-950 px-4 text-sm font-semibold text-white transition hover:bg-zinc-800 disabled:cursor-not-allowed disabled:bg-zinc-400 dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
            disabled={selectedNames.length === 0}
            type="button"
            onClick={handleDownloadSelected}
          >
            <Download aria-hidden="true" className="h-4 w-4" />
            Baixar selecionados
            {selectedNames.length > 0 ? (
              <span className="rounded-md bg-white/15 px-2 py-0.5 text-xs dark:bg-zinc-950/15">{selectedNames.length}</span>
            ) : null}
          </button>
        </div>
      </section>

      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, index) => (
            <DownloadSkeleton key={index} />
          ))}
        </div>
      ) : null}

      {error ? (
        <p className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200">
          {error}
        </p>
      ) : null}

      {!loading && items.length === 0 ? (
        <EmptyState
          title="Nenhum download encontrado"
          description="Quando um download for criado, ele aparecerá aqui com status e arquivo associado."
        />
      ) : null}

      {!loading && items.length > 0 && visibleItems.length === 0 ? (
        <EmptyState
          title="Nenhum resultado"
          description="Ajuste a busca, o filtro ou a ordenação para encontrar outro download."
        />
      ) : null}

      {!loading && visibleItems.length > 0 ? (
        <div className="space-y-3">
          {visibleItems.map((item) => (
            <FileCard
              cancelling={cancellingID === item.download?.id}
              deleting={deletingName === item.file?.name}
              download={item.download}
              downloadUrl={item.file ? getFileDownloadURL(item.file.name) : undefined}
              file={item.file}
              key={item.id}
              previewUrl={item.file ? getFilePreviewURL(item.file.name) : undefined}
              selectable={Boolean(item.file)}
              selected={Boolean(item.file && selectedNameSet.has(item.file.name))}
              showOwner={scope === 'all'}
              onAddToGroup={item.download ? handleOpenGroupDialog : undefined}
              onCancel={handleCancel}
              onDelete={setPendingDeleteName}
              onSelectionChange={handleSelectionChange}
            />
          ))}
        </div>
      ) : null}

      <ConfirmDialog
        confirmLabel="Excluir"
        description="Esta ação remove o arquivo do volume de downloads."
        loading={Boolean(deletingName)}
        open={Boolean(pendingDeleteName)}
        title="Deseja realmente excluir este arquivo?"
        onCancel={() => setPendingDeleteName('')}
        onConfirm={() => void handleDelete(pendingDeleteName)}
      />

      {groupTarget ? (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-zinc-950/50 px-4 py-8">
          <section className="w-full max-w-md rounded-lg border border-zinc-200 bg-white p-5 shadow-xl dark:border-zinc-800 dark:bg-zinc-900">
            <div className="flex items-start gap-3">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300">
                <FolderPlus aria-hidden="true" className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <h2 className="text-base font-semibold text-zinc-950 dark:text-zinc-50">Adicionar a grupo</h2>
                <p className="mt-1 truncate text-sm text-zinc-600 dark:text-zinc-300">
                  {groupTarget.filename ?? groupTarget.id}
                </p>
              </div>
            </div>

            {groups.length > 0 ? (
              <label className="mt-5 block">
                <span className="text-sm font-semibold text-zinc-700 dark:text-zinc-200">Grupo</span>
                <select
                  className="mt-2 min-h-11 w-full rounded-lg border border-zinc-300 bg-white px-3 text-sm text-zinc-950 outline-none transition focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-50 dark:focus:border-emerald-400 dark:focus:ring-emerald-400/20"
                  value={selectedGroupID}
                  onChange={(event) => setSelectedGroupID(event.target.value)}
                >
                  {groups.map((group) => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
              </label>
            ) : (
              <p className="mt-5 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-200">
                Crie um grupo na página Grupos antes de adicionar downloads.
              </p>
            )}

            <div className="mt-5 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <button
                className="inline-flex min-h-10 items-center justify-center rounded-lg border border-zinc-200 px-4 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-100 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
                type="button"
                onClick={() => setGroupTarget(null)}
              >
                Cancelar
              </button>
              <button
                className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg bg-emerald-600 px-4 text-sm font-semibold text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:bg-zinc-400 dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
                disabled={addingToGroup || !selectedGroupID}
                type="button"
                onClick={() => void handleAddToGroup()}
              >
                <FolderPlus aria-hidden="true" className="h-4 w-4" />
                {addingToGroup ? 'Adicionando...' : 'Adicionar'}
              </button>
            </div>
          </section>
        </div>
      ) : null}
    </section>
  )
}

function triggerFileDownload(url: string, name: string) {
  const link = document.createElement('a')
  link.href = url
  link.download = name
  link.rel = 'noreferrer'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

function SummaryCard({ label, value }: { label: string; value: string }) {
  return (
    <article className="rounded-lg border border-zinc-200 bg-white p-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
          <DownloadCloud aria-hidden="true" className="h-5 w-5" />
        </div>
        <div className="min-w-0">
          <p className="text-sm text-zinc-500 dark:text-zinc-400">{label}</p>
          <p className="truncate text-xl font-semibold text-zinc-950 dark:text-zinc-50">{value}</p>
        </div>
      </div>
    </article>
  )
}

function matchesFilter(item: DownloadFileItem, filter: FilterValue): boolean {
  const status = item.download?.status

  if (filter === 'all') {
    return true
  }

  if (filter === 'active') {
    return status === 'queued' || status === 'running'
  }

  if (!status && filter === 'completed') {
    return true
  }

  return status === filter
}

function compareItems(first: DownloadFileItem, second: DownloadFileItem, sort: SortValue): number {
  if (sort === 'name') {
    return getItemName(first).localeCompare(getItemName(second), 'pt-BR')
  }

  if (sort === 'size') {
    return getItemSize(second) - getItemSize(first)
  }

  if (sort === 'status') {
    return getItemStatus(first).localeCompare(getItemStatus(second), 'pt-BR')
  }

  return getItemTimestamp(second) - getItemTimestamp(first)
}

function getItemName(item: DownloadFileItem): string {
  return item.file?.name ?? item.download?.filename ?? item.download?.id ?? 'Arquivo'
}

function getItemExtension(item: DownloadFileItem): string {
  return item.file?.extension ?? item.download?.extension ?? ''
}

function getItemSize(item: DownloadFileItem): number {
  return item.file?.size ?? item.download?.file_size ?? 0
}

function getItemStatus(item: DownloadFileItem): string {
  return item.download?.status ?? 'completed'
}

function getItemTimestamp(item: DownloadFileItem): number {
  const value = item.download?.created_at ?? item.file?.modified_at
  const timestamp = value ? new Date(value).getTime() : 0

  return Number.isFinite(timestamp) ? timestamp : 0
}
