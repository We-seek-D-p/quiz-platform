import {
  getCurrentThemeMode,
  subscribeThemeMode,
  toggleThemeMode,
  type ThemeMode,
} from '~/theme/mode'

export const useThemeMode = () => {
  const mode = useState<ThemeMode>('theme-mode', () => getCurrentThemeMode())
  let removeThemeSubscription: (() => void) | null = null

  const syncFromRuntime = () => {
    mode.value = getCurrentThemeMode()
  }

  onMounted(() => {
    removeThemeSubscription = subscribeThemeMode((nextMode) => {
      mode.value = nextMode
    })
    syncFromRuntime()
  })

  onUnmounted(() => {
    removeThemeSubscription?.()
    removeThemeSubscription = null
  })

  const toggle = () => {
    mode.value = toggleThemeMode()
    return mode.value
  }

  return {
    mode,
    toggle,
    syncFromRuntime,
  }
}
