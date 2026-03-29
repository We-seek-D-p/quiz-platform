<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue'

const router = useRouter()
</script>

<template>
  <div class="flex flex-col min-h-screen bg-transparent">
    <header
      class="flex h-14 items-center justify-between px-4 border-b border-surface-200 bg-surface-0 dark:border-surface-800 dark:bg-surface-900"
    >
      <div
        class="text-xl text-primary font-bold cursor-pointer select-none"
        @click="router.push('/')"
      >
        Quiz Platform
      </div>

      <div class="flex items-center gap-3">
        <ThemeSwitcher />

        <Button
          icon="pi pi-user"
          variant="outlined"
          class="layout-icon-button"
          @click="router.push('/host')"
        />
      </div>
    </header>

    <main class="flex flex-1 items-center justify-center p-6">
      <slot />
    </main>
  </div>
</template>
