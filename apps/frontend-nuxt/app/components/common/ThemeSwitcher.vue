<script setup lang="ts">
import Button from 'primevue/button'
import {
  getCurrentThemeMode,
  getStoredThemeMode,
  initThemeMode,
  toggleThemeMode,
  type ThemeMode,
} from '~/theme/mode'

const mode = ref<ThemeMode>('system')

const icon = computed(() => {
  if (mode.value === 'dark') {
    return 'pi pi-moon'
  }

  if (mode.value === 'light') {
    return 'pi pi-sun'
  }

  return 'pi pi-desktop'
})

const toggleTheme = () => {
  mode.value = toggleThemeMode()
}

onMounted(() => {
  mode.value = getStoredThemeMode()
  initThemeMode()
  mode.value = getCurrentThemeMode()
})
</script>

<template>
  <Button
    :icon="icon"
    text
    severity="secondary"
    :aria-label="`Тема: ${mode}`"
    @click="toggleTheme"
  />
</template>
