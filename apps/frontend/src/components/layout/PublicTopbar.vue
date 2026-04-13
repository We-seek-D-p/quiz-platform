<script setup lang="ts">
import Button from 'primevue/button'
import { RouterLink } from 'vue-router'
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue'
import AppLogo from '@/components/layout/AppLogo.vue'
</script>

<template>
  <header
    class="flex h-14 items-center justify-between gap-4 border-b border-[var(--app-color-border)] bg-[var(--app-color-surface)] px-4 md:px-6"
  >
    <div class="flex min-w-0 items-center">
      <AppLogo />
    </div>

    <div class="flex items-center gap-3 lg:gap-4">
      <ThemeSwitcher />

      <RouterLink to="/host" class="inline-flex">
        <Button label="ЛК хоста" icon="pi pi-user" variant="outlined" />
      </RouterLink>
    </div>
  </header>
</template>
