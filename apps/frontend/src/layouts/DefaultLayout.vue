<script setup lang="ts">
import { useRouter } from 'vue-router'
import AppTopbar from '@/components/layout/AppTopbar.vue'

const router = useRouter()
</script>

<template>
  <div class="layout-shell">
    <AppTopbar show-brand show-profile-button @profile-click="router.push('/host')" />

    <main class="layout-main flex items-center justify-center px-4 py-6 md:px-6 md:py-10">
      <slot />
    </main>
  </div>
</template>
