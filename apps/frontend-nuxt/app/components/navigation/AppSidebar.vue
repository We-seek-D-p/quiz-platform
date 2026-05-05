<script setup lang="ts">
import AppLogo from '~/components/navigation/AppLogo.vue'
import SidebarNavItem from '~/components/navigation/SidebarNavItem.vue'
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
  <aside class="app-sidebar">
    <div v-if="showLogo" class="app-sidebar__logo">
      <AppLogo />
    </div>

    <nav class="app-sidebar__nav">
      <SidebarNavItem
        v-for="item in dashboardNavigationItems"
        :key="item.key"
        :icon="item.icon"
        :label="item.label"
        :to="item.to"
        @click="handleNavigate"
      />
    </nav>

    <div class="app-sidebar__footer">
      <button type="button" class="app-sidebar__logout" @click="handleLogout">
        <i class="pi pi-sign-out app-sidebar__logout-icon" aria-hidden="true"></i>
        <span>Выйти</span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.app-sidebar {
  display: flex;
  height: 100%;
  min-height: 0;
  width: 100%;
  flex-direction: column;
  padding: 1rem 0.75rem;
}

.app-sidebar__logo {
  padding: 0.25rem 0.5rem 0.75rem;
}

.app-sidebar__nav {
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
  gap: 0.25rem;
  overflow-y: auto;
}

.app-sidebar__footer {
  margin-top: auto;
  padding-top: 1rem;
  border-top: 1px solid var(--app-color-border);
}

.app-sidebar__logout {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--app-color-border);
  border-radius: var(--app-control-radius);
  background-color: transparent;
  color: var(--app-color-text-muted);
  font: inherit;
  font-size: 0.875rem;
  font-weight: 500;
  justify-content: flex-start;
  cursor: pointer;
  transition: background-color var(--app-transition-fast), color var(--app-transition-fast);
}

.app-sidebar__logout:hover {
  background-color: var(--app-color-surface-hover);
  color: var(--app-color-text-strong);
}

.app-sidebar__logout-icon {
  font-size: 1rem;
}
</style>
