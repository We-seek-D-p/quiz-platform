<script setup lang="ts">
import { useRoute } from 'vue-router'
import AppLogo from '@/components/layout/AppLogo.vue'
import AppMenuItem from '@/components/layout/AppMenuItem.vue'
import type { DashboardNavigationItem } from '@/layouts/dashboardNavigation'

const props = withDefaults(
  defineProps<{
    menuItems: DashboardNavigationItem[]
    showLogo?: boolean
  }>(),
  {
    showLogo: false,
  },
)

const emit = defineEmits<{
  itemClick: []
  logoutClick: []
}>()

const route = useRoute()

const isItemActive = (item: DashboardNavigationItem): boolean => {
  if (!item.routeName) {
    return false
  }

  return route.matched.some((record) => record.name === item.routeName)
}
</script>

<template>
  <div class="flex h-full flex-col">
    <div v-if="showLogo" class="px-5 py-4">
      <AppLogo />
    </div>

    <div class="flex flex-1 flex-col px-3 pb-6">
      <nav class="mt-2 flex flex-1 flex-col gap-1">
        <AppMenuItem
          v-for="item in props.menuItems"
          :key="item.key"
          :label="item.label"
          :icon="item.icon"
          :to="item.routeName ? { name: item.routeName } : undefined"
          :active="isItemActive(item)"
          :disabled="item.disabled"
          @select="emit('itemClick')"
        />
      </nav>

      <div class="mt-auto pt-4">
        <AppMenuItem label="Выйти" icon="pi pi-sign-out" @select="emit('logoutClick')" />
      </div>
    </div>
  </div>
</template>
