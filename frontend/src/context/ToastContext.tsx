import { AlertCircle, CheckCircle2, Info, X } from 'lucide-react'
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react'

type ToastType = 'success' | 'error' | 'info'

type ToastInput = {
  type?: ToastType
  title: string
  message?: string
}

type Toast = ToastInput & {
  id: string
  type: ToastType
}

type ToastContextValue = {
  showToast: (toast: ToastInput) => void
}

const ToastContext = createContext<ToastContextValue | undefined>(undefined)

type ToastProviderProps = {
  children: ReactNode
}

export function ToastProvider({ children }: ToastProviderProps) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const removeToast = useCallback((id: string) => {
    setToasts((currentToasts) => currentToasts.filter((toast) => toast.id !== id))
  }, [])

  const showToast = useCallback(
    (toast: ToastInput) => {
      const id = createToastID()
      const nextToast: Toast = {
        id,
        type: toast.type ?? 'info',
        title: toast.title,
        message: toast.message,
      }

      setToasts((currentToasts) => [nextToast, ...currentToasts].slice(0, 4))
      window.setTimeout(() => removeToast(id), 4500)
    },
    [removeToast],
  )

  const value = useMemo<ToastContextValue>(() => ({ showToast }), [showToast])

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div
        aria-live="polite"
        className="pointer-events-none fixed right-4 top-4 z-50 flex w-[calc(100%-2rem)] max-w-sm flex-col gap-3 sm:right-6 sm:top-6"
      >
        {toasts.map((toast) => (
          <ToastItem key={toast.id} onClose={() => removeToast(toast.id)} toast={toast} />
        ))}
      </div>
    </ToastContext.Provider>
  )
}

function createToastID(): string {
  if (window.crypto.randomUUID) {
    return window.crypto.randomUUID()
  }

  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export function useToast() {
  const value = useContext(ToastContext)

  if (!value) {
    throw new Error('useToast must be used inside ToastProvider')
  }

  return value
}

function ToastItem({ toast, onClose }: { toast: Toast; onClose: () => void }) {
  const Icon = toast.type === 'success' ? CheckCircle2 : toast.type === 'error' ? AlertCircle : Info

  return (
    <div
      className={[
        'pointer-events-auto rounded-lg border bg-white p-4 shadow-lg shadow-zinc-900/10',
        'dark:border-zinc-800 dark:bg-zinc-900 dark:shadow-black/30',
        toast.type === 'success' ? 'border-emerald-200 dark:border-emerald-500/30' : '',
        toast.type === 'error' ? 'border-red-200 dark:border-red-500/30' : '',
        toast.type === 'info' ? 'border-sky-200 dark:border-sky-500/30' : '',
      ].join(' ')}
      role="status"
    >
      <div className="flex gap-3">
        <Icon
          aria-hidden="true"
          className={[
            'mt-0.5 h-5 w-5 shrink-0',
            toast.type === 'success' ? 'text-emerald-600 dark:text-emerald-300' : '',
            toast.type === 'error' ? 'text-red-600 dark:text-red-300' : '',
            toast.type === 'info' ? 'text-sky-600 dark:text-sky-300' : '',
          ].join(' ')}
        />
        <div className="min-w-0 flex-1">
          <p className="text-sm font-semibold text-zinc-950 dark:text-zinc-50">{toast.title}</p>
          {toast.message ? (
            <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-300">{toast.message}</p>
          ) : null}
        </div>
        <button
          aria-label="Fechar notificação"
          className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-zinc-500 transition hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
          type="button"
          onClick={onClose}
        >
          <X aria-hidden="true" className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}
