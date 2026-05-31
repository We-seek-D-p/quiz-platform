<script setup lang="ts">
import Button from 'primevue/button'
import AppLogo from '~/components/navigation/AppLogo.vue'
import { dashboardNavigationItems } from '~/constants/navigation'
import { useAuthStore } from '~/stores/auth'

withDefaults(
  defineProps<{
    showLogo?: boolean
  }>(),
  {
    showLogo: false,
  },
)

const emit = defineEmits<{
  close: []
}>()

const authStore = useAuthStore()
const route = useRoute()

const isActiveRoute = (to: string): boolean => {
  return route.path === to
}

const handleNavigate = () => {
  emit('close')
}

const handleLogout = async () => {
  await authStore.logout()
  emit('close')
  await navigateTo('/login')
}
</script>

<template>
  <aside class="flex h-full min-h-0 w-full flex-col px-3 py-4">
    <div v-if="showLogo" class="px-2 pt-1 pb-3">
      <AppLogo />
    </div>

    <nav class="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto">
      <Button
        v-for="item in dashboardNavigationItems"
        :key="item.key"
        as="router-link"
        :to="item.to"
        :icon="item.icon"
        :label="item.label"
        severity="secondary"
        :pt="{ label: { class: 'flex-1 text-left' } }"
        size="small"
        text
        class="sidebar-nav-button"
        :class="{ 'sidebar-nav-button--active': isActiveRoute(item.to) }"
        @click="handleNavigate"
      />
    </nav>

    <div class="mt-auto border-t border-(--app-color-border) pt-4">
      <Button
        label="Выйти"
        icon="pi pi-sign-out"
        severity="secondary"
        :pt="{ label: { class: 'flex-1 text-left' } }"
        size="small"
        text
        class="sidebar-nav-button"
        @click="handleLogout"
      />
    </div>
  </aside>
</template>

<style scoped>
.sidebar-nav-button {
  width: 100%;
  justify-content: flex-start !important;
  border: 1px solid transparent;
  color: var(--app-color-text-muted) !important;
}

.sidebar-nav-button:hover,
.sidebar-nav-button--active {
  background-color: var(--app-color-surface-hover) !important;
  color: var(--app-color-text-strong) !important;
}
</style>
