<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import Drawer from 'primevue/drawer'
import Button from 'primevue/button'
import { useRouter } from 'vue-router'
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue'
import AppMenuItem from '@/components/layout/AppMenuItem.vue'

const router = useRouter()
const isMobileMenuOpen = ref(false)
const userEmail = ref('host@example.com')
const isDesktop = ref(false)
const isDesktopSidebarVisible = ref(true)

let desktopQuery: MediaQueryList | null = null

const syncDesktopState = () => {
  if (!desktopQuery) {
    return
  }
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

const menuItems = ref([
  { label: 'Дашборд', icon: 'pi pi-home' },
  { label: 'Создать квиз', icon: 'pi pi-plus' },
  { label: 'Мои квизы', icon: 'pi pi-list' },
  { label: 'Запуск квиза', icon: 'pi pi-send' },
])

const handleMenuClick = () => {
  if (!isDesktop.value) {
    isMobileMenuOpen.value = false
  }
}

const goHome = () => {
  router.push('/')
  handleMenuClick()
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
  <div class="min-h-screen flex bg-transparent">
    <aside
      class="hidden lg:flex flex-col border-r border-surface-200 dark:border-surface-800 bg-surface-0 dark:bg-surface-900 overflow-hidden transition-[width,opacity,transform] duration-200 ease-in-out"
      :class="
        isDesktopSidebarVisible
          ? 'w-64 opacity-100 translate-x-0'
          : 'w-0 opacity-0 -translate-x-3 pointer-events-none border-r-0'
      "
    >
      <button
        type="button"
        class="px-5 py-4 text-left text-lg font-semibold text-primary hover:text-primary-600 transition-colors"
        @click="goHome"
      >
        Quiz Platform
      </button>
      <div class="flex-1 flex flex-col px-3 pb-4">
        <nav class="flex flex-col gap-1">
          <AppMenuItem
            v-for="item in menuItems"
            :key="item.label"
            :label="item.label"
            :icon="item.icon"
            @click="handleMenuClick"
          />
        </nav>
        <div class="mt-auto pt-4">
          <AppMenuItem label="Выйти" icon="pi pi-sign-out" @click="handleMenuClick" />
        </div>
      </div>
    </aside>

    <Drawer v-if="!isDesktop" v-model:visible="isMobileMenuOpen" class="w-64" :closable="false">
      <template #header>
        <div class="text-xl font-bold text-primary cursor-pointer select-none" @click="goHome">
          Quiz Platform
        </div>
      </template>

      <div class="flex flex-col h-full">
        <div class="mt-6 flex flex-col gap-1 flex-1">
          <AppMenuItem
            v-for="item in menuItems"
            :key="item.label"
            :label="item.label"
            :icon="item.icon"
            @click="handleMenuClick"
          />
        </div>

        <div class="mt-auto pt-4">
          <AppMenuItem label="Выйти" icon="pi pi-sign-out" @click="handleMenuClick" />
        </div>
      </div>
    </Drawer>

    <div class="flex-1 flex flex-col">
      <header
        class="h-14 flex justify-between items-center px-4 border-b border-surface-200 dark:border-surface-800 bg-surface-0 dark:bg-surface-900"
      >
        <Button icon="pi pi-bars" class="inline-flex" text @click="toggleSidebar" />
        <div class="hidden lg:block font-semibold text-surface-600 dark:text-surface-400">
          Панель управления
        </div>
        <div class="flex items-center gap-4">
          <ThemeSwitcher />
          <span class="font-medium hidden sm:inline">{{ userEmail }}</span>
        </div>
      </header>
      <main class="flex-1 p-6 overflow-y-auto">
        <slot />
      </main>
    </div>
  </div>
</template>
