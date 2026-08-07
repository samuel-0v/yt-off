import { AlertTriangle } from 'lucide-react'

type ConfirmDialogProps = {
  title: string
  description: string
  confirmLabel: string
  open: boolean
  loading?: boolean
  onCancel: () => void
  onConfirm: () => void
}

export function ConfirmDialog({
  title,
  description,
  confirmLabel,
  open,
  loading = false,
  onCancel,
  onConfirm,
}: ConfirmDialogProps) {
  if (!open) {
    return null
  }

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center bg-zinc-950/50 px-4 py-6">
      <div
        aria-modal="true"
        className="w-full max-w-md rounded-lg border border-zinc-200 bg-white p-5 shadow-xl dark:border-zinc-800 dark:bg-zinc-900"
        role="dialog"
      >
        <div className="flex gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-300">
            <AlertTriangle aria-hidden="true" className="h-5 w-5" />
          </div>
          <div>
            <h2 className="text-base font-semibold text-zinc-950 dark:text-zinc-50">{title}</h2>
            <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">{description}</p>
          </div>
        </div>

        <div className="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <button
            className="min-h-10 rounded-lg border border-zinc-300 px-4 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-100 disabled:cursor-not-allowed dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
            disabled={loading}
            type="button"
            onClick={onCancel}
          >
            Cancelar
          </button>
          <button
            className="min-h-10 rounded-lg bg-red-600 px-4 text-sm font-semibold text-white transition hover:bg-red-700 disabled:cursor-not-allowed disabled:bg-zinc-400"
            disabled={loading}
            type="button"
            onClick={onConfirm}
          >
            {loading ? 'Excluindo...' : confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
