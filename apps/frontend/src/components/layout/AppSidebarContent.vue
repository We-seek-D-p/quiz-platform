<script setup lang="ts">
import AppLogo from '@/components/layout/AppLogo.vue'
import AppMenuItem from '@/components/layout/AppMenuItem.vue'

withDefaults(
  defineProps<{
    menuItems: Array<{ label: string; icon: string }>
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
</script>

<template>
  <div class="flex h-full flex-col">
    <div v-if="showLogo" class="px-5 py-4">
      <AppLogo />
    </div>

    <div class="flex flex-1 flex-col px-3 pb-6">
      <nav class="mt-2 flex flex-1 flex-col gap-1">
        <AppMenuItem
          v-for="item in menuItems"
          :key="item.label"
          :label="item.label"
          :icon="item.icon"
          @click="emit('itemClick')"
        />
      </nav>

      <div class="mt-auto pt-4">
        <AppMenuItem label="Выйти" icon="pi pi-sign-out" @click="emit('logoutClick')" />
      </div>
    </div>
  </div>
</template>
