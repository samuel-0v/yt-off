import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'

type Theme = 'light' | 'dark'
export type ThemePreference = Theme | 'system'

type ThemeContextValue = {
  resolvedTheme: Theme
  themePreference: ThemePreference
  setThemePreference: (theme: ThemePreference) => void
  toggleTheme: () => void
}

const STORAGE_KEY = 'yt-off-theme'
const ThemeContext = createContext<ThemeContextValue | undefined>(undefined)

type ThemeProviderProps = {
  children: ReactNode
}

export function ThemeProvider({ children }: ThemeProviderProps) {
  const [themePreference, setThemePreference] = useState<ThemePreference>(() => getInitialThemePreference())
  const [systemTheme, setSystemTheme] = useState<Theme>(() => getSystemTheme())
  const resolvedTheme = themePreference === 'system' ? systemTheme : themePreference

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = () => setSystemTheme(mediaQuery.matches ? 'dark' : 'light')

    mediaQuery.addEventListener('change', handleChange)
    return () => {
      mediaQuery.removeEventListener('change', handleChange)
    }
  }, [])

  useEffect(() => {
    document.documentElement.classList.toggle('dark', resolvedTheme === 'dark')
    window.localStorage.setItem(STORAGE_KEY, themePreference)
  }, [resolvedTheme, themePreference])

  const value = useMemo<ThemeContextValue>(
    () => ({
      resolvedTheme,
      themePreference,
      setThemePreference,
      toggleTheme: () => setThemePreference((currentTheme) => (currentTheme === 'dark' ? 'light' : 'dark')),
    }),
    [resolvedTheme, themePreference],
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  const value = useContext(ThemeContext)

  if (!value) {
    throw new Error('useTheme must be used inside ThemeProvider')
  }

  return value
}

function getInitialThemePreference(): ThemePreference {
  if (typeof window === 'undefined') {
    return 'system'
  }

  const storedTheme = window.localStorage.getItem(STORAGE_KEY)
  if (storedTheme === 'light' || storedTheme === 'dark' || storedTheme === 'system') {
    return storedTheme
  }

  return 'system'
}

function getSystemTheme(): Theme {
  if (typeof window === 'undefined') {
    return 'light'
  }

  if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return 'dark'
  }

  return 'light'
}
