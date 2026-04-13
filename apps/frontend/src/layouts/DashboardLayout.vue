<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Drawer from 'primevue/drawer'
import AppLogo from '@/components/layout/AppLogo.vue'
import AppSidebarContent from '@/components/layout/AppSidebarContent.vue'
import AppTopbar from '@/components/layout/AppTopbar.vue'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isMobileMenuOpen = ref(false)
const isDesktop = ref(false)
const isDesktopSidebarVisible = ref(true)
const isLoggingOut = ref(false)

const pageTitle = computed(() => {
  return typeof route.meta.title === 'string' ? route.meta.title : ''
})

let desktopQuery: MediaQueryList | null = null

const syncDesktopState = () => {
  if (!desktopQuery) return

  isDesktop.value = desktopQuery.matches

  if (isDesktop.value) {
    isMobileMenuOpen.value = false
    isDesktopSidebarVisible.value = true
  }
}

onMounted(() => {
  desktopQuery = window.matchMedia('(min-width: 1024px)')
  syncDesktopState()
  desktopQuery.addEventListener('change', syncDesktopState)
})

onBeforeUnmount(() => {
  desktopQuery?.removeEventListener('change', syncDesktopState)
  desktopQuery = null
})

const menuItems = [
  { label: 'Дашборд', icon: 'pi pi-home' },
  { label: 'Создать квиз', icon: 'pi pi-plus' },
  { label: 'Мои квизы', icon: 'pi pi-list' },
  { label: 'Запуск квиза', icon: 'pi pi-send' },
]

const handleMenuClick = () => {
  if (!isDesktop.value) {
    isMobileMenuOpen.value = false
  }
}

const handleLogoutClick = async () => {
  if (isLoggingOut.value) {
    return
  }

  isLoggingOut.value = true

  try {
    await authStore.logout()
    await router.replace('/login')
  } finally {
    isLoggingOut.value = false
    handleMenuClick()
  }
}

const toggleSidebar = () => {
  if (isDesktop.value) {
    isDesktopSidebarVisible.value = !isDesktopSidebarVisible.value
    return
  }

  isMobileMenuOpen.value = true
}
</script>

<template>
  <div class="layout-shell layout-shell--row">
    <aside
      v-if="isDesktop"
      class="dashboard-sidebar"
      :class="{ 'dashboard-sidebar--collapsed': !isDesktopSidebarVisible }"
    >
      <AppSidebarContent
        show-logo
        :menu-items="menuItems"
        @item-click="handleMenuClick"
        @logout-click="handleLogoutClick"
      />
    </aside>

    <Drawer v-else v-model:visible="isMobileMenuOpen" class="w-64">
      <template #header>
        <AppLogo />
      </template>

      <AppSidebarContent
        :menu-items="menuItems"
        @item-click="handleMenuClick"
        @logout-click="handleLogoutClick"
      />
    </Drawer>

    <div class="layout-main flex min-w-0 flex-col">
      <AppTopbar :title="pageTitle" show-menu-button @menu-click="toggleSidebar" />

      <main class="flex-1 overflow-y-auto p-5 md:p-8">
        <slot />
      </main>
    </div>
  </div>
</template>
