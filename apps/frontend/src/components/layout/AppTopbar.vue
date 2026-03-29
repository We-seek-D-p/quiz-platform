<script setup lang="ts">
import Button from 'primevue/button'
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue'
import AppLogo from '@/components/layout/AppLogo.vue'

withDefaults(
  defineProps<{
    title?: string
    showBrand?: boolean
    showMenuButton?: boolean
    showProfileButton?: boolean
  }>(),
  {
    title: '',
    showBrand: false,
    showMenuButton: false,
    showProfileButton: false,
  },
)

const emit = defineEmits<{
  menuClick: []
  profileClick: []
}>()
</script>

<template>
  <header
    class="flex h-14 items-center justify-between gap-4 px-4 border-b border-[var(--app-color-border)] bg-[var(--app-color-surface)] md:px-6"
  >
    <div class="flex min-w-0 items-center gap-3">
      <Button
        v-if="showMenuButton"
        icon="pi pi-bars"
        text
        class="inline-flex"
        @click="emit('menuClick')"
      />

      <AppLogo v-if="showBrand" />

      <p
        v-if="title"
        class="hidden truncate text-sm font-semibold text-[var(--app-color-text-muted)] lg:block"
      >
        {{ title }}
      </p>
    </div>

    <div class="flex items-center gap-3 lg:gap-4">
      <ThemeSwitcher />

      <Button
        v-if="showProfileButton"
        icon="pi pi-user"
        variant="outlined"
        class="h-11 w-11 rounded-full"
        @click="emit('profileClick')"
      />
    </div>
  </header>
</template>
