<script setup lang="ts">
import Drawer from 'primevue/drawer'
import AppLogo from '~/components/navigation/AppLogo.vue'
import AppSidebar from '~/components/navigation/AppSidebar.vue'
import DashboardTopbar from '~/components/navigation/DashboardTopbar.vue'

const DESKTOP_MEDIA_QUERY = '(min-width: 1024px)'

const route = useRoute()
const isMobileMenuOpen = ref(false)
const isDesktop = ref(false)
const isDesktopSidebarVisible = ref(true)

let desktopQuery: MediaQueryList | null = null

const pageTitle = computed(() => {
  return typeof route.meta.title === 'string' ? route.meta.title : ''
})

const syncDesktopState = () => {
  if (!desktopQuery) {
    return
  }

  isDesktop.value = desktopQuery.matches

  if (isDesktop.value) {
    isMobileMenuOpen.value = false
  }
}

const toggleMenu = () => {
  if (isDesktop.value) {
    isDesktopSidebarVisible.value = !isDesktopSidebarVisible.value
    return
  }

  isMobileMenuOpen.value = true
}

const closeMenu = () => {
  isMobileMenuOpen.value = false
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
      class="dashboard-sidebar dashboard-sidebar--desktop"
      :class="{ 'dashboard-sidebar--collapsed': !isDesktopSidebarVisible }"
    >
      <AppSidebar show-logo />
    </aside>

    <Drawer v-model:visible="isMobileMenuOpen" class="dashboard-drawer" position="left">
      <template #header>
        <AppLogo />
      </template>

      <AppSidebar @close="closeMenu" />
    </Drawer>

    <div class="dashboard-content">
      <DashboardTopbar :title="pageTitle" @menu-click="toggleMenu" />

      <main class="layout-main dashboard-content__main">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
.dashboard-sidebar {
  width: var(--app-sidebar-width);
  height: 100vh;
  border-right: 1px solid var(--app-color-border);
  background-color: var(--app-color-surface);
  overflow: hidden;
  transition:
    width var(--app-transition-fast),
    opacity var(--app-transition-fast),
    transform var(--app-transition-fast),
    border-color var(--app-transition-fast);
}

.dashboard-sidebar--desktop {
  display: none;
}

.dashboard-sidebar--collapsed {
  width: 0;
  opacity: 0;
  transform: translateX(calc(-1 * var(--app-sidebar-collapse-offset)));
  pointer-events: none;
  border-right-color: transparent;
}

.dashboard-drawer {
  width: var(--app-sidebar-width);
}

.dashboard-content {
  display: flex;
  min-width: 0;
  min-height: 0;
  flex: 1;
  flex-direction: column;
}

.dashboard-content__main {
  min-height: 0;
  overflow-y: auto;
  padding: 1.25rem;
}

@media (min-width: 1024px) {
  .dashboard-sidebar--desktop {
    display: block;
  }

  .dashboard-drawer {
    display: none;
  }

  .dashboard-content__main {
    padding: 2rem;
  }
}
</style>
