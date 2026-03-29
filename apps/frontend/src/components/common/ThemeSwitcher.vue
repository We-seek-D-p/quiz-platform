<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Button from 'primevue/button'
import { getStoredThemeMode, toggleThemeMode, type ThemeMode } from '@/theme'

const currentMode = ref<ThemeMode>('system')

onMounted(() => {
  currentMode.value = getStoredThemeMode()
})

const currentIcon = computed(() => {
  if (currentMode.value === 'dark') return 'pi pi-moon'
  if (currentMode.value === 'light') return 'pi pi-sun'
  return 'pi pi-desktop'
})

const handleToggle = () => {
  currentMode.value = toggleThemeMode()
}
</script>

<template>
  <Button
    :icon="currentIcon"
    text
    severity="secondary"
    :aria-label="`Theme: ${currentMode}`"
    @click="handleToggle"
  />
</template>
