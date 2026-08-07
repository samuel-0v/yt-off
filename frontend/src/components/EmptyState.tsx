import type { ReactNode } from 'react'

type EmptyStateProps = {
  title: string
  description?: string
  action?: ReactNode
}

export function EmptyState({ title, description, action }: EmptyStateProps) {
  return (
    <div className="rounded-lg border border-dashed border-zinc-300 bg-white px-4 py-10 text-center dark:border-zinc-700 dark:bg-zinc-900">
      <p className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">{title}</p>
      {description ? <p className="mx-auto mt-2 max-w-md text-sm text-zinc-600 dark:text-zinc-300">{description}</p> : null}
      {action ? <div className="mt-5">{action}</div> : null}
    </div>
  )
}
