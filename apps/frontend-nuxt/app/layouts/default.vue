<script setup lang="ts">
import PublicTopbar from '~/components/navigation/PublicTopbar.vue'
</script>

<template>
  <div class="layout-shell">
    <PublicTopbar />

    <main class="layout-main p-5 md:p-8">
      <slot />
    </main>
  </div>
</template>
