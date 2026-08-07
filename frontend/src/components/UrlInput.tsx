import { Link2, Search } from 'lucide-react'

type UrlInputProps = {
  url: string
  loading: boolean
  onUrlChange: (url: string) => void
  onSubmit: () => void
}

export function UrlInput({ url, loading, onUrlChange, onSubmit }: UrlInputProps) {
  return (
    <form
      className="space-y-3"
      onSubmit={(event) => {
        event.preventDefault()
        onSubmit()
      }}
    >
      <label className="block text-sm font-semibold text-zinc-800 dark:text-zinc-100" htmlFor="video-url">
        Cole o link do vídeo
      </label>
      <div className="flex flex-col gap-3 lg:flex-row">
        <div className="relative flex-1">
          <Link2
            aria-hidden="true"
            className="pointer-events-none absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-zinc-400"
          />
          <input
            id="video-url"
            className="min-h-12 w-full rounded-lg border border-zinc-300 bg-white pl-10 pr-3 text-base text-zinc-950 outline-none transition placeholder:text-zinc-400 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-50 dark:focus:border-emerald-400 dark:focus:ring-emerald-400/20"
            placeholder="https://..."
            type="url"
            value={url}
            onChange={(event) => onUrlChange(event.target.value)}
          />
        </div>
        <button
          className="inline-flex min-h-12 items-center justify-center gap-2 rounded-lg bg-zinc-950 px-5 text-sm font-semibold text-white transition hover:bg-zinc-800 disabled:cursor-not-allowed disabled:bg-zinc-400 dark:bg-emerald-500 dark:text-zinc-950 dark:hover:bg-emerald-400"
          disabled={loading}
          type="submit"
        >
          <Search aria-hidden="true" className="h-4 w-4" />
          {loading ? 'Analisando...' : 'Analisar'}
        </button>
      </div>
    </form>
  )
}
