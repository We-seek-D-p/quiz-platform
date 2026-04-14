<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Toast from 'primevue/toast'
import DefaultLayout from './layouts/DefaultLayout.vue'
import DashboardLayout from './layouts/DashboardLayout.vue'
import GameLayout from './layouts/GameLayout.vue'

const route = useRoute()

const layout = computed(() => {
  switch (route.meta.layout) {
    case 'DashboardLayout':
      return DashboardLayout
    case 'GameLayout':
      return GameLayout
    default:
      return DefaultLayout
  }
})
</script>

<template>
  <Toast group="global" position="top-right" />
  <component :is="layout">
    <router-view />
  </component>
</template>
