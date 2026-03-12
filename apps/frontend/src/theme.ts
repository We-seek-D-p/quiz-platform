const THEME_STORAGE_KEY = 'quiz-theme-mode'
const DARK_CLASS = 'app-dark'
const DARK_QUERY = '(prefers-color-scheme: dark)'

const THEME_MODE_ORDER = ['system', 'dark', 'light'] as const

export type ThemeMode = (typeof THEME_MODE_ORDER)[number]

let currentMode: ThemeMode = 'system'
let removeSystemListener: (() => void) | null = null

const isBrowser = () => typeof window !== 'undefined' && typeof document !== 'undefined'

const isThemeMode = (value: string | null): value is ThemeMode => {
  return THEME_MODE_ORDER.includes(value as ThemeMode)
}

const getMediaQuery = (): MediaQueryList | null => {
  if (!isBrowser() || typeof window.matchMedia !== 'function') {
    return null
  }

  return window.matchMedia(DARK_QUERY)
}

const getSystemPrefersDark = (): boolean => {
  return getMediaQuery()?.matches ?? false
}

const applyDarkClass = (isDark: boolean) => {
  if (!isBrowser()) {
    return
  }

  document.documentElement.classList.toggle(DARK_CLASS, isDark)
}

const resolveIsDark = (mode: ThemeMode): boolean => {
  if (mode === 'dark') {
    return true
  }

  if (mode === 'light') {
    return false
  }

  return getSystemPrefersDark()
}

const saveThemeMode = (mode: ThemeMode) => {
  if (!isBrowser()) {
    return
  }

  localStorage.setItem(THEME_STORAGE_KEY, mode)
}

export const getStoredThemeMode = (): ThemeMode => {
  if (!isBrowser()) {
    return 'system'
  }

  const value = localStorage.getItem(THEME_STORAGE_KEY)
  return isThemeMode(value) ? value : 'system'
}

export const setThemeMode = (mode: ThemeMode) => {
  currentMode = mode
  saveThemeMode(mode)
  applyDarkClass(resolveIsDark(mode))
}

export const getNextThemeMode = (mode: ThemeMode): ThemeMode => {
  const currentIndex = THEME_MODE_ORDER.indexOf(mode)
  const nextMode = THEME_MODE_ORDER[(currentIndex + 1) % THEME_MODE_ORDER.length]
  return nextMode ?? 'system'
}

export const toggleThemeMode = (): ThemeMode => {
  const nextMode = getNextThemeMode(currentMode)
  setThemeMode(nextMode)
  return nextMode
}

export const initThemeMode = (): ThemeMode => {
  currentMode = getStoredThemeMode()
  applyDarkClass(resolveIsDark(currentMode))

  if (!isBrowser()) {
    return currentMode
  }

  removeSystemListener?.()

  const mediaQuery = getMediaQuery()
  if (mediaQuery) {
    const handleChange = () => {
      if (currentMode === 'system') {
        applyDarkClass(mediaQuery.matches)
      }
    }

    mediaQuery.addEventListener('change', handleChange)
    removeSystemListener = () => mediaQuery.removeEventListener('change', handleChange)
  }

  return currentMode
}
