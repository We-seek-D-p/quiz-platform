<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Drawer from 'primevue/drawer'
import AppLogo from '@/components/layout/AppLogo.vue'
import AppSidebarContent from '@/components/layout/AppSidebarContent.vue'
import DashboardTopbar from '@/components/layout/DashboardTopbar.vue'
import { dashboardNavigationItems } from '@/layouts/dashboardNavigation'
import { useAuthStore } from '@/stores/auth'

const DESKTOP_MEDIA_QUERY = '(min-width: 1024px)'

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
  if (!desktopQuery) {
    return
  }

  isDesktop.value = desktopQuery.matches

  if (isDesktop.value) {
    isMobileMenuOpen.value = false
  }
}

const closeMobileMenu = () => {
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
    closeMobileMenu()
  }
}

const toggleSidebar = () => {
  if (isDesktop.value) {
    isDesktopSidebarVisible.value = !isDesktopSidebarVisible.value
    return
  }

  isMobileMenuOpen.value = true
}

onMounted(() => {
  desktopQuery = window.matchMedia(DESKTOP_MEDIA_QUERY)
  syncDesktopState()
  desktopQuery.addEventListener('change', syncDesktopState)
})

onBeforeUnmount(() => {
  desktopQuery?.removeEventListener('change', syncDesktopState)
  desktopQuery = null
})
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
        :menu-items="dashboardNavigationItems"
        @item-click="closeMobileMenu"
        @logout-click="handleLogoutClick"
      />
    </aside>

    <Drawer v-else v-model:visible="isMobileMenuOpen" class="dashboard-layout__drawer">
      <template #header>
        <AppLogo />
      </template>

      <AppSidebarContent
        :menu-items="dashboardNavigationItems"
        @item-click="closeMobileMenu"
        @logout-click="handleLogoutClick"
      />
    </Drawer>

    <div class="dashboard-layout__content">
      <DashboardTopbar :title="pageTitle" @menu-click="toggleSidebar" />

      <main class="dashboard-layout__main">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
.dashboard-layout__drawer {
  width: var(--app-sidebar-width);
}

.dashboard-layout__content {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
}

.dashboard-layout__main {
  flex: 1;
  overflow-y: auto;
  padding: 1.25rem;
}

@media (min-width: 768px) {
  .dashboard-layout__main {
    padding: 2rem;
  }
}
</style>
