import { Copy, ExternalLink, QrCode } from 'lucide-react'
import QRCode from 'qrcode'
import { useEffect, useState } from 'react'

type NetworkQRCodeProps = {
  url: string
  onCopy: () => void
}

export function NetworkQRCode({ url, onCopy }: NetworkQRCodeProps) {
  const [imageURL, setImageURL] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false

    async function generateQRCode() {
      setError('')

      try {
        const generated = await QRCode.toDataURL(url, {
          color: {
            dark: '#18181b',
            light: '#ffffff',
          },
          errorCorrectionLevel: 'M',
          margin: 2,
          width: 224,
        })

        if (!cancelled) {
          setImageURL(generated)
        }
      } catch {
        if (!cancelled) {
          setImageURL('')
          setError('Não foi possível gerar o QR Code.')
        }
      }
    }

    if (url) {
      void generateQRCode()
    }

    return () => {
      cancelled = true
    }
  }, [url])

  return (
    <div className="rounded-lg border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-800 dark:bg-zinc-950">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-white text-zinc-700 shadow-sm dark:bg-zinc-900 dark:text-zinc-200">
          <QrCode aria-hidden="true" className="h-5 w-5" />
        </div>
        <div>
          <h4 className="text-sm font-semibold text-zinc-950 dark:text-zinc-50">QR Code</h4>
          <p className="text-sm text-zinc-600 dark:text-zinc-300">Acesso pela rede local.</p>
        </div>
      </div>

      <div className="mt-4 flex justify-center rounded-lg bg-white p-4 shadow-sm dark:bg-white">
        {imageURL ? (
          <img
            alt={`QR Code para ${url}`}
            className="h-56 w-56"
            height={224}
            src={imageURL}
            width={224}
          />
        ) : (
          <div className="flex h-56 w-56 items-center justify-center rounded-lg border border-dashed border-zinc-300 text-center text-sm text-zinc-500">
            {error || 'Gerando...'}
          </div>
        )}
      </div>

      <div className="mt-4 grid gap-2 sm:grid-cols-2">
        <button
          className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg border border-zinc-200 bg-white px-3 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-100 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800"
          type="button"
          onClick={onCopy}
        >
          <Copy aria-hidden="true" className="h-4 w-4" />
          Copiar
        </button>
        <a
          className="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg bg-zinc-950 px-3 text-sm font-semibold text-white transition hover:bg-zinc-800 dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
          href={url}
        >
          <ExternalLink aria-hidden="true" className="h-4 w-4" />
          Abrir
        </a>
      </div>
    </div>
  )
}
