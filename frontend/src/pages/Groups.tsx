import { Download, FolderOpen, FolderPlus, Plus, Trash2, User, Users, XCircle } from 'lucide-react'
import { useEffect, useState, type FormEvent } from 'react'
import {
  copyDownload,
  createGroup,
  deleteGroup,
  getGroup,
  listGroups,
  removeGroupItem,
} from '../api/client'
import { EmptyState } from '../components/EmptyState'
import { DownloadSkeleton } from '../components/Skeleton'
import { useToast } from '../context/ToastContext'
import { useUser } from '../context/UserContext'
import type { DownloadGroup, DownloadGroupItem, DownloadTask } from '../types/download'
import { formatDate, statusClassName, statusLabel } from '../utils/format'

type ScopeValue = 'mine' | 'all'

export function Groups() {
  const { activeUser } = useUser()
  const { showToast } = useToast()
  const [groups, setGroups] = useState<DownloadGroup[]>([])
  const [selectedGroup, setSelectedGroup] = useState<DownloadGroup | null>(null)
  const [scope, setScope] = useState<ScopeValue>('mine')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [copyingID, setCopyingID] = useState('')
  const [removingItemID, setRemovingItemID] = useState('')
  const [error, setError] = useState('')

  const ownsSelectedGroup = Boolean(activeUser && selectedGroup?.user_id === activeUser.id)

  useEffect(() => {
    if (!activeUser) {
      return
    }

    void loadGroups(true)
  }, [activeUser?.id, scope])

  async function loadGroups(showLoading = true, preferredGroupID = selectedGroup?.id) {
    if (!activeUser) {
      return
    }

    if (showLoading) {
      setLoading(true)
    }
    setError('')

    try {
      const items = await listGroups(scope, activeUser.id)
      setGroups(items)

      const nextGroupID = items.some((group) => group.id === preferredGroupID)
        ? preferredGroupID
        : items[0]?.id
      if (nextGroupID) {
        setSelectedGroup(await getGroup(nextGroupID))
      } else {
        setSelectedGroup(null)
      }
    } catch {
      setError('Não foi possível carregar grupos.')
      showToast({ type: 'error', title: 'Não foi possível carregar grupos' })
    } finally {
      if (showLoading) {
        setLoading(false)
      }
    }
  }

  async function handleSelectGroup(groupID: string) {
    setError('')
    try {
      setSelectedGroup(await getGroup(groupID))
    } catch {
      setError('Não foi possível abrir o grupo.')
      showToast({ type: 'error', title: 'Não foi possível abrir o grupo' })
    }
  }

  async function handleCreateGroup(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!activeUser) {
      return
    }

    const groupName = name.trim()
    if (!groupName) {
      setError('Informe o nome do grupo.')
      return
    }

    setCreating(true)
    setError('')
    try {
      const group = await createGroup(activeUser.id, groupName, description.trim())
      setName('')
      setDescription('')
      setScope('mine')
      await loadGroups(false, group.id)
      showToast({ type: 'success', title: 'Grupo criado', message: group.name })
    } catch {
      setError('Não foi possível criar grupo.')
      showToast({ type: 'error', title: 'Erro ao criar grupo' })
    } finally {
      setCreating(false)
    }
  }

  async function handleDeleteGroup() {
    if (!activeUser || !selectedGroup) {
      return
    }

    setDeleting(true)
    setError('')
    try {
      await deleteGroup(selectedGroup.id, activeUser.id)
      showToast({ type: 'success', title: 'Grupo excluído', message: selectedGroup.name })
      await loadGroups(false, '')
    } catch {
      setError('Não foi possível excluir grupo.')
      showToast({ type: 'error', title: 'Erro ao excluir grupo' })
    } finally {
      setDeleting(false)
    }
  }

  async function handleCopyDownload(download: DownloadTask) {
    if (!activeUser) {
      return
    }

    setCopyingID(download.id)
    setError('')
    try {
      await copyDownload(download.id, activeUser.id)
      showToast({ type: 'success', title: 'Download iniciado', message: download.filename })
    } catch {
      setError('Não foi possível iniciar download.')
      showToast({ type: 'error', title: 'Erro ao iniciar download' })
    } finally {
      setCopyingID('')
    }
  }

  async function handleRemoveItem(item: DownloadGroupItem) {
    if (!activeUser || !selectedGroup) {
      return
    }

    setRemovingItemID(item.id)
    setError('')
    try {
      await removeGroupItem(selectedGroup.id, item.id, activeUser.id)
      setSelectedGroup(await getGroup(selectedGroup.id))
      setGroups((currentGroups) =>
        currentGroups.map((group) =>
          group.id === selectedGroup.id
            ? { ...group, item_count: Math.max(0, group.item_count - 1) }
            : group,
        ),
      )
      showToast({ type: 'success', title: 'Item removido' })
    } catch {
      setError('Não foi possível remover item.')
      showToast({ type: 'error', title: 'Erro ao remover item' })
    } finally {
      setRemovingItemID('')
    }
  }

  return (
    <section className="mx-auto flex w-full max-w-7xl flex-col gap-6">
      <div className="flex flex-col gap-2">
        <p className="text-sm font-semibold text-emerald-700 dark:text-emerald-300">Playlists locais</p>
        <h2 className="text-2xl font-semibold text-zinc-950 dark:text-zinc-50">Grupos de downloads</h2>
        <p className="max-w-2xl text-sm text-zinc-600 dark:text-zinc-300">
          Organize downloads por perfil e abra grupos de outros usuários para baixar itens para você.
        </p>
      </div>

      <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
        <form className="grid gap-3 lg:grid-cols-[minmax(180px,1fr)_minmax(220px,1.4fr)_auto]" onSubmit={handleCreateGroup}>
          <label>
            <span className="text-sm font-semibold text-zinc-700 dark:text-zinc-200">Nome</span>
            <input
              className="mt-2 min-h-11 w-full rounded-lg border border-zinc-300 bg-white px-3 text-sm text-zinc-950 outline-none transition placeholder:text-zinc-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-50 dark:focus:border-emerald-400 dark:focus:ring-emerald-400/20"
              placeholder="Favoritos"
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
          </label>
          <label>
            <span className="text-sm font-semibold text-zinc-700 dark:text-zinc-200">Descrição</span>
            <input
              className="mt-2 min-h-11 w-full rounded-lg border border-zinc-300 bg-white px-3 text-sm text-zinc-950 outline-none transition placeholder:text-zinc-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-50 dark:focus:border-emerald-400 dark:focus:ring-emerald-400/20"
              placeholder="Opcional"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
            />
          </label>
          <button
            className="mt-7 inline-flex min-h-11 items-center justify-center gap-2 rounded-lg bg-emerald-600 px-4 text-sm font-semibold text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:bg-zinc-400 lg:mt-auto dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
            disabled={creating}
            type="submit"
          >
            <Plus aria-hidden="true" className="h-4 w-4" />
            {creating ? 'Criando...' : 'Criar grupo'}
          </button>
        </form>
      </section>

      <div className="flex flex-wrap gap-2">
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

      {error ? (
        <p className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200">
          {error}
        </p>
      ) : null}

      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, index) => (
            <DownloadSkeleton key={index} />
          ))}
        </div>
      ) : null}

      {!loading && groups.length === 0 ? (
        <EmptyState title="Nenhum grupo encontrado" description="Crie um grupo para começar a organizar downloads." />
      ) : null}

      {!loading && groups.length > 0 ? (
        <div className="grid gap-5 xl:grid-cols-[360px_minmax(0,1fr)]">
          <aside className="space-y-3">
            {groups.map((group) => (
              <button
                className={[
                  'w-full rounded-lg border p-4 text-left shadow-sm transition',
                  selectedGroup?.id === group.id
                    ? 'border-emerald-500 bg-emerald-50 dark:border-emerald-400 dark:bg-emerald-500/10'
                    : 'border-zinc-200 bg-white hover:bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900 dark:hover:bg-zinc-800',
                ].join(' ')}
                key={group.id}
                type="button"
                onClick={() => void handleSelectGroup(group.id)}
              >
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
                    <FolderOpen aria-hidden="true" className="h-5 w-5" />
                  </div>
                  <div className="min-w-0">
                    <h3 className="truncate text-sm font-semibold text-zinc-950 dark:text-zinc-50">{group.name}</h3>
                    <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-300">
                      {group.item_count} item{group.item_count === 1 ? '' : 's'}
                      {scope === 'all' && group.owner_username ? ` · ${group.owner_username}` : ''}
                    </p>
                  </div>
                </div>
              </button>
            ))}
          </aside>

          <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
            {selectedGroup ? (
              <>
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="min-w-0">
                    <h2 className="truncate text-xl font-semibold text-zinc-950 dark:text-zinc-50">{selectedGroup.name}</h2>
                    <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-sm text-zinc-600 dark:text-zinc-300">
                      {selectedGroup.owner_username ? <span>{selectedGroup.owner_username}</span> : null}
                      <span>{selectedGroup.item_count} item{selectedGroup.item_count === 1 ? '' : 's'}</span>
                      <span>{formatDate(selectedGroup.updated_at)}</span>
                    </div>
                    {selectedGroup.description ? (
                      <p className="mt-3 text-sm text-zinc-600 dark:text-zinc-300">{selectedGroup.description}</p>
                    ) : null}
                  </div>
                  {ownsSelectedGroup ? (
                    <button
                      className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg border border-red-200 px-3 text-sm font-semibold text-red-700 transition hover:bg-red-50 disabled:cursor-not-allowed disabled:border-zinc-200 disabled:text-zinc-400 dark:border-red-500/30 dark:text-red-300 dark:hover:bg-red-500/10"
                      disabled={deleting}
                      type="button"
                      onClick={() => void handleDeleteGroup()}
                    >
                      <Trash2 aria-hidden="true" className="h-4 w-4" />
                      {deleting ? 'Excluindo...' : 'Excluir'}
                    </button>
                  ) : null}
                </div>

                <div className="mt-5 space-y-3">
                  {selectedGroup.items && selectedGroup.items.length > 0 ? (
                    selectedGroup.items.map((item) => (
                      <GroupDownloadItem
                        copying={copyingID === item.download_id}
                        item={item}
                        key={item.id}
                        removing={removingItemID === item.id}
                        showRemove={ownsSelectedGroup}
                        onCopy={handleCopyDownload}
                        onRemove={handleRemoveItem}
                      />
                    ))
                  ) : (
                    <EmptyState title="Grupo vazio" description="Adicione downloads pela página Downloads." />
                  )}
                </div>
              </>
            ) : (
              <EmptyState title="Selecione um grupo" description="Os itens aparecem aqui." />
            )}
          </section>
        </div>
      ) : null}
    </section>
  )
}

function GroupDownloadItem({
  item,
  copying,
  removing,
  showRemove,
  onCopy,
  onRemove,
}: {
  item: DownloadGroupItem
  copying: boolean
  removing: boolean
  showRemove: boolean
  onCopy: (download: DownloadTask) => void
  onRemove: (item: DownloadGroupItem) => void
}) {
  const download = item.download
  const name = download?.filename ?? download?.id ?? item.download_id

  return (
    <article className="rounded-lg border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-800 dark:bg-zinc-950">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            {download ? <span className={statusClassName(download.status)}>{statusLabel(download.status)}</span> : null}
            {download?.owner_username ? (
              <span className="rounded-md bg-white px-2.5 py-1 text-xs font-semibold text-zinc-600 dark:bg-zinc-900 dark:text-zinc-300">
                {download.owner_username}
              </span>
            ) : null}
          </div>
          <h3 className="mt-3 truncate text-base font-semibold text-zinc-950 dark:text-zinc-50">{name}</h3>
          <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-sm text-zinc-600 dark:text-zinc-300">
            {download?.extension ? <span>{download.extension.toUpperCase()}</span> : null}
            {download?.created_at ? <span>{formatDate(download.created_at)}</span> : null}
          </div>
        </div>

        <div className="flex flex-wrap gap-2 lg:justify-end">
          <button
            className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg bg-zinc-950 px-3 text-sm font-semibold text-white transition hover:bg-zinc-800 disabled:cursor-not-allowed disabled:bg-zinc-400 dark:bg-emerald-500 dark:text-zinc-950 dark:hover:bg-emerald-400"
            disabled={!download || copying}
            type="button"
            onClick={() => download && onCopy(download)}
          >
            <Download aria-hidden="true" className="h-4 w-4" />
            {copying ? 'Iniciando...' : 'Baixar para mim'}
          </button>

          {showRemove ? (
            <button
              className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg border border-red-200 px-3 text-sm font-semibold text-red-700 transition hover:bg-red-50 disabled:cursor-not-allowed disabled:border-zinc-200 disabled:text-zinc-400 dark:border-red-500/30 dark:text-red-300 dark:hover:bg-red-500/10"
              disabled={removing}
              type="button"
              onClick={() => onRemove(item)}
            >
              <XCircle aria-hidden="true" className="h-4 w-4" />
              {removing ? 'Removendo...' : 'Remover'}
            </button>
          ) : null}
        </div>
      </div>
    </article>
  )
}
