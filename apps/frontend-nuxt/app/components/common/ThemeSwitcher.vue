<script setup lang="ts">
import Button from 'primevue/button'
import { useThemeMode } from '~/composables/useThemeMode'

const { mode, toggle, syncFromRuntime } = useThemeMode()

const icon = computed(() => {
  if (mode.value === 'dark') {
    return 'pi pi-moon'
  }

  if (mode.value === 'light') {
    return 'pi pi-sun'
  }

  return 'pi pi-desktop'
})

onMounted(() => {
  syncFromRuntime()
})
</script>

<template>
  <Button
    :icon="icon"
    text
    severity="secondary"
    :aria-label="`Тема: ${mode}`"
    @click="toggle"
  />
</template>
