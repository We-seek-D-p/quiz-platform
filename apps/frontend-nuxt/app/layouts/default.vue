<script setup lang="ts">
import PublicTopbar from '~/components/navigation/PublicTopbar.vue'
</script>

<template>
  <div class="layout-shell">
    <PublicTopbar />

    <main class="layout-main layout-main--centered p-5 md:p-8">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.layout-main--centered {
  min-height: calc(100dvh - var(--app-topbar-height));
  display: grid;
  place-items: center;
}
</style>
