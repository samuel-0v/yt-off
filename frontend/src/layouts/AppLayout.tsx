import { BookOpen, Download, FolderOpen, Home, Menu, Moon, Settings, Server, Sun, UserPlus, X } from 'lucide-react'
import { useEffect, useMemo, useState, type FormEvent, type ReactNode } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import { getHealth, getSettings, getVersion } from '../api/client'
import { useTheme } from '../context/ThemeContext'
import { useUser } from '../context/UserContext'

type AppLayoutProps = {
  children: ReactNode
}

const navItems = [
  { label: 'Home', to: '/', exact: true, icon: Home },
  { label: 'Downloads', to: '/downloads', exact: false, icon: Download },
  { label: 'Grupos', to: '/groups', exact: false, icon: FolderOpen },
  { label: 'Tutorial', to: '/tutorial', exact: false, icon: BookOpen },
  { label: 'Configurações', to: '/settings', exact: false, icon: Settings },
]

export function AppLayout({ children }: AppLayoutProps) {
  const location = useLocation()
  const { resolvedTheme, toggleTheme } = useTheme()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [backendOnline, setBackendOnline] = useState<boolean | null>(null)
  const [version, setVersion] = useState('0.1.0')
  const [appName, setAppName] = useState('yt-off')

  const pageTitle = useMemo(() => {
    if (location.pathname.startsWith('/downloads')) {
      return 'Downloads'
    }

    if (location.pathname.startsWith('/groups')) {
      return 'Grupos'
    }

    if (location.pathname.startsWith('/settings')) {
      return 'Configurações'
    }

    if (location.pathname.startsWith('/tutorial')) {
      return 'Tutorial'
    }

    return 'Home'
  }, [location.pathname])

  useEffect(() => {
    setMobileMenuOpen(false)
  }, [location.pathname])

  useEffect(() => {
    let cancelled = false

    async function refreshStatus() {
      try {
        const [health, info, settings] = await Promise.all([getHealth(), getVersion(), getSettings()])

        if (cancelled) {
          return
        }

        setBackendOnline(health.status === 'ok')
        setVersion(info.version)
        setAppName(settings.app_name || 'yt-off')
      } catch {
        if (!cancelled) {
          setBackendOnline(false)
        }
      }
    }

    void refreshStatus()
    const interval = window.setInterval(() => {
      void refreshStatus()
    }, 10000)

    return () => {
      cancelled = true
      window.clearInterval(interval)
    }
  }, [])

  return (
    <div className="min-h-screen bg-zinc-100 text-zinc-950 dark:bg-zinc-950 dark:text-zinc-50">
      <button
        aria-label="Abrir navegação"
        className="fixed left-4 top-4 z-30 inline-flex h-11 w-11 items-center justify-center rounded-lg border border-zinc-200 bg-white text-zinc-700 shadow-sm transition hover:bg-zinc-50 lg:hidden dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800"
        type="button"
        onClick={() => setMobileMenuOpen(true)}
      >
        <Menu aria-hidden="true" className="h-5 w-5" />
      </button>

      {mobileMenuOpen ? (
        <button
          aria-label="Fechar navegação"
          className="fixed inset-0 z-30 bg-zinc-950/50 lg:hidden"
          type="button"
          onClick={() => setMobileMenuOpen(false)}
        />
      ) : null}

      <aside
        className={[
          'fixed inset-y-0 left-0 z-40 flex w-72 flex-col border-r border-zinc-200 bg-white transition-transform duration-200 lg:translate-x-0 dark:border-zinc-800 dark:bg-zinc-900',
          mobileMenuOpen ? 'translate-x-0' : '-translate-x-full',
        ].join(' ')}
      >
        <div className="flex min-h-16 items-center justify-between border-b border-zinc-200 px-5 dark:border-zinc-800">
          <div>
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-zinc-950 text-sm font-black text-white dark:bg-emerald-400 dark:text-zinc-950">
                yt
              </div>
              <div>
                <p className="text-sm font-bold text-zinc-950 dark:text-zinc-50">{appName}</p>
                <p className="text-xs text-zinc-500 dark:text-zinc-400">v{version}</p>
              </div>
            </div>
          </div>
          <button
            aria-label="Fechar navegação"
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg text-zinc-500 transition hover:bg-zinc-100 lg:hidden dark:text-zinc-400 dark:hover:bg-zinc-800"
            type="button"
            onClick={() => setMobileMenuOpen(false)}
          >
            <X aria-hidden="true" className="h-5 w-5" />
          </button>
        </div>

        <nav className="flex-1 space-y-1 px-3 py-4">
          {navItems.map((item) => {
            const Icon = item.icon

            return (
              <NavLink
                activeClassName="bg-zinc-950 text-white dark:bg-emerald-400 dark:text-zinc-950"
                className="flex min-h-11 items-center gap-3 rounded-lg px-3 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-100 hover:text-zinc-950 dark:text-zinc-300 dark:hover:bg-zinc-800 dark:hover:text-zinc-50"
                exact={item.exact}
                key={item.to}
                to={item.to}
              >
                <Icon aria-hidden="true" className="h-5 w-5" />
                {item.label}
              </NavLink>
            )
          })}
        </nav>

        <div className="border-t border-zinc-200 p-4 dark:border-zinc-800">
          <UserSwitcher compact={false} />

          <div className="rounded-lg bg-zinc-50 p-3 dark:bg-zinc-950">
            <div className="flex items-center gap-2 text-sm font-semibold text-zinc-900 dark:text-zinc-50">
              <Server aria-hidden="true" className="h-4 w-4" />
              Backend
            </div>
            <div className="mt-2 flex items-center gap-2 text-sm text-zinc-600 dark:text-zinc-300">
              <span
                className={[
                  'h-2.5 w-2.5 rounded-full',
                  backendOnline ? 'bg-emerald-500' : backendOnline === false ? 'bg-red-500' : 'bg-zinc-400',
                ].join(' ')}
              />
              {backendOnline ? 'Online' : backendOnline === false ? 'Offline' : 'Verificando'}
            </div>
          </div>
        </div>
      </aside>

      <div className="min-h-screen lg:pl-72">
        <header className="sticky top-0 z-20 border-b border-zinc-200 bg-white/90 backdrop-blur dark:border-zinc-800 dark:bg-zinc-950/85">
          <div className="flex min-h-16 items-center justify-between gap-4 px-4 pl-20 sm:px-6 lg:px-8 lg:pl-8">
            <div>
              <p className="text-xs font-semibold uppercase text-zinc-500 dark:text-zinc-400">{appName}</p>
              <h1 className="text-xl font-semibold text-zinc-950 dark:text-zinc-50">{pageTitle}</h1>
            </div>
            <div className="flex items-center gap-2">
              <UserSwitcher compact />
              <div className="hidden items-center gap-2 rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-700 sm:flex dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200">
                <span
                  className={[
                    'h-2.5 w-2.5 rounded-full',
                    backendOnline ? 'bg-emerald-500' : backendOnline === false ? 'bg-red-500' : 'bg-zinc-400',
                  ].join(' ')}
                />
                {backendOnline ? 'Online' : backendOnline === false ? 'Offline' : 'Verificando'}
              </div>
              <button
                aria-label={resolvedTheme === 'dark' ? 'Usar tema claro' : 'Usar tema escuro'}
                className="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-zinc-200 bg-white text-zinc-700 transition hover:bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800"
                type="button"
                onClick={toggleTheme}
              >
                {resolvedTheme === 'dark' ? (
                  <Sun aria-hidden="true" className="h-5 w-5" />
                ) : (
                  <Moon aria-hidden="true" className="h-5 w-5" />
                )}
              </button>
            </div>
          </div>
        </header>

        <main className="px-4 py-6 sm:px-6 lg:px-8">{children}</main>
      </div>
    </div>
  )
}

function UserSwitcher({ compact }: { compact: boolean }) {
  const { activeUser, users, selectUser, createAndSelectUser } = useUser()
  const [creating, setCreating] = useState(false)
  const [username, setUsername] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  async function handleCreateUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const nextUsername = username.trim()
    if (!nextUsername) {
      setError('Informe um username.')
      return
    }

    setSaving(true)
    setError('')
    try {
      await createAndSelectUser(nextUsername)
      setUsername('')
      setCreating(false)
    } catch {
      setError('Não foi possível criar usuário.')
    } finally {
      setSaving(false)
    }
  }

  if (compact) {
    return (
      <div className="hidden items-center gap-2 md:flex">
        <label className="flex min-h-10 items-center gap-2 rounded-lg border border-zinc-200 bg-white px-3 text-sm text-zinc-700 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200">
          <span className="sr-only">Usuário ativo</span>
          <select
            className="max-w-36 bg-transparent font-semibold outline-none"
            value={activeUser?.id ?? ''}
            onChange={(event) => selectUser(event.target.value)}
          >
            {users.map((user) => (
              <option key={user.id} value={user.id}>
                {user.username}
              </option>
            ))}
          </select>
        </label>
        <button
          aria-label="Criar usuário"
          className="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-zinc-200 bg-white text-zinc-700 transition hover:bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800"
          type="button"
          onClick={() => setCreating((current) => !current)}
        >
          <UserPlus aria-hidden="true" className="h-4 w-4" />
        </button>
        {creating ? (
          <form
            className="absolute right-24 top-14 z-30 w-72 rounded-lg border border-zinc-200 bg-white p-3 shadow-lg dark:border-zinc-800 dark:bg-zinc-900"
            onSubmit={handleCreateUser}
          >
            <input
              className="min-h-10 w-full rounded-lg border border-zinc-300 bg-white px-3 text-sm text-zinc-950 outline-none transition placeholder:text-zinc-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-50 dark:focus:border-emerald-400 dark:focus:ring-emerald-400/20"
              placeholder="Novo username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
            />
            {error ? <p className="mt-2 text-sm text-red-600 dark:text-red-300">{error}</p> : null}
            <button
              className="mt-2 inline-flex min-h-10 w-full items-center justify-center gap-2 rounded-lg bg-emerald-600 px-3 text-sm font-semibold text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:bg-zinc-400 dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
              disabled={saving}
              type="submit"
            >
              <UserPlus aria-hidden="true" className="h-4 w-4" />
              {saving ? 'Criando...' : 'Criar'}
            </button>
          </form>
        ) : null}
      </div>
    )
  }

  return (
    <div className="mb-3 rounded-lg bg-zinc-50 p-3 dark:bg-zinc-950">
      <div className="flex items-center justify-between gap-2">
        <p className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">Usuário</p>
        <button
          aria-label="Criar usuário"
          className="inline-flex h-8 w-8 items-center justify-center rounded-md text-zinc-500 transition hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
          type="button"
          onClick={() => setCreating((current) => !current)}
        >
          <UserPlus aria-hidden="true" className="h-4 w-4" />
        </button>
      </div>
      <select
        className="mt-2 min-h-10 w-full rounded-lg border border-zinc-200 bg-white px-3 text-sm font-semibold text-zinc-800 outline-none dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100"
        value={activeUser?.id ?? ''}
        onChange={(event) => selectUser(event.target.value)}
      >
        {users.map((user) => (
          <option key={user.id} value={user.id}>
            {user.username}
          </option>
        ))}
      </select>
      {creating ? (
        <form className="mt-3 space-y-2" onSubmit={handleCreateUser}>
          <input
            className="min-h-10 w-full rounded-lg border border-zinc-300 bg-white px-3 text-sm text-zinc-950 outline-none transition placeholder:text-zinc-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-50 dark:focus:border-emerald-400 dark:focus:ring-emerald-400/20"
            placeholder="Novo username"
            value={username}
            onChange={(event) => setUsername(event.target.value)}
          />
          {error ? <p className="text-sm text-red-600 dark:text-red-300">{error}</p> : null}
          <button
            className="inline-flex min-h-10 w-full items-center justify-center gap-2 rounded-lg bg-emerald-600 px-3 text-sm font-semibold text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:bg-zinc-400 dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
            disabled={saving}
            type="submit"
          >
            <UserPlus aria-hidden="true" className="h-4 w-4" />
            {saving ? 'Criando...' : 'Criar'}
          </button>
        </form>
      ) : null}
    </div>
  )
}
