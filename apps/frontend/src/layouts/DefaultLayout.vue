<script setup lang="ts">
import PublicTopbar from '@/components/layout/PublicTopbar.vue'
</script>

<template>
  <div class="layout-shell">
    <PublicTopbar />

    <main class="layout-main flex items-center justify-center px-4 py-6 md:px-6 md:py-10">
      <slot />
    </main>
  </div>
</template>
