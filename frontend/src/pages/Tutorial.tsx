import {
  AlertTriangle,
  CheckCircle2,
  Cookie,
  Download,
  ExternalLink,
  FileText,
  Globe2,
  Puzzle,
  Settings,
  ShieldCheck,
  Upload,
} from 'lucide-react'
import { Link } from 'react-router-dom'

const steps = [
  {
    title: 'Instale a extensão',
    description: 'No navegador onde você usa o YouTube, instale a extensão Get cookies.txt LOCALLY.',
    icon: Puzzle,
  },
  {
    title: 'Entre no YouTube',
    description: 'Abra youtube.com e confirme que sua conta está logada antes de exportar.',
    icon: Globe2,
  },
  {
    title: 'Exporte os cookies',
    description: 'Clique na extensão e exporte os cookies do site atual em formato Netscape cookies.txt.',
    icon: Download,
  },
  {
    title: 'Importe no yt-off',
    description: 'Abra Configurações, envie o arquivo cookies.txt ou cole o conteúdo no campo de cookies.',
    icon: Upload,
  },
]

const checks = [
  'O arquivo deve começar com # Netscape HTTP Cookie File.',
  'Exporte cookies do YouTube usando a mesma conta que consegue assistir ao vídeo.',
  'Se o YouTube voltar a pedir confirmação, exporte um cookies.txt novo.',
  'Não use arquivo JSON ou CSV; o yt-off espera o formato Netscape.',
]

const extensionLinks = [
  {
    label: 'Chrome Web Store',
    url: 'https://chromewebstore.google.com/detail/get-cookiestxt-locally/cclelndahbckbenkjhflpdbgdldlbecc',
  },
  {
    label: 'Firefox Add-ons',
    url: 'https://addons.mozilla.org/firefox/addon/get-cookies-txt-locally/',
  },
  {
    label: 'Código no GitHub',
    url: 'https://github.com/kairi003/Get-cookies.txt-LOCALLY',
  },
]

export function Tutorial() {
  return (
    <section className="mx-auto flex w-full max-w-7xl flex-col gap-6">
      <div className="flex flex-col gap-2">
        <p className="text-sm font-semibold text-emerald-700 dark:text-emerald-300">Tutorial</p>
        <h2 className="text-2xl font-semibold text-zinc-950 dark:text-zinc-50">
          Como obter cookies do YouTube
        </h2>
        <p className="max-w-3xl text-sm text-zinc-600 dark:text-zinc-300">
          Use este passo a passo quando o YouTube mostrar confirmação de navegador, limite de acesso ou mensagem de bot.
        </p>
      </div>

      <section className="rounded-lg border border-amber-200 bg-amber-50 p-5 text-amber-950 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-100">
        <div className="flex gap-3">
          <AlertTriangle aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0" />
          <div>
            <h3 className="text-base font-semibold">Cookies são dados sensíveis</h3>
            <p className="mt-1 text-sm">
              O arquivo cookies.txt pode permitir acesso à sua sessão. Não envie esse arquivo para outras pessoas, não
              publique em prints e não cole em chats. Use apenas dentro do yt-off na sua máquina.
            </p>
          </div>
        </div>
      </section>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
        <section className="space-y-4">
          <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
                <Puzzle aria-hidden="true" className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-base font-semibold text-zinc-950 dark:text-zinc-50">Instalar extensão</h3>
                <p className="text-sm text-zinc-600 dark:text-zinc-300">Use os links oficiais do navegador.</p>
              </div>
            </div>

            <div className="mt-5 grid gap-3 sm:grid-cols-3">
              {extensionLinks.map((link) => (
                <a
                  className="inline-flex min-h-11 items-center justify-center gap-2 rounded-lg border border-zinc-200 bg-zinc-50 px-4 text-sm font-semibold text-zinc-700 transition hover:bg-zinc-100 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-200 dark:hover:bg-zinc-800"
                  href={link.url}
                  key={link.url}
                  rel="noreferrer"
                  target="_blank"
                >
                  <ExternalLink aria-hidden="true" className="h-4 w-4" />
                  {link.label}
                </a>
              ))}
            </div>
          </section>

          <div className="grid gap-4 sm:grid-cols-2">
            {steps.map((step, index) => {
              const Icon = step.icon

              return (
                <article
                  className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
                  key={step.title}
                >
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
                      <Icon aria-hidden="true" className="h-5 w-5" />
                    </div>
                    <span className="text-sm font-semibold text-zinc-400">{index + 1}</span>
                  </div>
                  <h3 className="mt-4 text-base font-semibold text-zinc-950 dark:text-zinc-50">{step.title}</h3>
                  <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">{step.description}</p>
                </article>
              )
            })}
          </div>

          <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
                <FileText aria-hidden="true" className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-base font-semibold text-zinc-950 dark:text-zinc-50">Passo a passo detalhado</h3>
                <p className="text-sm text-zinc-600 dark:text-zinc-300">Fluxo recomendado com Get cookies.txt LOCALLY.</p>
              </div>
            </div>

            <ol className="mt-5 space-y-3 text-sm text-zinc-700 dark:text-zinc-200">
              <li className="rounded-lg bg-zinc-50 p-3 dark:bg-zinc-950">
                Abra o YouTube no navegador e faça login na conta que você usa normalmente.
              </li>
              <li className="rounded-lg bg-zinc-50 p-3 dark:bg-zinc-950">
                Na página do YouTube, clique no ícone da extensão <strong>Get cookies.txt LOCALLY</strong>.
              </li>
              <li className="rounded-lg bg-zinc-50 p-3 dark:bg-zinc-950">
                Escolha a opção de exportar cookies do site atual, preferencialmente youtube.com, em formato
                <strong> Netscape cookies.txt</strong>.
              </li>
              <li className="rounded-lg bg-zinc-50 p-3 dark:bg-zinc-950">
                Salve o arquivo com o nome <code className="font-mono">cookies.txt</code> ou
                <code className="font-mono"> youtube.txt</code>.
              </li>
              <li className="rounded-lg bg-zinc-50 p-3 dark:bg-zinc-950">
                No yt-off, vá para Configurações e use <strong>Enviar arquivo cookies.txt</strong>.
              </li>
              <li className="rounded-lg bg-zinc-50 p-3 dark:bg-zinc-950">
                Volte para Home, analise o link novamente e escolha a qualidade do download.
              </li>
            </ol>
          </section>
        </section>

        <aside className="space-y-4">
          <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300">
                <ShieldCheck aria-hidden="true" className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-base font-semibold text-zinc-950 dark:text-zinc-50">Checklist</h3>
                <p className="text-sm text-zinc-600 dark:text-zinc-300">Antes de testar de novo.</p>
              </div>
            </div>

            <ul className="mt-5 space-y-3">
              {checks.map((item) => (
                <li className="flex gap-2 text-sm text-zinc-700 dark:text-zinc-200" key={item}>
                  <CheckCircle2 aria-hidden="true" className="mt-0.5 h-4 w-4 shrink-0 text-emerald-600 dark:text-emerald-300" />
                  <span>{item}</span>
                </li>
              ))}
            </ul>
          </section>

          <section className="rounded-lg border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200">
                <Cookie aria-hidden="true" className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-base font-semibold text-zinc-950 dark:text-zinc-50">Importar no yt-off</h3>
                <p className="text-sm text-zinc-600 dark:text-zinc-300">Abra a área de cookies.</p>
              </div>
            </div>

            <Link
              className="mt-5 inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg bg-emerald-600 px-4 text-sm font-semibold text-white transition hover:bg-emerald-700 dark:bg-emerald-400 dark:text-zinc-950 dark:hover:bg-emerald-300"
              to="/settings"
            >
              <Settings aria-hidden="true" className="h-4 w-4" />
              Ir para Configurações
            </Link>
          </section>
        </aside>
      </div>
    </section>
  )
}
