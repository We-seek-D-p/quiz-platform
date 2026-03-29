<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Button from 'primevue/button'
import { toggleThemeMode, getStoredThemeMode, type ThemeMode } from '@/theme'

const currentMode = ref<ThemeMode>('system')

onMounted(() => {
  currentMode.value = getStoredThemeMode()
})

const handleToggle = () => {
  currentMode.value = toggleThemeMode()
}

const getIcon = () => {
  if (currentMode.value === 'dark') return 'pi pi-moon'
  if (currentMode.value === 'light') return 'pi pi-sun'
  return 'pi pi-desktop'
}
</script>

<template>
  <Button
    :icon="getIcon()"
    severity="secondary"
    text
    rounded
    aria-label="Toggle Theme"
    :p-tooltip="`Тема: ${currentMode}`"
    tooltip-position="bottom"
    @click="handleToggle"
  />
</template>
