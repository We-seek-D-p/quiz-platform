<script setup lang="ts">
import Button from 'primevue/button'
import { RouterLink } from 'vue-router'
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue'
import AppLogo from '@/components/layout/AppLogo.vue'
</script>

<template>
  <header class="public-topbar">
    <AppLogo />

    <div class="flex items-center gap-3">
      <ThemeSwitcher />

      <RouterLink to="/host" class="inline-flex" aria-label="Личный кабинет хоста">
        <Button icon="pi pi-user" variant="outlined" />
      </RouterLink>
    </div>
  </header>
</template>

<style scoped>
.public-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  height: var(--app-topbar-height);
  padding: 0 1rem;
  border-bottom: 1px solid var(--app-color-border);
  background-color: var(--app-color-surface);
}

@media (min-width: 768px) {
  .public-topbar {
    padding: 0 1.5rem;
  }
}
</style>
