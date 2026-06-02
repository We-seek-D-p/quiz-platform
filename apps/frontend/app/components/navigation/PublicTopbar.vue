<script setup lang="ts">
import Button from 'primevue/button'
import ThemeSwitcher from '~/components/common/ThemeSwitcher.vue'
import AppLogo from '~/components/navigation/AppLogo.vue'
import { useAuthStore } from '~/stores/auth'

const authStore = useAuthStore()

onMounted(async () => {
  await authStore.initializeSession()
})

const accountPath = computed(() => {
  return authStore.isAuthenticated ? '/host' : '/login'
})
</script>

<template>
  <header class="public-topbar">
    <AppLogo />

    <div class="public-topbar__actions">
      <ThemeSwitcher />

      <NuxtLink :to="accountPath" aria-label="Личный кабинет">
        <Button icon="pi pi-user" variant="outlined" />
      </NuxtLink>
    </div>
  </header>
</template>

<style scoped>
.public-topbar {
  display: flex;
  height: var(--app-topbar-height);
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0 var(--app-topbar-padding-x);
  border-bottom: 1px solid var(--app-color-border);
  background-color: var(--app-color-surface);
}

.public-topbar__actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

</style>
