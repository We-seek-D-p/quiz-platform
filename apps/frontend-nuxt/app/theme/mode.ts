const THEME_STORAGE_KEY = 'quiz-theme-mode'
const DARK_CLASS = 'app-dark'
const DARK_QUERY = '(prefers-color-scheme: dark)'

const THEME_MODE_ORDER = ['system', 'dark', 'light'] as const

export type ThemeMode = (typeof THEME_MODE_ORDER)[number]
type ThemeModeSubscriber = (mode: ThemeMode) => void

let currentMode: ThemeMode = 'system'
let removeSystemListener: (() => void) | null = null
let removeStorageListener: (() => void) | null = null
const subscribers = new Set<ThemeModeSubscriber>()

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

const getStoredThemeMode = (): ThemeMode => {
  if (!isBrowser()) {
    return 'system'
  }

  const value = localStorage.getItem(THEME_STORAGE_KEY)
  return isThemeMode(value) ? value : 'system'
}

const notifySubscribers = () => {
  subscribers.forEach((subscriber) => {
    subscriber(currentMode)
  })
}

const setThemeMode = (mode: ThemeMode, persist: boolean) => {
  currentMode = mode
  if (persist) {
    saveThemeMode(mode)
  }
  applyDarkClass(resolveIsDark(mode))
  notifySubscribers()
}

export const getCurrentThemeMode = (): ThemeMode => {
  return currentMode
}

const getNextThemeMode = (mode: ThemeMode): ThemeMode => {
  const currentIndex = THEME_MODE_ORDER.indexOf(mode)
  const nextMode = THEME_MODE_ORDER[(currentIndex + 1) % THEME_MODE_ORDER.length]
  return nextMode ?? 'system'
}

export const subscribeThemeMode = (callback: ThemeModeSubscriber) => {
  subscribers.add(callback)
  callback(currentMode)

  return () => {
    subscribers.delete(callback)
  }
}

export const toggleThemeMode = (): ThemeMode => {
  const nextMode = getNextThemeMode(currentMode)
  setThemeMode(nextMode, true)
  return nextMode
}

export const initThemeMode = (): ThemeMode => {
  setThemeMode(getStoredThemeMode(), false)

  if (!isBrowser()) {
    return currentMode
  }

  removeSystemListener?.()
  removeStorageListener?.()

  const mediaQuery = getMediaQuery()
  if (mediaQuery) {
    const handleChange = () => {
      if (currentMode === 'system') {
        applyDarkClass(mediaQuery.matches)
        notifySubscribers()
      }
    }

    mediaQuery.addEventListener('change', handleChange)
    removeSystemListener = () => mediaQuery.removeEventListener('change', handleChange)
  }

  const handleStorageChange = (event: StorageEvent) => {
    if (event.key !== THEME_STORAGE_KEY) {
      return
    }

    const nextMode = isThemeMode(event.newValue) ? event.newValue : 'system'
    setThemeMode(nextMode, false)
  }

  window.addEventListener('storage', handleStorageChange)
  removeStorageListener = () => window.removeEventListener('storage', handleStorageChange)

  return currentMode
}
