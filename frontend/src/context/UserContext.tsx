import { UserPlus } from 'lucide-react'
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type FormEvent,
  type ReactNode,
} from 'react'
import { createUser, listUsers } from '../api/client'
import type { User } from '../types/download'

type UserContextValue = {
  users: User[]
  activeUser: User | null
  loading: boolean
  selectUser: (userID: string) => void
  createAndSelectUser: (username: string) => Promise<User>
  refreshUsers: () => Promise<void>
}

const STORAGE_KEY = 'yt-off-active-user-id'
const UserContext = createContext<UserContextValue | undefined>(undefined)

type UserProviderProps = {
  children: ReactNode
}

export function UserProvider({ children }: UserProviderProps) {
  const [users, setUsers] = useState<User[]>([])
  const [activeUser, setActiveUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')

  async function refreshUsers() {
    const items = await listUsers()
    setUsers(items)
    setActiveUser((currentUser) => {
      const storedID = window.localStorage.getItem(STORAGE_KEY)
      const preferredID = currentUser?.id ?? storedID
      const preferredUser = items.find((user) => user.id === preferredID)
      if (preferredUser) {
        window.localStorage.setItem(STORAGE_KEY, preferredUser.id)
        return preferredUser
      }

      return null
    })
  }

  useEffect(() => {
    let cancelled = false

    async function load() {
      setLoading(true)
      setLoadError('')
      try {
        const items = await listUsers()
        if (cancelled) {
          return
        }

        const storedID = window.localStorage.getItem(STORAGE_KEY)
        setUsers(items)
        setActiveUser(items.find((user) => user.id === storedID) ?? null)
      } catch {
        if (!cancelled) {
          setLoadError('Não foi possível carregar usuários.')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void load()

    return () => {
      cancelled = true
    }
  }, [])

  const value = useMemo<UserContextValue>(
    () => ({
      users,
      activeUser,
      loading,
      selectUser: (userID: string) => {
        const selectedUser = users.find((user) => user.id === userID) ?? null
        setActiveUser(selectedUser)
        if (selectedUser) {
          window.localStorage.setItem(STORAGE_KEY, selectedUser.id)
        }
      },
      createAndSelectUser: async (username: string) => {
        const user = await createUser(username)
        const items = await listUsers()
        setUsers(items)
        const selectedUser = items.find((item) => item.id === user.id) ?? user
        setActiveUser(selectedUser)
        window.localStorage.setItem(STORAGE_KEY, selectedUser.id)
        return selectedUser
      },
      refreshUsers,
    }),
    [activeUser, loading, users],
  )

  return (
    <UserContext.Provider value={value}>
      {children}
      {!loading && !activeUser ? <UserGate error={loadError} users={users} /> : null}
    </UserContext.Provider>
  )
}

export function useUser() {
  const value = useContext(UserContext)

  if (!value) {
    throw new Error('useUser must be used inside UserProvider')
  }

  return value
}

function UserGate({ error, users }: { error: string; users: User[] }) {
  const { createAndSelectUser, selectUser } = useUser()
  const [username, setUsername] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState('')

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const nextUsername = username.trim()
    if (!nextUsername) {
      setFormError('Informe um username.')
      return
    }

    setSubmitting(true)
    setFormError('')
    try {
      await createAndSelectUser(nextUsername)
    } catch {
      setFormError('Não foi possível criar usuário.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-zinc-950/60 px-4 py-8 backdrop-blur">
      <section className="w-full max-w-md rounded-lg border border-zinc-200 bg-white p-5 shadow-xl dark:border-zinc-800 dark:bg-zinc-900">
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300">
            <UserPlus aria-hidden="true" className="h-5 w-5" />
          </div>
          <div>
            <h2 className="text-lg font-semibold text-zinc-950 dark:text-zinc-50">Escolha um usuário</h2>
            <p className="text-sm text-zinc-600 dark:text-zinc-300">Downloads e grupos serão organizados por perfil.</p>
          </div>
        </div>

        {users.length > 0 ? (
          <div className="mt-5 grid gap-2">
            {users.map((user) => (
              <button
                className="flex min-h-11 items-center justify-between rounded-lg border border-zinc-200 px-3 text-left text-sm font-semibold text-zinc-800 transition hover:bg-zinc-50 dark:border-zinc-700 dark:text-zinc-100 dark:hover:bg-zinc-800"
                key={user.id}
                type="button"
                onClick={() => selectUser(user.id)}
              >
                <span>{user.username}</span>
                <span className="text-xs text-zinc-500 dark:text-zinc-400">Entrar</span>
              </button>
            ))}
          </div>
        ) : null}

        <form className="mt-5 space-y-3" onSubmit={handleSubmit}>
          <label className="block">
            <span className="text-sm font-semibold text-zinc-700 dark:text-zinc-200">Novo username</span>
            <input
              className="mt-2 min-h-11 w-full rounded-lg border border-zinc-300 bg-white px-3 text-sm text-zinc-950 outline-none transition placeholder:text-zinc-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-50 dark:focus:border-emerald-400 dark:focus:ring-emerald-400/20"
              placeholder="ex: samuel"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
            />
          </label>

          {error || formError ? (
            <p className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200">
              {formError || error}
            </p>
          ) : null}

          <button
            className="inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg bg-emerald-600 px-4 text-sm font-semibold text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:bg-zinc-400 dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
            disabled={submitting}
            type="submit"
          >
            <UserPlus aria-hidden="true" className="h-4 w-4" />
            {submitting ? 'Criando...' : 'Criar e entrar'}
          </button>
        </form>
      </section>
    </div>
  )
}
